package commits

import (
	"sync"
	"time"

	"github.com/VonC/barerepo/internal/filequeue"
	"github.com/hack-pad/hackpadfs"
)

type job struct {
	message   string
	timestamp time.Time
}

type Queue interface {
	Add(message string, t time.Time) error
	Run()
}

type queue struct {
	jobChan    chan *job
	cancelChan chan struct{}
	state      *state
	q          filequeue.Queue
}

func NewQueue(basedir string, fs hackpadfs.FS) (*queue, error) {
	q, err := filequeue.New(basedir, fs)
	if err != nil {
		return nil, err
	}
	return &queue{
		jobChan:    make(chan *job, 1),
		cancelChan: make(chan struct{}),
		state:      &state{},
		q:          q,
	}, nil
}

// Add a commit to the queue, to be processed (or saved to disk if program stops too soon)
func (q *queue) Add(message string, t time.Time) error {
	q.state.Lock()
	defer q.state.RUnlock()
	j := &job{
		message:   message,
		timestamp: t,
	}
	if q.state.fileOnly {
		return q.save(j)
	}
	select {
	case q.jobChan <- j:
		return nil
	default:
		q.state.fileOnly = true
		return q.save(j)
	}
}

type state struct {
	// https://stackoverflow.com/questions/52863273/how-to-make-a-variable-thread-safe
	sync.RWMutex
	fileOnly bool
}

func (q *queue) save(j *job) error {
	return nil
}

// Run starts the queue, waiting for new commits or processing those saved on disk.
func (q *queue) Run() {
	// https://www.opsdash.com/blog/job-queues-in-go.html
	go func() {
		var j *job
		for {
			q.state.Lock()
			defer q.state.RUnlock()
			select {
			case <-q.cancelChan:
				// TODO save remaining job from channel to file, after loading existing files
				return
			case j = <-q.jobChan:
			default:
				j = q.load()
			}
			q.state.RUnlock()
			q.process(j)
		}
	}()
}

func (q *queue) process(j *job) {
	if j == nil {
		return
	}
	return
}

func (q *queue) load() *job {
	return nil
}
