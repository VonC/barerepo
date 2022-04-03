package commits_test

import (
	"fmt"
	"testing"

	"github.com/VonC/barerepo/internal/commits"
	"github.com/hack-pad/hackpadfs/os"
)

func TestQueue(t *testing.T) {
	fs := os.NewFS()
	var q commits.Queue
	var err error
	if q, err = commits.NewQueue("test", fs); err != nil {
		t.Errorf("Error on NewQueue: %+v", err)
	}
	fmt.Println(q)
}
