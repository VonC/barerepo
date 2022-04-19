package commits_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/VonC/barerepo/internal/commits"
	"github.com/hack-pad/hackpadfs/os"
)

func TestQueue(t *testing.T) {
	ofs := os.NewFS()
	var err error
	/*
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			t.Errorf("unable to get the current filename")
		}
		dirname := filepath.Dir(filename)
		var ffs hackpadfs.FS
		dirname = strings.ReplaceAll(dirname, "c:\\", "c/")
		dirname = strings.ReplaceAll(dirname, "\\", "/")
		ffs, err = ofs.Sub(dirname)
		if err != nil {
			t.Fatalf("Error on fs.Sub: %+v", err)
		}
		ofs, ok = ffs.(*os.FS)
		if !ok {
			t.Fatalf("unable to get the current FS")
		}
	*/
	commits.Printf(fmt.Sprintf("Start test"))
	var q commits.Queue
	if q, err = commits.NewQueue("test", ofs, process); err != nil {
		t.Fatalf("Error on NewQueue: %+v", err)
	}
	q.Run()
	c1 := commits.NewCommit("c1", time.Now())
	c2 := commits.NewCommit("c2", time.Now())
	c3 := commits.NewCommit("c3", time.Now())
	q.Add(c1)
	q.Add(c2)
	q.Add(c3)
	time.Sleep(3 * time.Second)
	q.Stop()
	time.Sleep(1 * time.Second)
	commits.Printf(fmt.Sprintf("test done"))
}

func process(c *commits.Commit) {
	commits.Printf(fmt.Sprintf("Process commit %s\n", c.String()))
	time.Sleep(2 * time.Second)
}
