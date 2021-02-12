package main

import (
	"fmt"
	"github.com/Zytekaron/go-workers"
	"time"
)

// An optional struct to hold useful data
// about a job being executed in the pool
type Job struct {
	Value interface{}
}

func main() {
	// Create a new worker pool with 100 workers
	// and a function which handles jobs by printing them.
	// All 10 allocated workers here are created upfront.
	pool := workers.NewPool(10, func(i ...interface{}) {
		job := i[0].(*Job)
		fmt.Println("Worker ran with values:", job.Value, i[1])
		<-time.After(time.Millisecond)
	})
	// See NewBufferedPool (source/godoc) for buffered job queues
	// that will not block when all of the workers are busy

	// Run jobs!
	pool.Run(&Job{Value: 123}, ":)") // Worker ran with values: 123 :)
	pool.Run(&Job{Value: "f"}, ":(") // Worker ran with values: f :(
	pool.Run(&Job{Value: true}, "E") // Worker ran with values: true E

	<-time.After(time.Microsecond) // wait for Run workers to start doing work
	// Busy & total workers
	busy := pool.Busy()
	total := pool.Size()
	waiting := pool.Waiting()
	// Not guaranteed (or likely) to print 2 for this example code!
	// WorkerPool#Run does not wait for a job to start; only for it
	// to be accepted by a worker, which will run it shortly after
	fmt.Printf("%d/%d workers are busy. %d workers are waiting for jobs.\n", busy, total, waiting) // 3/10, 7

	// You can use ScaleTo if you don't know which direction to scale in.
	// Scale methods will block until all of the excess workers
	// have been stopped, or new workers have been created.
	_ = pool.ScaleTo(5)
	// or ScaleUp/ScaleDown
	_ = pool.ScaleUp(15)
	go pool.ScaleDown(1)
	<-time.After(time.Microsecond) // wait for ScaleDown to finish variable changes, but not for workers to stop

	// or you can run it in a new goroutine and check up on the excess workers
	excess := pool.Excess()
	fmt.Println("There are", excess, "excess workers (active workers waiting to be stopped)") // 2 excess

	// or, you can implement automatic scaling (see below)

	// Automatic worker scaling (not required, but can be useful)
	go func() {
		// Make sure to avoid bouncing! Don't upscale with so many workers
		// that it sets off the downscale operation due to under-usage, or
		// remove enough workers to set off the upscale operation :)
		for {
			// Upscale when the pool is using 95% or more of its workers
			if float64(pool.Busy()) > float64(pool.Size())*0.95 {
				// Add 10% more workers
				// Example: 100 -> 110 -> 121 -> 133 -> 146 -> 160
				count := float64(pool.Size()) * (1 + 0.10)
				_ = pool.ScaleUp(int(count))
			}

			// downscale when the pool is using 50% or less of its workers
			if float64(pool.Busy()) < float64(pool.Size())*0.50 {
				// Remove 25% of the workers
				// Example: 100 -> 75 -> 56 -> 42 -> 31 -> 23
				count := float64(pool.Size()) * (1 - 0.25)
				_ = pool.ScaleDown(int(count))
			}

			// run auto-scaler once per 5 seconds
			<-time.After(5 * time.Second)
		}
	}()

	_, _ = fmt.Scanln() // Wait for user input before stopping the pool and exiting

	// Stop all workers once they complete their current job.
	// Any active workers may take time to stop (based on your run function).
	// If there is a job *queue* (see NewBufferedChannel), it will be discarded.
	pool.Stop()
	// this method does not keep track of excess workers, but is a little faster
	// and closes the job channels immediately / doesn't maintain a counter

	// Or, you can stop it like this
	pool.StopAndCount()
	// which will block until all workers have been stopped, or run it in another
	// goroutine and keep track of lingering workers by checking the excess
	pool.Excess()
}
