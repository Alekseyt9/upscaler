package workerpool

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool_AddTask(t *testing.T) {
	var count int32
	task := func() {
		atomic.AddInt32(&count, 1)
	}
	wp := New(3)
	defer wp.Close()

	for i := 0; i < 10; i++ {
		wp.AddTask(task)
	}
	time.Sleep(1 * time.Second)

	assert.Equal(t, int32(10), count, "expected 10 tasks to be executed")
}

func TestWorkerPool_Close(t *testing.T) {
	var count int32
	task := func() {
		<-time.After(100 * time.Millisecond)
		atomic.AddInt32(&count, 1)
	}

	wp := New(3)
	for i := 0; i < 10; i++ {
		wp.AddTask(task)
	}

	wp.Close()
	assert.Equal(t, int32(10), count, "expected 10 tasks to be executed before close")
}

func TestWorkerPool_WithCustomQueueSize(t *testing.T) {
	var count int32
	task := func() {
		atomic.AddInt32(&count, 1)
	}

	wp := NewWithQS(3, 10)
	defer wp.Close()

	for i := 0; i < 10; i++ {
		wp.AddTask(task)
	}

	time.Sleep(1 * time.Second)
	assert.Equal(t, int32(10), count, "expected 10 tasks to be executed")
}
