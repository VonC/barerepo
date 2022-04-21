package commits

import (
	"fmt"
	"os"
	"time"

	"github.com/VonC/barerepo/internal/print"
)

type Commit struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	timestamp time.Time
}

func NewCommit(message string, timestamp time.Time) *Commit {
	return &Commit{
		Message:   message,
		timestamp: timestamp,
		Timestamp: timestamp.Format(time.RFC3339),
	}
}

func (c *Commit) time() time.Time {
	var err error
	if c.timestamp.IsZero() {
		c.timestamp, err = time.Parse(time.RFC3339, c.Timestamp)
	}
	if err != nil {
		print.Printf(fmt.Sprintf("Error on commit Timestamp '%s' parsing '%+v'", c.Timestamp, err))
		c.timestamp = time.Time{}
	}
	return c.timestamp
}

func (c *Commit) String() string {
	if c == nil {
		return "<nil>"
	}
	return fmt.Sprintf("message '%s' (%s from %s)", c.Message, c.time().Format(time.RFC1123), c.Timestamp)
}
