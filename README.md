# Problem statement

In Go, prior to Go 1.25, there weren't any good ways of testing time-based code. Thus, a common pattern was to rely on a fakeClock, such as github.com/jonboulle/clockwork and have go routines and methods swallow a clock as an argument, so that production code could rely on a clockwork.NewRealClock() while test code could rely on a `clockwork.NewFakeClock()` that they could advance at will in the code.

However now that Go 1.25 has a very nice built-in test framework for such time-dependent code, it is time for code that relies on such FakeClock to look into migrating to `synctest` tests.
The big question is: how does that work, can we do it without revamping the whole codebase?

This is a little repo with a few example to show what are one's options, a few tests to show what works and what doesn't work, and finally a `FakeClock` compatible interface that relies on `synctest` ready to be used as a drop-in replacement for `clockwork.NewFakeClock()` while relying on `synctest` time and `Wait` flows.
