package send

import (
	"context"

	"github.com/wneessen/go-mail"
)

// Manager is the interface for managing email task queues and workers.
// Implementations can use different backends (channels, taskq, machinery, etc.)
type Manager interface {
	// Enqueue adds an email message to the queue
	// Returns an error if the task cannot be queued
	Enqueue(ctx context.Context, msg *mail.Msg) error

	// Wait blocks until all queued tasks are completed
	// Returns an error if any tasks failed
	Wait() error

	// Close gracefully shuts down the queue manager
	// It should stop accepting new tasks and wait for pending tasks to complete
	Close() error
}

// Sender interface for creating per-worker connections
// Implemented by mail.smtpSender and mail.testSender
type Sender interface {
	NewConn() (Conn, error)
}

// Conn interface for sending emails
type Conn interface {
	Send(msg ...*mail.Msg) error
	Close() error
}
