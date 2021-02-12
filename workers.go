package workers

import (
	"errors"
	"sync"
)

type RunFunc func(...interface{})

type WorkerPool struct {
	// The worker's run function
	run RunFunc

	// The channel for workers to listen for jobs
	jobs chan []interface{}

	// The channel to stop a certain number of workers
	stop chan struct{}

	// The size of this worker pool (number of workers)
	size      int
	sizeMutex sync.Mutex

	// The number of busy workers in this worker pool
	busy      int
	busyMutex sync.Mutex

	// The number of workers waiting to close
	closing      int
	closingMutex sync.Mutex
}

// Create a new WorkerPool with an initial worker count
//
// Panics when size < 0
func NewPool(size int, run RunFunc) *WorkerPool {
	if size < 0 {
		panic("size must be greater than zero")
	}
	pool := &WorkerPool{
		run:  run,
		jobs: make(chan []interface{}),
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
//
// Panics when size < 0
func NewBufferedPool(size, bufSize int, run RunFunc) *WorkerPool {
	if size < 0 {
		panic("size must be greater than zero")
	}
	pool := &WorkerPool{
		run:  run,
		jobs: make(chan []interface{}, bufSize),
		stop: make(chan struct{}),
		size: size,
		busy: 0,
	}
	// spawn workers up to the limit
	pool.createWorkers(size)
	return pool
}

// Add a job to this WorkerPool
func (w *WorkerPool) Run(data ...interface{}) {
	w.jobs <- data
}

// Resize the WorkerPool by scaling up or down to accommodate a new size
func (w *WorkerPool) ScaleTo(newSize int) error {
	if newSize < w.size {
		return w.ScaleDown(newSize)
	}
	if newSize > w.size {
		return w.ScaleUp(newSize)
	}
	return errors.New("newSize must not be equal to the current size")
}

// Scale the WorkerPool up to a new specified size
//
// Safe to run in the background.
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
// Blocks until all workers have been stopped.
// Safe to run in the background.
func (w *WorkerPool) ScaleDown(newSize int) error {
	if newSize < 0 || newSize >= w.size {
		return errors.New("the new size must be between zero and the current size")
	}

	w.sizeMutex.Lock()
	delta := w.size - newSize
	w.size = newSize
	w.sizeMutex.Unlock()

	w.modClose(delta)
	for i := 0; i < delta; i++ {
		w.stop <- struct{}{}
		w.modClose(-1)
	}
	return nil
}

// Stop the WorkerPool by closing all channels and stopping all workers
func (w *WorkerPool) Stop() {
	close(w.jobs)
	close(w.stop)
}

// Stop the WorkerPool and keep track of the channels waiting to close
// by sending a closing signal to each worker. Slower than Stop()
//
// Blocks until all workers that were running have been closed. Workers
// stopped due to down-scaling do not cause this function to block.
func (w *WorkerPool) StopAndCount() {
	_ = w.ScaleDown(0)
	close(w.jobs)
	close(w.stop)
}

// Get the total number of workers in this WorkerPool
func (w *WorkerPool) Size() int {
	w.sizeMutex.Lock()
	defer w.sizeMutex.Unlock()
	return w.size
}

// Get the number of busy workers in this WorkerPool
func (w *WorkerPool) Busy() int {
	w.busyMutex.Lock()
	defer w.busyMutex.Unlock()
	return w.busy
}

// Get the number of workers currently waiting for jobs
//
// Equivalent to Size() - Busy()
func (w *WorkerPool) Waiting() int {
	return w.Size() - w.Busy() // mutex methods
}

// Get the number of workers waiting to close in this WorkerPool
func (w *WorkerPool) Excess() int {
	w.closingMutex.Lock()
	defer w.closingMutex.Unlock()
	return w.closing
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
					w.run(job...)
					w.decBusy()
				case <-w.stop:
					return
				}
			}
		}()
	}
}

func (w *WorkerPool) modClose(change int) {
	w.closingMutex.Lock()
	w.closing += change
	w.closingMutex.Unlock()
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
