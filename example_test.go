package syncclock_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/AnomalRoil/syncclock"
	"github.com/jonboulle/clockwork"
)

func myFunc(clock clockwork.Clock, i *atomic.Int64) {
	clock.Sleep(3 * time.Second)
	i.Add(1)
}

func TestMyFunc(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var i atomic.Int64
		c := syncclock.NewFakeClock()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			myFunc(c, &i)
			wg.Done()
		}()
		synctest.Wait()

		if v := i.Load(); v != 0 {
			t.Fatalf("expected 0, got %d", v)
		}

		c.Advance(1 * time.Hour)
		wg.Wait()
		synctest.Wait()

		if v := i.Load(); v != 1 {
			t.Fatalf("expected 1, got %d", v)
		}
	})
}
