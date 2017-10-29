package retry

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestDefaultNew(t *testing.T) {
	t.Parallel()

	r := New()
	if r.Tries != MaxRetries {
		t.Fatalf("bad tries config, got %d want %d", r.Tries, MaxRetries)
	}

	if r.Tries != MaxRetries {
		t.Fatalf("bad tries config, got %d want %d", r.Tries, MaxRetries)
	}

	err := r.Do(happy)
	if err != nil {
		t.Errorf("should have succeeded without an error, got %v. Retryer state %#v", err, r)
	}

	err = r.Do(sad)
	if err == nil {
		t.Errorf("should have failed with an error. Retryer state %#v", r)
	}
}

func TestInfNumberOfTries(t *testing.T) {
	t.Parallel()

	r := New(Tries(0))
	if r.Tries != 0 {
		t.Fatalf("bad tries config, got %d want %d", r.Tries, 0)
	}

	for i := 0; i <= 1000; i++ {
		err := r.Do(sad)
		if err == nil {
			t.Errorf("should have failed with an error")
		}
	}
}

func TestSleep(t *testing.T) {
	t.Parallel()

	// using a ever failing function
	tcs := []struct {
		sleep int
		tries int
		wait  int
	}{
		{
			sleep: 50,
			tries: 1,
			wait:  100,
		},
		{
			sleep: 50,
			tries: 5,
			wait:  100,
		},
	}

	for i, tc := range tcs {
		r := New(Sleep(tc.sleep), Tries(tc.tries))
		ch := make(chan error)

		start := time.Now()
		go func() {
			ch <- r.Do(sad)
		}()

		select {
		case <-time.After(time.Duration((tc.sleep*tc.tries)+tc.wait) * time.Millisecond):
			t.Errorf("tc %d: should have slept only for %v took too long", i, time.Duration((tc.sleep*tc.tries)+tc.wait)*time.Millisecond)
		case err := <-ch:
			// have we finished sooner?
			if d := time.Since(start); d < time.Duration(tc.sleep*tc.tries)*time.Millisecond {
				t.Errorf("tc %d: retryer didn't sleep for the desired time, ended after %v", i, d)
			}
			if err == nil {
				t.Errorf("tc %d: should have failed with an error, Retryer state %#v", i, r)
			}
		}
	}
}

func TestSleepFn(t *testing.T) {
	t.Parallel()

	// we want to sleep for 100+200+300 ms = 600 ms -> linearly growing backoff
	sleepFn := func(attempts int) {
		sleep := time.Duration(100*attempts) * time.Millisecond
		time.Sleep(sleep)
	}

	r := New(SleepFn(sleepFn), Tries(3))
	ch := make(chan error)
	start := time.Now()
	go func() {
		ch <- r.Do(sad)
	}()

	expectedSleepDur := time.Duration(600 * time.Millisecond)
	select {
	case <-time.After(expectedSleepDur + 100*time.Millisecond):
		t.Errorf("should have slept only for %v took too long", expectedSleepDur)
	case err := <-ch:
		// have we finished sooner?
		if d := time.Since(start); d < expectedSleepDur {
			t.Errorf("retryer didn't sleep for the desired time, ended after %v", d)
		}
		if err == nil {
			t.Errorf("should have failed with an error, Retryer state %#v", r)
		}
	}
}

func TestPanicRecoveryEnabled(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("retryer with default panic recovery option - Recover - shouldn't have panicked: %v", r)
		}
	}()

	// testing retryer with panic recovery
	err := New(Recover()).Do(panicked)
	if err == nil {
		t.Error("expected an error containing panic stacktracke")
	}
	if !strings.HasPrefix(err.Error(), "retryer has recovered panic: explicit trigger of panic goroutine") ||
		!strings.Contains(err.Error(), "stack") {
		t.Errorf("unexpected error returned from panic recovery %v", err)
	}
}

func TestPanicRecoveryDisabled(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("retryer should have panicked with default settings")
		}
	}()

	New().Do(panicked)
}
func TestEnsureFn(t *testing.T) {
	t.Parallel()

	touched := false
	toggler := func(error) { touched = true }

	r := New(Ensure(toggler))
	err := r.Do(sad)
	if err == nil {
		t.Errorf("should have succeeded without an error, got %v. Retryer state %#v", err, r)
	}
	if !touched {
		t.Error("ensure function wasn't called")
	}
}

