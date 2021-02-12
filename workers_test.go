package workers

import (
	"math/rand"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	NewPool(10, func(interface{}) {})
}

func TestWorkerPool_ScaleUp(t *testing.T) {
	pool := NewPool(10, func(interface{}) {})

	err := pool.ScaleUp(5)
	if err == nil {
		t.Error("pool must not accept a lower value when scaling up")
	}

	_ = pool.ScaleUp(15)
	if pool.size != 15 {
		t.Error("pool size should be 15, not", pool.size)
	}
}

func TestWorkerPool_ScaleDown(t *testing.T) {
	pool := NewPool(10, func(interface{}) {})

	err := pool.ScaleDown(15)
	if err == nil {
		t.Error("pool must not accept a higher value when scaling down")
	}

	_ = pool.ScaleDown(5)
	if pool.size != 5 {
		t.Error("pool size should be 5, not", pool.size)
	}
}

func TestWorkerPool_Busy(t *testing.T) {
	pool := NewPool(10, func(interface{}) {
		<-make(chan bool) // block forever (until test ends)
	})

	for i := 0; i < 5; i++ {
		go pool.Run(struct{}{})
	}
	<-time.After(10 * time.Microsecond) // wait for goroutines to start jobs
	if pool.Busy() != 5 {
		t.Error("busy workers should equal 5, not", pool.Busy())
	}

	for i := 0; i < 10; i++ {
		// 5 running, 10 additional | max 10
		go pool.Run(struct{}{})
	}
	<-time.After(10 * time.Microsecond) // wait for goroutines to start jobs
	if pool.Busy() != 10 {
		t.Error("busy workers should equal 10, not", pool.Busy())
	}
}

func TestNewBufferedPool(t *testing.T) {
	NewBufferedPool(10, 5, func(interface{}) {})
}

func TestWorkerPool_ScaleRandom(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	pool := NewPool(10, func(interface{}) {
		<-make(chan bool) // block forever (until test ends)
	})

	// scaling random in goroutines has the
	// potential to cause scaling issues

	// this is mostly used to ensure that upsizing
	// *during* a downsize doesn't cause problems
	// assuming you always upsize AFTER a downsize
	// (goroutines here prevent that order)

	go pool.ScaleTo(100)
	go pool.ScaleTo(25)
	go pool.ScaleTo(80)
	go pool.ScaleTo(125)
	go pool.ScaleTo(60)

	// run this one last
	<-time.After(time.Microsecond)
	pool.ScaleTo(50)

	<-time.After(1 * time.Millisecond)
	if pool.Size() != 50 {
		t.Error("pool size should be 50, not", pool.Size())
	}
}
