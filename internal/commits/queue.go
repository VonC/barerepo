package commits

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"sync"

	"github.com/VonC/barerepo/internal/filequeue"
)

type Queue interface {
	Add(c *Commit) error
	Run()
	Stop()
}

type queue struct {
	commitChan  chan *Commit
	cancelChan  chan struct{}
	state       *state
	fq          filequeue.Queue
	processFunc func(*Commit)
}

func NewQueue(basedir string, fs fs.FS, process func(*Commit)) (*queue, error) {
	fq, err := filequeue.New(basedir, fs)
	if err != nil {
		return nil, err
	}
	q := &queue{
		commitChan: make(chan *Commit, 1),
		cancelChan: make(chan struct{}),
		state: &state{
			fileOnly: false,
		},
		fq:          fq,
		processFunc: process,
	}
	l, err := q.fq.Len()
	if err == nil && l > 0 {
		Printf(fmt.Sprintf("Init fileonly true: files detected"))
		q.state.fileOnly = true
	}
	return q, nil
}

// Add a commit to the queue, to be processed (or saved to disk if program stops too soon)
func (q *queue) Add(c *Commit) error {
	q.state.RLock()
	defer q.state.RUnlock()
	Printf(fmt.Sprintf("ADD: Add commit %s\n", c))
	if q.state.fileOnly {
		Printf(fmt.Sprintf("ADD: fileonly\n"))
		return q.save(c)
	}
	select {
	case q.commitChan <- c:
		Printf(fmt.Sprintf("ADD: Commit sent to queue '%s'\n", c))
		return nil
	default:
		q.state.fileOnly = true
		Printf(fmt.Sprintf("ADD: set fileony, save '%s'\n", c))
		return q.save(c)
	}
}

type state struct {
	// https://stackoverflow.com/questions/52863273/how-to-make-a-variable-thread-safe
	sync.RWMutex
	fileOnly bool
}

func (q *queue) save(c *Commit) error {
	b, err := json.Marshal(c)
	if err == nil {
		err = q.fq.Push(b)
	}
	Printf(fmt.Sprintf("save: b: '%s', err '%+v'\n", string(b), err))
	return err
}

// Run starts the queue, waiting for new commits or processing those saved on disk.
func (q *queue) Run() {
	// https://www.opsdash.com/blog/job-queues-in-go.html
	go func() {
		var c *Commit
		for {
			q.state.RLock()
			select {
			case <-q.cancelChan:
				// TODO save remaining job from channel to file, after loading existing files
				Printf(fmt.Sprintf("Number commits left in channel: %d\n", len(q.commitChan)))
				q.state.RUnlock()
				return
			case c = <-q.commitChan:
				Printf(fmt.Sprintf("RUN: Commit received '%s'\n", c))
			default:
				if c == nil {
					c = q.load()
					if c == nil {
						if q.state.fileOnly {
							Printf(fmt.Sprintf("Reset fileOnly to false"))
							q.state.fileOnly = false
						}
					}
				}
			}
			q.state.RUnlock()
			q.process(c)
		}
	}()
}

// Stop send a struct to queue cancel channel
func (q *queue) Stop() {
	q.cancelChan <- struct{}{}
}

func (q *queue) process(c *Commit) {
	if c == nil {
		return
	}
	Printf(fmt.Sprintf("Processing %s\n", c))
	if q.processFunc != nil {
		q.processFunc(c)
	}
	return
}

func (q *queue) load() *Commit {
	b, err := q.fq.Pop()
	res := &Commit{}
	if err == nil && b != nil {
		err = json.Unmarshal(b, res)
	}
	if b == nil && err == nil {
		return nil
	}
	Printf(fmt.Sprintf("load: Commit loaded '%s', err='%+v'\n", res, err))
	if err != nil {
		return nil
	}
	q.state.fileOnly = true
	return res
}

func Printf(msg string) {
	// https://github.com/golang/go/issues/36619
	bufStdout := bufio.NewWriter(os.Stdout)
	_, err := bufStdout.WriteString(msg)
	defer bufStdout.Flush()
	if err == nil {
		fmt.Fprintf(bufStdout, "\n")
	} else {
		fmt.Printf("Error on Printf '%s': %+v\n", msg, err)
	}
}