func TestErrorFnOn(t *testing.T) {
	t.Parallel()

	// errorTypeC is not amongst the ones to retry on, on 5th attempt we expect to succeed
	fn := func() error { return &errorTypeC{S: "error c triggered"} }
	ab := attemptsBased{
		succeedOnNth: 5,
		fn:           fn,
	}

	r := New(Tries(6), On([]error{&errorTypeA{}, &errorTypeB{}}))
	err := r.Do(ab.run)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if r.attempts != 6 {
		t.Errorf("incorrect attempts count, got %d want 5", r.attempts)

	}

	// errorTypeC is not amongst the ones to retry on, we try only once
	fn = func() error { return &errorTypeC{S: "error c triggered"} }
	err = New(Tries(1), On([]error{&errorTypeA{}, &errorTypeC{}})).Do(fn)
	if err == nil {
		t.Errorf("expected errorType error")
	}

	// errorTypeC is not amongst the ones to retry on, as we supply empty slice of errors, we try only once
	fn = func() error { return &errorTypeC{S: "error c triggered"} }
	err = New(Tries(1), On([]error{})).Do(fn)
	if err == nil {
		t.Errorf("expected errorType error")
	}
}

func TestErrorFnNot(t *testing.T) {
	t.Parallel()

	// errorTypeC is amongst the ones to ignore in the not slice on, we expect a success after 1st run
	errMsg := "error c triggered"
	fn := func() error { return &errorTypeC{S: errMsg} }

	r := New(Tries(3), Not([]error{&errorTypeA{}, &errorTypeC{}}))
	err := r.Do(fn)
	if err == nil {
		t.Errorf("unexpected error: %v", err)
	}
	if r.attempts != 1 {
		t.Errorf("incorrect attempts count, got %d want 1", r.attempts)
	}

	// errorTypeC is not amongst the ones to retry on, we try only once
	fn = func() error { return &errorTypeC{S: "error c triggered"} }
	err = New(Tries(1), Not([]error{&errorTypeA{}, &errorTypeB{}})).Do(fn)
	if err == nil {
		t.Errorf("expected errorType error")
	}
	if !strings.Contains(err.Error(), errMsg) {
		t.Errorf("unexpected error returned, got: type:%v msg:'%v',  want to contain: type:%v msg:'%v'", reflect.TypeOf(err), err.Error(), "errorTypeC", errMsg)
	}
}

func TestAfterEachFail(t *testing.T) {
	t.Parallel()

	ch := make(chan error)
	fn := func(err error) { ch <- err }

	bodyFn := func() error { return errors.New("generic error") }
	ab := attemptsBased{
		succeedOnNth: 5,
		fn:           bodyFn,
	}

	go func() {
		New(AfterEachFail(fn), Tries(5)).Do(ab.run)
	}()

	for i := 0; i < 5; i++ {
		select {
		case <-ch:
			continue
		case <-time.After(50 * time.Millisecond):
			t.Fatal("didn't receive an error from the fail callback function")
		}
	}
}

func TestCombinedOptions(t *testing.T) {
	t.Parallel()

	ab := attemptsBased{
		succeedOnNth: 3,
		fn:           sad,
	}
	touched := false
	toggler := func(error) { touched = true }

	r := New(Tries(5), Sleep(10), Ensure(toggler))
	err := r.Do(ab.run)
	if err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if !touched {
		t.Error("ensure function wasn't called")
	}
}

func TestSleepFnPriorityOverSleep(t *testing.T) {
	t.Parallel()

	// we want to sleep for 100+200+300 ms = 600 ms -> linearly growing backoff
	sleepFn := func(attempts int) {
		sleep := time.Duration(100*attempts) * time.Millisecond
		time.Sleep(sleep)
	}

	// the Sleep(1000) won't be used, if yes, it will be caught by the timers below
	r := New(SleepFn(sleepFn), Sleep(1000), Tries(3))
	ch := make(chan error)
	start := time.Now()
	go func() {
		ch <- r.Do(sad)
	}()

	expectedSleepDur := time.Duration(600 * time.Millisecond)
	select {
	case <-time.After(expectedSleepDur + 100*time.Millisecond):
		t.Errorf("should have slept only for %v took too long", expectedSleepDur)
	case err := <-ch:
		// have we finished sooner?
		if d := time.Since(start); d < expectedSleepDur {
			t.Errorf("retryer didn't sleep for the desired time, ended after %v", d)
		}
		if err == nil {
			t.Errorf("should have failed with an error, Retryer state %#v", r)
		}
	}
}

type errorTypeA struct {
	s string
}

func (e *errorTypeA) Error() string {
	return e.s
}

type errorTypeB struct {
	s string
}

func (e *errorTypeB) Error() string {
	return e.s
}

type errorTypeC struct {
	S string
}

func (e *errorTypeC) Error() string {
	return e.S
}

func happy() error {
	_ = 2 + 3
	return nil
}

func sad() error {
	_ = 2 + 3
	return errors.New("error on primitive addition")
}

func panicked() error {
	panic("explicit trigger of panic")
}

type attemptsBased struct {
	attempts     int
	succeedOnNth int
	fn           func() error
}

func (ab *attemptsBased) run() error {
	if ab.attempts == ab.succeedOnNth {
		return nil
	}

	ab.attempts++
	return ab.fn()
}
