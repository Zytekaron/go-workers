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

func TestWorkerPool_ScaleRandom(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	pool := NewPool(10, func(interface{}) {})

	// scaling random in goroutines has the
	// potential to cause scaling issues

	// this is mostly used to ensure that upsizing
	// *during* a downsize doesn't cause problems
	// assuming you always upsize AFTER a downsize
	// (goroutines here prevent that order)

	go pool.ScaleUp(100)
	<-time.After(1 * time.Millisecond)
	go pool.ScaleDown(25)
	<-time.After(1 * time.Millisecond)
	go pool.ScaleUp(80)
	<-time.After(1 * time.Millisecond)
	go pool.ScaleUp(125)
	<-time.After(1 * time.Millisecond)
	pool.ScaleDown(60)

	<-time.After(1 * time.Millisecond)
	pool.ScaleDown(50)

	<-time.After(1 * time.Millisecond)
	if pool.size != 50 {
		t.Error("pool size should be 50, not", pool.size)
	}
}
