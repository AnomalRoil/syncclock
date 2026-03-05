# syncclock

A near drop-in replacement for [`clockwork.FakeClock`](https://github.com/jonboulle/clockwork) backed by Go 1.25's [`testing/synctest`](https://pkg.go.dev/testing/synctest) package.

Instead of maintaining its own fake time and waiter lists, `syncclock` delegates to the real `time` package inside a `synctest` bubble and uses `synctest.Wait` to synchronise goroutines. Time only advances when `Advance` is called, or if you are `Sleep`ing and calling `synctest.Wait()` yourself.

## Install

```
go get github.com/AnomalRoil/syncclock@latest
```

Requires **Go 1.25** or later.

## Usage

```go
package mypkg_test

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
```

### Migration from clockwork

1. Wrap your test function with `synctest.Test`.
2. Replace `clockwork.NewFakeClock()` with `syncclock.NewFakeClock()` (or `clockwork.NewFakeClockAt(t)` with `syncclock.NewFakeClockAt(t)`).

Production code can keep using `clockwork.NewRealClock()` unchanged because `SyncClock` implements `clockwork.Clock`.

## Not implemented

The following `clockwork.FakeClock` methods have no equivalent in `SyncClock` because `synctest`'s cooperative scheduling model makes them unnecessary:

| Method | Replacement |
|---|---|
| `BlockUntil(n int)` | Use `synctest.Wait()` as it blocks until all goroutines in the bubble are durably blocked, which is a stronger guarantee than counting waiters. |
| `BlockUntilContext(ctx context.Context, n int) error` | Use `synctest.Wait()` for the same reason. |

`NewRealClock()` is also not provided; keep using `clockwork.NewRealClock()` in production since `SyncClock` is only meaningful inside a `synctest` bubble.

## Limitations

- A `SyncClock` **must** be created inside a `synctest.Test` bubble.
- `NewFakeClockAt` panics if the requested time is before midnight UTC 2000-01-01 (a `synctest` limitation).
- `synctest` does not consider blocking I/O operations (e.g. network reads) as durable blocks, so HTTP servers and similar code must be mocked using `net.Pipe` or equivalent. See the [synctest docs](https://pkg.go.dev/testing/synctest#hdr-Example__HTTP_100_Continue) for details.

## License

MIT
