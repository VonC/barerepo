package commits_test

import (
	"testing"
	"time"

	"github.com/VonC/barerepo/internal/commits"
	"github.com/VonC/barerepo/internal/print"
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
	print.Printf("Start test")
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
	//t.SkipNow()
	time.Sleep(25 * time.Second)
	q.Stop()
	time.Sleep(1 * time.Second)
	print.Printf("test done")
}

func process(c *commits.Commit) error {
	print.Printf("Process TEST commit %s", c.String())
	time.Sleep(2 * time.Second)
	return nil
}
