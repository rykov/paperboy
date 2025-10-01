package send

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wneessen/go-mail"
)

// memoryQueue is the built-in channel-based queue implementation
type memoryQueue struct {
	tasks   chan *mail.Msg
	waiter  *sync.WaitGroup
	context context.Context
	sender  Sender

	stop  bool
	stopL sync.Mutex

	// ID for logging
	ID string

	// Rate limiting
	throttle time.Duration
	workers  int
}

// Delivery configuration
type Config struct {
	QueueID  string
	SendRate float32
	Workers  int
}

// NewInMemory creates a new in-memory channel-based queue
func NewInMemory(ctx context.Context, cfg *Config, sender Sender) (Manager, error) {
	// Rate configuration
	throttle := time.Duration(0)
	if cfg.SendRate > 0 {
		throttle = time.Duration(1000 / cfg.SendRate)
		throttle = throttle * time.Millisecond
	}

	workers := cfg.Workers
	if workers == 0 {
		workers = 1
	}

	// Display configuration
	fmt.Printf("Sending an email every %s via %d workers\n", throttle, workers)

	queue := &memoryQueue{
		ID:       cfg.QueueID,
		tasks:    make(chan *mail.Msg, 10),
		waiter:   &sync.WaitGroup{},
		context:  ctx,
		sender:   sender,
		throttle: throttle,
		workers:  workers,
	}

	// Capture context cancellation for graceful exit
	done := ctx.Done()
	go func() {
		<-done
		queue.shutdown()
	}()

	// Start delivery workers
	for i := 0; i < workers; i++ {
		if err := queue.startWorker(i); err != nil {
			queue.shutdown()
			return nil, err
		}

		// HACK: Avoid race warning
		queue.stopL.Lock()
		stopped := queue.stop
		queue.stopL.Unlock()
		if stopped {
			break
		}
	}

	return queue, nil
}

// Enqueue adds an email message to the queue
func (d *memoryQueue) Enqueue(ctx context.Context, msg *mail.Msg) error {
	// Check if queue is stopped
	d.stopL.Lock()
	stopped := d.stop
	d.stopL.Unlock()

	if stopped {
		return fmt.Errorf("queue is closed")
	}

	// Apply rate limiting
	if d.throttle > 0 {
		time.Sleep(d.throttle)
	}

	// Send to channel (non-blocking check for closure)
	select {
	case d.tasks <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Channel might be closed
		if stopped {
			return fmt.Errorf("queue is closed")
		}
		// Try again with blocking send
		d.tasks <- msg
		return nil
	}
}

// Wait blocks until all queued tasks are completed
func (d *memoryQueue) Wait() error {
	d.waiter.Wait()
	return nil
}

// Close gracefully shuts down the queue
func (d *memoryQueue) Close() error {
	d.shutdown()
	return nil
}

// shutdown signals workers to stop (internal method)
func (d *memoryQueue) shutdown() {
	d.stopL.Lock()
	defer d.stopL.Unlock()

	if !d.stop {
		d.stop = true
		close(d.tasks)
	}
}

// startWorker spawns a worker goroutine to process tasks
func (d *memoryQueue) startWorker(id int) error {
	fmt.Printf("[%d] Starting worker...\n", id)
	d.waiter.Add(1)

	// Dial up the sender
	conn, err := d.sender.NewConn()
	if err != nil {
		return err
	}

	go func() {
		defer d.waiter.Done()
		defer fmt.Printf("[%d] Stopping worker...\n", id)
		defer conn.Close()

		for {
			select {
			case <-d.context.Done():
				fmt.Printf("[%d] Worker stopped on cancellation\n", id)
				return
			case msg, more := <-d.tasks:
				if !more {
					return
				}

				// Log the sending
				toList := msg.GetToString()
				fmt.Printf("[%d] Sending %s to %s\n", id, d.ID, toList)

				// Send the message
				if err := conn.Send(msg); err != nil {
					fmt.Printf("[%d] Could not send email: %s\n", id, err)
					conn.Close() // Replace errored connection
					conn, err = d.sender.NewConn()
					if err != nil {
						fmt.Printf("[%d] Failed to recreate sender after error: %v\n", id, err)
						return
					}
				}
			}
		}
	}()

	return nil
}
