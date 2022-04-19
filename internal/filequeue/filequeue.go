// The filequeue package defines a Queue interface and a default
// implementation that uses files.
// From https://github.com/rstudio/filequeue/blob/main/filequeue.go
package filequeue

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/VonC/barerepo/internal/print"
)

// Queue implements a FIFO Queue backed with files so that multiple
// processes may consume items as long as they have access to the
// same filesystem (which may be NFS-mounted).
type Queue interface {
	Len() (int, error)
	Pop() ([]byte, DropFunc, error)
	Push([]byte) error
}

func New(baseDir string, ofs fs.FS) (Queue, error) {
	baseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, err
	}

	baseDir = strings.ReplaceAll(baseDir, "c:\\", "")

	err = os.MkdirAll(baseDir, fs.ModeDir)

	//fs, err = os.Sub(fs, baseDir)

	fq := &FileQueue{
		baseDir: baseDir,
		ofs:     ofs,
	}
	print.Printf(fmt.Sprintf("New file queue: '%s'", baseDir))
	return fq, err
}

// FileQueue implements the Queue interface via files and
// filesystem operations.
type FileQueue struct {
	baseDir string
	ofs     fs.FS
}

// Len returns the number of items known at this moment in time.
//
// In the case of an unreadable directory or any other error, the
// error will be returned along with length -1.
func (fq *FileQueue) Len() (int, error) {
	items, err := fq.listItemsSorted()
	if err != nil {
		return -1, err
	}

	return len(items), nil
}

type DropFunc func() error

// Pop returns the least-recently added item, if available.
//
// In the case of an empty queue, the return value will be nil and
// there will not be an error. If an item is popped, presumably by
// another consumer, before it may be read, then the next available
// item known at the time the item list was built will be tried.
func (fq *FileQueue) Pop() ([]byte, DropFunc, error) {
	items, err := fq.listItemsSorted()
	if err != nil {
		return nil, nil, err
	}

	print.Printf(fmt.Sprintf("Load queue len: '%d'", len(items)))
	if len(items) == 0 {
		return nil, nil, nil
	}

	item := items[0]

	fullPath := filepath.Join(fq.baseDir, item)

	itemBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, nil, err
	}

	return itemBytes, func() error {
			err := os.Remove(fullPath)
			return err
		},
		nil

}

// Push writes the item bytes to a timestamped file, returning any
// error from os.WriteFile.
func (fq *FileQueue) Push(b []byte) error {
	fullPath := filepath.Join(
		fq.baseDir,
		fmt.Sprintf("%v.item", time.Now().UnixNano()),
	)

	var err error
	if _, err = os.Create(fullPath); err != nil {
		return err
	}

	err = os.WriteFile(fullPath, b, 0755)
	return err
}

func (fq *FileQueue) listItemsSorted() ([]string, error) {
	dirEnts, err := os.ReadDir(fq.baseDir)
	if err != nil {
		return nil, err
	}

	items := []string{}

	for _, loopDirEnt := range dirEnts {
		basename := loopDirEnt.Name()
		if strings.HasSuffix(basename, ".item") {
			items = append(items, basename)
		}
	}

	sort.Strings(items)

	return items, nil
}
