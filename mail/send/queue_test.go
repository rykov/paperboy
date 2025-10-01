package send

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wneessen/go-mail"
)

// mockSender implements Sender interface for testing
type mockSender struct {
	connFunc func() (Conn, error)
}

func (m *mockSender) NewConn() (Conn, error) {
	if m.connFunc != nil {
		return m.connFunc()
	}
	return &mockConn{}, nil
}

// mockConn implements Conn interface for testing
type mockConn struct {
	sendFunc  func(msg ...*mail.Msg) error
	closeFunc func() error
	mu        sync.Mutex
	sent      []*mail.Msg
	closed    bool
}

func (m *mockConn) Send(msg ...*mail.Msg) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sendFunc != nil {
		return m.sendFunc(msg...)
	}

	m.sent = append(m.sent, msg...)
	return nil
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockConn) getSent() []*mail.Msg {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]*mail.Msg{}, m.sent...)
}

func TestNewInMemory_Success(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		QueueID:  "test-campaign",
		SendRate: 0,
		Workers:  2,
	}
	sender := &mockSender{}

	queue, err := NewInMemory(ctx, cfg, sender)
	if err != nil {
		t.Fatalf("NewInMemory() failed: %v", err)
	}
	if queue == nil {
		t.Fatal("NewInMemory() returned nil queue")
	}

	// Clean up
	queue.Close()
	queue.Wait()
}

func TestNewInMemory_DefaultWorkers(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		QueueID:  "test-campaign",
		SendRate: 0,
		Workers:  0, // Should default to 1
	}
	sender := &mockSender{}

	queue, err := NewInMemory(ctx, cfg, sender)
	if err != nil {
		t.Fatalf("NewInMemory() failed: %v", err)
	}

	// Verify queue was created successfully
	if queue == nil {
		t.Fatal("NewInMemory() returned nil queue")
	}

	// Clean up
	queue.Close()
	queue.Wait()
}

func TestNewInMemory_SenderConnectionError(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		QueueID:  "test-campaign",
		SendRate: 0,
		Workers:  1,
	}

	expectedErr := errors.New("connection failed")
	sender := &mockSender{
		connFunc: func() (Conn, error) {
			return nil, expectedErr
		},
	}

	queue, err := NewInMemory(ctx, cfg, sender)
	if err == nil {
		t.Fatal("NewInMemory() should have failed with connection error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if queue != nil {
		t.Error("NewInMemory() should return nil queue on error")
	}
}

func TestMemoryQueue_EnqueueAndSend(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		QueueID:  "test-campaign",
		SendRate: 0,
		Workers:  1,
	}

	conn := &mockConn{}
	sender := &mockSender{
		connFunc: func() (Conn, error) {
			return conn, nil
		},
	}

	queue, err := NewInMemory(ctx, cfg, sender)
	if err != nil {
		t.Fatalf("NewInMemory() failed: %v", err)
	}

	// Create and enqueue test messages
	numMessages := 5
	for i := 0; i < numMessages; i++ {
		msg := mail.NewMsg()
		if err := msg.To("test@example.com"); err != nil {
			t.Fatalf("Failed to set To: %v", err)
		}
		if err := queue.Enqueue(ctx, msg); err != nil {
			t.Fatalf("Enqueue() failed: %v", err)
		}
	}

	// Close and wait
	queue.Close()
	if err := queue.Wait(); err != nil {
		t.Fatalf("Wait() failed: %v", err)
	}

	// Verify all messages were sent
	sent := conn.getSent()
	if len(sent) != numMessages {
		t.Errorf("Expected %d messages sent, got %d", numMessages, len(sent))
	}
}

