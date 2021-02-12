package workers

import (
	"errors"
	"sync"
)

type RunFunc func(interface{})

type WorkerPool struct {
	// The worker's run function
	run RunFunc

	// The channel for workers to listen for jobs
	jobs chan interface{}

	// The channel to stop a certain number of workers
	stop chan struct{}

	// The size of this worker pool (number of workers)
	size      int
	sizeMutex sync.Mutex

	// The number of busy workers in this worker pool
	busy      int
	busyMutex sync.Mutex
}

// Create a new WorkerPool with an initial worker count
func NewPool(size int, run RunFunc) *WorkerPool {
	pool := &WorkerPool{
		run:  run,
		jobs: make(chan interface{}),
		stop: make(chan struct{}),
		size: size,
		busy: 0,
	}
	// spawn workers up to the limit
	pool.createWorkers(size)
	return pool
}

// Create a new WorkerPool with an initial worker count and job buffer size
//
// The job buffer allows new jobs to be queued without blocking if
// all the workers are busy
func NewBufferedPool(size, bufSize int, run RunFunc) *WorkerPool {
	pool := &WorkerPool{
		run:  run,
		jobs: make(chan interface{}, bufSize),
		stop: make(chan struct{}),
		size: size,
		busy: 0,
	}
	// spawn workers up to the limit
	pool.createWorkers(size)
	return pool
}

// Add a job to this WorkerPool
func (w *WorkerPool) Run(data interface{}) {
	w.jobs <- data
}

// Scale the WorkerPool up to a new specified size
//
// Not goroutine-safe; use the scale operations in the proper order
func (w *WorkerPool) ScaleUp(newSize int) error {
	if newSize <= w.size {
		return errors.New("the new size must be greater than the current size")
	}

	w.sizeMutex.Lock()
	delta := newSize - w.size
	w.size = newSize
	w.sizeMutex.Unlock()

	w.createWorkers(delta)
	return nil
}

// Scale the WorkerPool down to a new specified size
//
// Not goroutine-safe; use the scale operations in the proper order
func (w *WorkerPool) ScaleDown(newSize int) error {
	if newSize >= w.size {
		return errors.New("the new size must be less than the current size")
	}

	w.sizeMutex.Lock()
	delta := w.size - newSize
	w.size = newSize
	w.sizeMutex.Unlock()

	for i := 0; i < delta; i++ {
		w.stop <- struct{}{}
	}
	return nil
}

// Stop the WorkerPool by closing all channels and stopping all workers
func (w *WorkerPool) Stop() {
	close(w.jobs)
	close(w.stop)
}

// Get the total number of workers in this WorkerPool
func (w *WorkerPool) Size() int {
	return w.size
}

// Get the number of busy workers in this WorkerPool
func (w *WorkerPool) Busy() int {
	w.busyMutex.Lock()
	defer w.busyMutex.Unlock()
	return w.busy
}

func (w *WorkerPool) createWorkers(count int) {
	for i := 0; i < count; i++ {
		go func() {
			for {
				select {
				case job, ok := <-w.jobs:
					if !ok {
						return
					}
					w.incBusy()
					w.run(job)
					w.decBusy()
				case <-w.stop:
					return
				}
			}
		}()
	}
}

func (w *WorkerPool) incBusy() {
	w.busyMutex.Lock()
	w.busy++
	w.busyMutex.Unlock()
}

func (w *WorkerPool) decBusy() {
	w.busyMutex.Lock()
	w.busy--
	w.busyMutex.Unlock()
}
