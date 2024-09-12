// Package workerpool provides a simple implementation of a worker pool to manage concurrent tasks.
package workerpool

import "sync"

// Task represents a unit of work to be executed by the worker pool.
type Task func()

// WorkerPool manages a pool of workers to execute tasks concurrently.
type WorkerPool struct {
	tasks chan Task
	wg    sync.WaitGroup
}

const workerPoolQueueSize = 5

// New creates a new WorkerPool with the specified number of workers and a default queue size.
// Parameters:
//   - workerCount: the number of workers to create
//
// Returns a pointer to a WorkerPool.
func New(workerCount int) *WorkerPool {
	return NewWithQS(workerCount, workerPoolQueueSize)
}

// NewWithQS creates a new WorkerPool with the specified number of workers and a specified queue size.
// Parameters:
//   - workerCount: the number of workers to create
//   - queueCount: the size of the task queue
//
// Returns a pointer to a WorkerPool.
func NewWithQS(workerCount int, queueCount int) *WorkerPool {
	wp := &WorkerPool{
		tasks: make(chan Task, queueCount),
	}
	for range workerCount {
		wp.wg.Add(1)
		go wp.worker()
	}
	return wp
}

// AddTask adds a task to the worker pool's task queue.
// Parameters:
//   - task: the task to add to the queue
func (wp *WorkerPool) AddTask(task Task) {
	wp.tasks <- task
}

// Close signals the worker pool to stop accepting new tasks and waits for all workers to complete.
func (wp *WorkerPool) Close() {
	close(wp.tasks)
	wp.wg.Wait()
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for task := range wp.tasks {
		task()
	}
}