func TestMemoryQueue_RateLimiting(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		QueueID:  "test-campaign",
		SendRate: 10, // 10 messages per second = 100ms between messages
		Workers:  1,
	}

	conn := &mockConn{}
	sender := &mockSender{
		connFunc: func() (Conn, error) {
			return conn, nil
		},
	}

	queue, err := NewInMemory(ctx, cfg, sender)
	if err != nil {
		t.Fatalf("NewInMemory() failed: %v", err)
	}

	// Enqueue messages and track timing
	numMessages := 3
	start := time.Now()

	for i := 0; i < numMessages; i++ {
		msg := mail.NewMsg()
		if err := msg.To("test@example.com"); err != nil {
			t.Fatalf("Failed to set To: %v", err)
		}
		if err := queue.Enqueue(ctx, msg); err != nil {
			t.Fatalf("Enqueue() failed: %v", err)
		}
	}

	elapsed := time.Since(start)

	// Close and wait
	queue.Close()
	queue.Wait()

	// With rate limiting, 3 messages should take at least 200ms (100ms * 2 intervals)
	expectedMinDuration := 200 * time.Millisecond
	if elapsed < expectedMinDuration {
		t.Errorf("Rate limiting not working: expected at least %v, got %v", expectedMinDuration, elapsed)
	}
}

func TestMemoryQueue_MultipleWorkers(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		QueueID:  "test-campaign",
		SendRate: 0,
		Workers:  3,
	}

	var sendCount atomic.Int32
	var mu sync.Mutex
	conns := make([]*mockConn, 0)

	sender := &mockSender{
		connFunc: func() (Conn, error) {
			conn := &mockConn{
				sendFunc: func(msg ...*mail.Msg) error {
					sendCount.Add(1)
					time.Sleep(10 * time.Millisecond) // Simulate send time
					return nil
				},
			}
			mu.Lock()
			conns = append(conns, conn)
			mu.Unlock()
			return conn, nil
		},
	}

	queue, err := NewInMemory(ctx, cfg, sender)
	if err != nil {
		t.Fatalf("NewInMemory() failed: %v", err)
	}

	// Enqueue multiple messages
	numMessages := 10
	for i := 0; i < numMessages; i++ {
		msg := mail.NewMsg()
		if err := msg.To("test@example.com"); err != nil {
			t.Fatalf("Failed to set To: %v", err)
		}
		if err := queue.Enqueue(ctx, msg); err != nil {
			t.Fatalf("Enqueue() failed: %v", err)
		}
	}

	queue.Close()
	queue.Wait()

	// Verify all messages were sent
	if sendCount.Load() != int32(numMessages) {
		t.Errorf("Expected %d messages sent, got %d", numMessages, sendCount.Load())
	}

	// Verify multiple workers were used (at least 3 connections created)
	mu.Lock()
	numConns := len(conns)
	mu.Unlock()

	if numConns != 3 {
		t.Errorf("Expected 3 worker connections, got %d", numConns)
	}
}

func TestMemoryQueue_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := &Config{
		QueueID:  "test-campaign",
		SendRate: 0,
		Workers:  1,
	}

	var sendCount atomic.Int32
	conn := &mockConn{
		sendFunc: func(msg ...*mail.Msg) error {
			sendCount.Add(1)
			time.Sleep(50 * time.Millisecond) // Slow send
			return nil
		},
	}

	sender := &mockSender{
		connFunc: func() (Conn, error) {
			return conn, nil
		},
	}

	queue, err := NewInMemory(ctx, cfg, sender)
	if err != nil {
		t.Fatalf("NewInMemory() failed: %v", err)
	}

	// Enqueue one message
	msg := mail.NewMsg()
	if err := msg.To("test@example.com"); err != nil {
		t.Fatalf("Failed to set To: %v", err)
	}
	if err := queue.Enqueue(ctx, msg); err != nil {
		t.Fatalf("Enqueue() failed: %v", err)
	}

	// Cancel context while message is being sent
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait should complete
	queue.Wait()

	// At least one message should have been sent
	if sendCount.Load() < 1 {
		t.Error("Expected at least one message to be sent before cancellation")
	}
}

func TestMemoryQueue_EnqueueAfterClose(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		QueueID:  "test-campaign",
		SendRate: 0,
		Workers:  1,
	}

	sender := &mockSender{}
	queue, err := NewInMemory(ctx, cfg, sender)
	if err != nil {
		t.Fatalf("NewInMemory() failed: %v", err)
	}

	// Close the queue
	queue.Close()

	// Try to enqueue after close
	msg := mail.NewMsg()
	if err := msg.To("test@example.com"); err != nil {
		t.Fatalf("Failed to set To: %v", err)
	}

	err = queue.Enqueue(ctx, msg)
	if err == nil {
		t.Error("Enqueue() should return error after Close()")
	}
	if err != nil && err.Error() != "queue is closed" {
		t.Errorf("Expected 'queue is closed' error, got: %v", err)
	}

	queue.Wait()
}

