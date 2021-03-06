package requests

// SOURCE: https://github.com/valyala/fasthttp/blob/d4c739eee589f96f10f07f05db40f1cfb5ad0bd9/timer.go
// LICENCE: https://github.com/valyala/fasthttp/blob/d4c739eee589f96f10f07f05db40f1cfb5ad0bd9/LICENSE

/*
The MIT License (MIT)

Copyright (c) 2015-present Aliaksandr Valialkin, VertaMedia, Kirill Danshin, Erik Dubbelboer, FastHTTP Authors

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

*/
import (
	"sync"
	"time"
)

func initTimer(t *time.Timer, timeout time.Duration) *time.Timer {
	if t == nil {
		return time.NewTimer(timeout)
	}
	if t.Reset(timeout) {
		panic("BUG: active timer trapped into initTimer()")
	}
	return t
}

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		// Collect possibly added time from the channel
		// if timer has been stopped and nobody collected its' value.
		select {
		case <-t.C:
		default:
		}
	}
}

// AcquireTimer returns a time.Timer from the pool and updates it to
// send the current time on its channel after at least timeout.
//
// The returned Timer may be returned to the pool with ReleaseTimer
// when no longer needed. This allows reducing GC load.
func AcquireTimer(timeout time.Duration) *time.Timer {
	v := timerPool.Get()
	if v == nil {
		return time.NewTimer(timeout)
	}
	t, _ := v.(*time.Timer)
	initTimer(t, timeout)
	return t
}

// ReleaseTimer returns the time.Timer acquired via AcquireTimer to the pool
// and prevents the Timer from firing.
//
// Do not access the released time.Timer or read from it's channel otherwise
// data races may occur.
func ReleaseTimer(t *time.Timer) {
	stopTimer(t)
	timerPool.Put(t)
}

var timerPool sync.Pool
