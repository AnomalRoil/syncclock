package syncclock

import (
	"testing/synctest"
	"time"

	"github.com/jonboulle/clockwork"
)

var _ clockwork.Clock = &SyncClock{}
var _ clockwork.Ticker = &SyncTicker{}
var _ clockwork.Timer = &SyncTimer{}

type SyncTicker struct{ *time.Ticker }

func (st *SyncTicker) Chan() <-chan time.Time {
	return st.C
}

type SyncTimer struct{ *time.Timer }

func (sm *SyncTimer) Chan() <-chan time.Time {
	return sm.C
}

type SyncClock struct{}

func (s *SyncClock) Advance(d time.Duration) {
	time.Sleep(d)
	synctest.Wait()
}

func (s *SyncClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (s *SyncClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (s *SyncClock) Now() time.Time {
	return time.Now()
}

func (s *SyncClock) Since(t time.Time) time.Duration {
	return s.Now().Sub(t)
}

func (s *SyncClock) Until(t time.Time) time.Duration {
	return t.Sub(s.Now())
}

func (s *SyncClock) NewTicker(d time.Duration) clockwork.Ticker {
	return &SyncTicker{time.NewTicker(d)}
}

func (s *SyncClock) NewTimer(d time.Duration) clockwork.Timer {
	t := &SyncTimer{time.NewTimer(d)}
	return t
}

func (s *SyncClock) AfterFunc(d time.Duration, f func()) clockwork.Timer {
	return &SyncTimer{time.AfterFunc(d, f)}
}

func NewFakeClock() *SyncClock {
	return NewFakeClockAt(time.Now())
}

// NewFakeClockAt returns a FakeClock initialised at the given time.Time.
func NewFakeClockAt(t time.Time) *SyncClock {
	if t.Compare(time.Now()) < 0 {
		panic("synctest limitation: we cannot set time earlier than midnight UTC 2000-01-01")
	}
	time.Sleep(time.Until(t))
	synctest.Wait()
	return &SyncClock{}
}
