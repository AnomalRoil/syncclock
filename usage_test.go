package syncclock_test

import (
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/AnomalRoil/syncclock"
	"github.com/jonboulle/clockwork"
)

func myFunc(clock clockwork.Clock, i *int) {
	clock.Sleep(3 * time.Second)
	*i += 1
}

func TestMyFunc(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var i int
		c := syncclock.NewFakeClock()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			myFunc(c, &i)
			wg.Done()
		}()
		synctest.Wait()

		if i != 0 {
			t.Fatal("expected i == 0 before Advance")
		}

		c.Advance(1 * time.Hour)
		wg.Wait()
		synctest.Wait()

		if i != 1 {
			t.Fatal("expected i == 1 after Advance")
		}
	})
}
