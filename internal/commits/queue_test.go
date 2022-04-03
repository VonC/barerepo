package commits_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/VonC/barerepo/internal/commits"
	"github.com/hack-pad/hackpadfs/os"
)

func TestQueue(t *testing.T) {
	fs := os.NewFS()
	var q commits.Queue
	var err error
	if q, err = commits.NewQueue("test", fs, process); err != nil {
		t.Errorf("Error on NewQueue: %+v", err)
	}
	q.Run()
	c1 := commits.NewCommit("c1", time.Now())
	c2 := commits.NewCommit("c2", time.Now())
	c3 := commits.NewCommit("c3", time.Now())
	q.Add(c1)
	q.Add(c2)
	q.Add(c3)
	time.Sleep(1 * time.Second)
	q.Stop()
	time.Sleep(1 * time.Second)
	fmt.Println("test done")
}

func process(c *commits.Commit) {
	fmt.Printf("Process commit %s\n", c.String())
	time.Sleep(2 * time.Second)
}
