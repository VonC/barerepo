package commits

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sync"
	"sync/atomic"

	"github.com/VonC/barerepo/internal/filequeue"
	"github.com/VonC/barerepo/internal/print"
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
	processFunc func(*Commit) error
}

func NewQueue(basedir string, fs fs.FS, process func(*Commit) error) (*queue, error) {
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
		print.Printf(fmt.Sprintf("Init fileonly true: files detected"))
		q.state.setFileOnly(true)
	}
	return q, nil
}

// Add a commit to the queue, to be processed (or saved to disk if program stops too soon)
func (q *queue) Add(c *Commit) error {
	q.state.RLock()
	defer q.state.RUnlock()
	print.Printf(fmt.Sprintf("ADD: Add commit %s", c))
	if q.state.isFileOnly() {
		print.Printf(fmt.Sprintf("ADD: fileonly"))
		return q.save(c)
	}
	select {
	case q.commitChan <- c:
		print.Printf(fmt.Sprintf("ADD: Commit sent to queue '%s'", c))
		return nil
	default:
		q.state.setFileOnly(true)
		print.Printf(fmt.Sprintf("ADD: set fileony, save '%s'", c))
		return q.save(c)
	}
}

type state struct {
	// https://stackoverflow.com/questions/52863273/how-to-make-a-variable-thread-safe
	sync.RWMutex
	fileOnly bool
	atomic.Value
}

// https://stackoverflow.com/questions/52863273/how-to-make-a-variable-thread-safe
// https://stackoverflow.com/questions/39123453/concurrently-how-to-manage-values-states-and-avoiding-a-race-condition
func (s *state) isFileOnly() bool {
	return s.fileOnly
}

func (s *state) setFileOnly(fo bool) {
	s.Store(fo)
}

func (q *queue) save(c *Commit) error {
	b, err := json.Marshal(c)
	if err == nil {
		err = q.fq.Push(b)
	}
	print.Printf(fmt.Sprintf("save: b: '%s', err '%+v'", string(b), err))
	return err
}

// Run starts the queue, waiting for new commits or processing those saved on disk.
func (q *queue) Run() {
	// https://www.opsdash.com/blog/job-queues-in-go.html
	go func() {
		var c *Commit
		var dropFunc filequeue.DropFunc
		for {
			q.state.RLock()
			dropFunc = nil
			select {
			case <-q.cancelChan:
				// TODO save remaining job from channel to file, after loading existing files
				print.Printf(fmt.Sprintf("Number commits left in channel: %d", len(q.commitChan)))
				q.state.RUnlock()
				return
			case c = <-q.commitChan:
				print.Printf(fmt.Sprintf("RUN: Commit received '%s'", c))
			default:
				if c == nil {
					c, dropFunc = q.load()
					if c == nil {
						if q.state.isFileOnly() {
							print.Printf(fmt.Sprintf("Reset fileOnly to false"))
							q.state.setFileOnly(false)
						}
					} else {
						print.Printf(fmt.Sprintf("load from filequeue Commit '%s', dropFunc '%+v'", c, dropFunc))
					}
				}
			}
			q.state.RUnlock()
			if err := q.process(c, dropFunc); err != nil {
				print.Printf(fmt.Sprintf("Unable to process commit '%s': error '%+v'", c, err))
			} else if c != nil {
				print.Printf(fmt.Sprintf("Processed Commit '%s', dropFunc '%+v'", c, dropFunc))
			}
			c = nil
		}
	}()
}

// Stop send a struct to queue cancel channel
func (q *queue) Stop() {
	q.cancelChan <- struct{}{}
}

func (q *queue) process(c *Commit, dropFunc filequeue.DropFunc) error {
	if c == nil {
		return nil
	}
	print.Printf(fmt.Sprintf("Processing %s, dropfunc %v", c, dropFunc))
	if q.processFunc != nil {
		if err := q.processFunc(c); err != nil {
			return err
		}
		if dropFunc != nil {
			print.Printf(fmt.Sprintf("Calling DropFunc on %s", c))
			return dropFunc()
		}
	} else {
		print.Printf(fmt.Sprintf("nil processFunc: NO Processing %s", c))
	}
	return nil
}

func (q *queue) load() (*Commit, filequeue.DropFunc) {
	b, dropFunc, err := q.fq.Pop()
	res := &Commit{}
	if err == nil && b != nil {
		err = json.Unmarshal(b, res)
	} else {
		res = nil
	}
	if b == nil && err == nil {
		return nil, nil
	}
	print.Printf(fmt.Sprintf("load: Commit loaded '%s', err='%+v'", res, err))
	if err != nil {
		return nil, nil
	}
	q.state.setFileOnly(true)
	return res, dropFunc
}
