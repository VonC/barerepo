// The filequeue package defines a Queue interface and a default
// implementation that uses files.
// From https://github.com/rstudio/filequeue/blob/main/filequeue.go
package filequeue

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hack-pad/hackpadfs"
)

// Queue implements a FIFO Queue backed with files so that multiple
// processes may consume items as long as they have access to the
// same filesystem (which may be NFS-mounted).
type Queue interface {
	Len() (int, error)
	Pop() ([]byte, error)
	Push([]byte) error
}

func New(baseDir string, fs hackpadfs.FS) (Queue, error) {
	baseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, err
	}

	baseDir = strings.ReplaceAll(baseDir, "c:\\", "")

	err = hackpadfs.MkdirAll(fs, baseDir, hackpadfs.ModeDir)

	//fs, err = hackpadfs.Sub(fs, baseDir)

	fq := &FileQueue{
		baseDir: baseDir,
		fs:      fs,
	}

	return fq, err
}

// FileQueue implements the Queue interface via files and
// filesystem operations.
type FileQueue struct {
	baseDir string
	fs      hackpadfs.FS
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

// Pop returns the least-recently added item, if available.
//
// In the case of an empty queue, the return value will be nil and
// there will not be an error. If an item is popped, presumably by
// another consumer, before it may be read, then the next available
// item known at the time the item list was built will be tried.
func (fq *FileQueue) Pop() ([]byte, error) {
	items, err := fq.listItemsSorted()
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, nil
	}

	for _, loopItem := range items {
		item := loopItem

		fullPath := filepath.Join(fq.baseDir, item)
		tmpPath := fmt.Sprintf("%s.pop-%v", fullPath, rand.Float64())

		if err := hackpadfs.Rename(fq.fs, fullPath, tmpPath); err != nil {
			continue
		}

		itemBytes, err := hackpadfs.ReadFile(fq.fs, tmpPath)
		if err != nil {
			return nil, err
		}

		if err := hackpadfs.Remove(fq.fs, tmpPath); err != nil {
			return nil, err
		}

		return itemBytes, nil
	}

	return nil, nil
}

// Push writes the item bytes to a timestamped file, returning any
// error from os.WriteFile.
func (fq *FileQueue) Push(b []byte) error {
	fullPath := filepath.Join(
		fq.baseDir,
		fmt.Sprintf("%v.item", time.Now().UnixNano()),
	)

	var f hackpadfs.File
	var err error
	if f, err = hackpadfs.Create(fq.fs, fullPath); err != nil {
		return err
	}

	_, err = hackpadfs.WriteFile(f, b)
	return err
}

func (fq *FileQueue) listItemsSorted() ([]string, error) {
	dirEnts, err := hackpadfs.ReadDir(fq.fs, fq.baseDir)
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
