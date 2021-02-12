package main

import (
	"fmt"
	"time"
	"workers"
)

// An optional struct to hold useful data
// about a job being executed in the pool
type Job struct {
	Value interface{}
}

func main() {
	// Create a new worker pool with 100 workers
	// and a function which handles jobs by printing them.
	// All 100 allocated workers here are created upfront.
	pool := workers.NewPool(100, func(i interface{}) {
		job := i.(*Job)
		fmt.Println("Worker ran with value:", job.Value)
	})
	// See NewBufferedPool (source/godoc) for buffered job queues
	// that will not block when all of the workers are busy

	// Run jobs!
	pool.Run(&Job{Value: 123}) // Worker ran with value: 123
	pool.Run(&Job{Value: "f"}) // Worker ran with value: f

	// Busy & total workers
	busyWorkers := pool.Busy()
	totalWorkers := pool.Size()
	// Not guaranteed (or likely) to print 2 for this example code!
	// WorkerPool#Run does not wait for a job to start; only for it
	// to be accepted by a worker, which will run it shortly after
	fmt.Printf("%d/%d workers are busy\n", busyWorkers, totalWorkers)

	// Automatic worker scaling (not required, but can be useful)
	go func() {
		// Make sure to avoid bouncing! Don't upscale with so many workers
		// that it sets off the downscale operation due to under-usage, or
		// remove enough workers to set off the upscale operation :)
		for {
			// upscale when the pool is using 95% or more of its workers
			if float64(pool.Busy()) > float64(pool.Size()) * 0.95 {
				// add 10% more workers
				// example: 100 -> 110 -> 121 -> 133 -> 146 -> 160
				count := float64(pool.Size()) * (1 + 0.10)
				_ = pool.ScaleUp(int(count))
			}

			// downscale when the pool is using 50% or less of its workers
			if float64(pool.Busy()) < float64(pool.Size()) * 0.50 {
				// remove 25% of the workers
				// example: 100 -> 75 -> 56 -> 42 -> 31 -> 23
				count := float64(pool.Size()) * (1 - 0.25)
				_ = pool.ScaleDown(int(count))
			}

			// run auto-scaler once per 5 seconds
			<-time.After(5 * time.Second)
		}
	}()

	fmt.Scanln() // Wait for user input before stopping the pool and exiting

	// Stop all workers once they complete their current job.
	// Any active workers may take time to stop (based on your run function).
	// If there is a job *queue* (see NewBufferedChannel), it will be discarded.
	pool.Stop()
}