func TestMemoryQueue_SendError(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		QueueID:  "test-campaign",
		SendRate: 0,
		Workers:  1,
	}

	var sendAttempts atomic.Int32
	expectedErr := errors.New("send failed")

	conn := &mockConn{
		sendFunc: func(msg ...*mail.Msg) error {
			sendAttempts.Add(1)
			return expectedErr
		},
	}

	sender := &mockSender{
		connFunc: func() (Conn, error) {
			return conn, nil
		},
	}

	queue, err := NewInMemory(ctx, cfg, sender)
	if err != nil {
		t.Fatalf("NewInMemory() failed: %v", err)
	}

	// Enqueue a message
	msg := mail.NewMsg()
	if err := msg.To("test@example.com"); err != nil {
		t.Fatalf("Failed to set To: %v", err)
	}
	if err := queue.Enqueue(ctx, msg); err != nil {
		t.Fatalf("Enqueue() failed: %v", err)
	}

	queue.Close()
	queue.Wait()

	// Verify send was attempted
	if sendAttempts.Load() < 1 {
		t.Error("Expected at least one send attempt")
	}
}

func TestMemoryQueue_ConnectionRecreation(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		QueueID:  "test-campaign",
		SendRate: 0,
		Workers:  1,
	}

	var connCount atomic.Int32
	var sendCount atomic.Int32

	sender := &mockSender{
		connFunc: func() (Conn, error) {
			connCount.Add(1)
			conn := &mockConn{
				sendFunc: func(msg ...*mail.Msg) error {
					count := sendCount.Add(1)
					// First send fails, subsequent sends succeed
					if count == 1 {
						return errors.New("temporary error")
					}
					return nil
				},
			}
			return conn, nil
		},
	}

	queue, err := NewInMemory(ctx, cfg, sender)
	if err != nil {
		t.Fatalf("NewInMemory() failed: %v", err)
	}

	// Initial connection count should be 1
	if connCount.Load() != 1 {
		t.Errorf("Expected 1 initial connection, got %d", connCount.Load())
	}

	// Enqueue two messages
	for i := 0; i < 2; i++ {
		msg := mail.NewMsg()
		if err := msg.To("test@example.com"); err != nil {
			t.Fatalf("Failed to set To: %v", err)
		}
		if err := queue.Enqueue(ctx, msg); err != nil {
			t.Fatalf("Enqueue() failed: %v", err)
		}
	}

	queue.Close()
	queue.Wait()

	// Should have recreated connection after error
	if connCount.Load() < 2 {
		t.Errorf("Expected at least 2 connections (recreation after error), got %d", connCount.Load())
	}

	// Should have attempted both sends
	if sendCount.Load() < 2 {
		t.Errorf("Expected at least 2 send attempts, got %d", sendCount.Load())
	}
}

func TestMemoryQueue_ConcurrentEnqueue(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		QueueID:  "test-campaign",
		SendRate: 0,
		Workers:  2,
	}

	var sendCount atomic.Int32
	sender := &mockSender{
		connFunc: func() (Conn, error) {
			return &mockConn{
				sendFunc: func(msg ...*mail.Msg) error {
					sendCount.Add(1)
					return nil
				},
			}, nil
		},
	}

	queue, err := NewInMemory(ctx, cfg, sender)
	if err != nil {
		t.Fatalf("NewInMemory() failed: %v", err)
	}

	// Concurrently enqueue messages from multiple goroutines
	numGoroutines := 5
	messagesPerGoroutine := 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				msg := mail.NewMsg()
				if err := msg.To("test@example.com"); err != nil {
					t.Errorf("Failed to set To: %v", err)
					return
				}
				if err := queue.Enqueue(ctx, msg); err != nil {
					t.Errorf("Enqueue() failed: %v", err)
					return
				}
			}
		}()
	}

	// Wait for all enqueues to complete
	wg.Wait()
	queue.Close()
	queue.Wait()

	// Verify all messages were sent
	expectedCount := int32(numGoroutines * messagesPerGoroutine)
	if sendCount.Load() != expectedCount {
		t.Errorf("Expected %d messages sent, got %d", expectedCount, sendCount.Load())
	}
}
