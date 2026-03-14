// Package syncclock provides a near drop-in replacement for
// [clockwork.FakeClock] that is backed by Go 1.25's [testing/synctest] package
// instead of maintaining its own fake time.
//
// Rather than tracking waiters and manually expiring timers, [SyncClock]
// delegates to the real [time] package and uses [synctest.Wait] to
// synchronise goroutines inside a synctest bubble.  This means all standard
// [time] functions (After, Sleep, NewTimer, …) work as expected and time only
// advances when [SyncClock.Advance] is called.
//
// # Migration from clockwork
//
// In most cases migrating is a two-step process:
//
//  1. Wrap your test function with [synctest.Test].
//  2. Replace [clockwork.NewFakeClock] with [NewFakeClock] (or
//     [clockwork.NewFakeClockAt] with [NewFakeClockAt]).
//
// Production code can keep using [clockwork.NewRealClock] unchanged because
// [SyncClock] implements [clockwork.Clock].
//
// # Not implemented
//
// The following clockwork.FakeClock methods have no equivalent in SyncClock
// because synctest's cooperative scheduling model makes them unnecessary:
//
//   - BlockUntil – use [synctest.Wait] instead.
//   - BlockUntilContext – use [synctest.Wait] instead.
//
// # Limitations
//
// A [SyncClock] must be created inside a [synctest.Test] bubble.
//
// Because [synctest] does not consider blocking I/O operations (e.g. network
// reads) as durable blocks, HTTP servers and similar code must be mocked using
// [net.Pipe] or equivalent.  See the [synctest documentation] for details.
//
// [synctest documentation]: https://pkg.go.dev/testing/synctest
package syncclock

import (
	"testing/synctest"
	"time"

	"github.com/jonboulle/clockwork"
)

// Compile-time interface compatibility checks.
var (
	_ clockwork.Clock  = &SyncClock{}
	_ clockwork.Ticker = &SyncTicker{}
	_ clockwork.Timer  = &SyncTimer{}
	_ FakeClock        = (*clockwork.FakeClock)(nil)
	_ FakeClock        = &SyncClock{}
)

// FakeClock is a minimal interface satisfied by both [clockwork.FakeClock] and
// [SyncClock].  It allows code that only needs the Advance method to accept
// either implementation without a type assertion.
type FakeClock interface {
	clockwork.Clock
	Advance(d time.Duration)
}

// SyncTicker wraps a [time.Ticker] and implements [clockwork.Ticker].
type SyncTicker struct{ *time.Ticker }

// Chan returns the channel on which ticks are delivered.
func (st *SyncTicker) Chan() <-chan time.Time {
	return st.C
}

// SyncTimer wraps a [time.Timer] and implements [clockwork.Timer].
type SyncTimer struct{ *time.Timer }

// Chan returns the channel on which the timer fires.
func (st *SyncTimer) Chan() <-chan time.Time {
	return st.C
}

// SyncClock implements [clockwork.Clock] on top of [testing/synctest].
// All time operations delegate to the standard [time] package; advancing the
// clock is achieved by sleeping inside the synctest bubble and then calling
// [synctest.Wait] to let other goroutines observe the new time.
//
// A SyncClock must only be used inside a [synctest.Test] bubble.
type SyncClock struct{}

// Advance moves the fake clock forward by d and waits for all goroutines in
// the synctest bubble to reach a durable blocking state.  This is the
// equivalent of [clockwork.FakeClock.Advance].
func (s *SyncClock) Advance(d time.Duration) {
	time.Sleep(d)
	synctest.Wait()
}

// After mimics [time.After]; it waits for d to elapse on the fake clock, then
// sends the current time on the returned channel.
func (s *SyncClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// Sleep blocks until d has elapsed on the fake clock.
func (s *SyncClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

// Now returns the current time of the fake clock.
func (s *SyncClock) Now() time.Time {
	return time.Now()
}

// Since returns the time elapsed since t on the fake clock.
func (s *SyncClock) Since(t time.Time) time.Duration {
	return s.Now().Sub(t)
}

// Until returns the duration until t on the fake clock.
func (s *SyncClock) Until(t time.Time) time.Duration {
	return t.Sub(s.Now())
}

// NewTicker returns a [clockwork.Ticker] that ticks every d.
func (s *SyncClock) NewTicker(d time.Duration) clockwork.Ticker {
	return &SyncTicker{time.NewTicker(d)}
}

// NewTimer returns a [clockwork.Timer] that fires after d.
func (s *SyncClock) NewTimer(d time.Duration) clockwork.Timer {
	return &SyncTimer{time.NewTimer(d)}
}

// AfterFunc returns a [clockwork.Timer] that invokes f after d has elapsed.
func (s *SyncClock) AfterFunc(d time.Duration, f func()) clockwork.Timer {
	return &SyncTimer{time.AfterFunc(d, f)}
}

// NewFakeClock creates a [SyncClock] whose initial time is [time.Now] inside
// the synctest bubble.  It must be called from within a [synctest.Test] bubble.
func NewFakeClock() *SyncClock {
	return NewFakeClockAt(time.Now())
}

// NewFakeClockAt creates a [SyncClock] whose initial time is t.
// It must be called from within a [synctest.Test] bubble.
//
// It panics if t is before midnight UTC 2000-01-01 because synctest does not
// support rewinding time before that point.
func NewFakeClockAt(t time.Time) *SyncClock {
	if t.Compare(time.Date(2000, time.January, 01, 0, 0, 0, 0, time.UTC)) < 0 {
		panic("synctest limitation: we cannot set time earlier than midnight UTC 2000-01-01")
	}
	time.Sleep(time.Until(t))
	synctest.Wait()
	return &SyncClock{}
}
