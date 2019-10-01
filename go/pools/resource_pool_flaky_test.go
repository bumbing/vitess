package pools

import (
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestIdleTimeout(t *testing.T) {
	ctx := context.Background()
	lastID.Set(0)
	count.Set(0)
	p := NewResourcePool(PoolFactory, 1, 1, 10*time.Millisecond, 0)
	defer p.Close()

	r, err := p.Get(ctx)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if count.Get() != 1 {
		t.Errorf("Expecting 1, received %d", count.Get())
	}
	if p.IdleClosed() != 0 {
		t.Errorf("Expecting 0, received %d", p.IdleClosed())
	}
	p.Put(r)
	if lastID.Get() != 1 {
		t.Errorf("Expecting 1, received %d", count.Get())
	}
	if count.Get() != 1 {
		t.Errorf("Expecting 1, received %d", count.Get())
	}
	if p.IdleClosed() != 0 {
		t.Errorf("Expecting 0, received %d", p.IdleClosed())
	}
	time.Sleep(15 * time.Millisecond)

	if count.Get() != 1 {
		t.Errorf("Expecting 1, received %d", count.Get())
	}
	if p.IdleClosed() != 1 {
		t.Errorf("Expecting 1, received %d", p.IdleClosed())
	}
	r, err = p.Get(ctx)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if lastID.Get() != 2 {
		t.Errorf("Expecting 2, received %d", count.Get())
	}
	if count.Get() != 1 {
		t.Errorf("Expecting 1, received %d", count.Get())
	}
	if p.IdleClosed() != 1 {
		t.Errorf("Expecting 1, received %d", p.IdleClosed())
	}

	// sleep to let the idle closer run while all resources are in use
	// then make sure things are still as we expect
	time.Sleep(15 * time.Millisecond)
	if lastID.Get() != 2 {
		t.Errorf("Expecting 2, received %d", count.Get())
	}
	if count.Get() != 1 {
		t.Errorf("Expecting 1, received %d", count.Get())
	}
	if p.IdleClosed() != 1 {
		t.Errorf("Expecting 1, received %d", p.IdleClosed())
	}
	p.Put(r)
	r, err = p.Get(ctx)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if lastID.Get() != 2 {
		t.Errorf("Expecting 2, received %d", count.Get())
	}
	if count.Get() != 1 {
		t.Errorf("Expecting 1, received %d", count.Get())
	}
	if p.IdleClosed() != 1 {
		t.Errorf("Expecting 1, received %d", p.IdleClosed())
	}

	// the idle close thread wakes up every 1/100 of the idle time, so ensure
	// the timeout change applies to newly added resources
	p.SetIdleTimeout(1000 * time.Millisecond)
	p.Put(r)

	time.Sleep(15 * time.Millisecond)
	if lastID.Get() != 2 {
		t.Errorf("Expecting 2, received %d", count.Get())
	}
	if count.Get() != 1 {
		t.Errorf("Expecting 1, received %d", count.Get())
	}
	if p.IdleClosed() != 1 {
		t.Errorf("Expecting 1, received %d", p.IdleClosed())
	}

	// Get and Put to refresh timeUsed
	r, err = p.Get(ctx)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	p.Put(r)
	p.SetIdleTimeout(10 * time.Millisecond)
	time.Sleep(15 * time.Millisecond)
	if lastID.Get() != 3 {
		t.Errorf("Expecting 3, received %d", lastID.Get())
	}
	if count.Get() != 1 {
		t.Errorf("Expecting 1, received %d", count.Get())
	}
	if p.IdleClosed() != 2 {
		t.Errorf("Expecting 2, received %d", p.IdleClosed())
	}
}
