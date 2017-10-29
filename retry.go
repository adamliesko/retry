package retry

import (
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"time"
)

// MaxRetries is the maximum number of retries.
const MaxRetries = 10

type Retryer struct {
	Tries        int
	On           []error       // On is the slice of errors, on which Retryer will retry a function
	Not          []error       // Not is the slice of errors which Retryer won't consider as needed to retry
	SleepDur     time.Duration // Sleep duration in ms
	Recover bool // If enabled, panics will be recovered.

	SleepFn 	func(int) // Custom sleep function with access to the current # of attempts
	EnsureFn     func(error) // DeferredFn is called after repeated function finishes, regardless of outcome
	ErrorFn      func(error) // DeferredFn is called after repeated function finishes, regardless of outcome

	attempts int
}

// New creates a Retryer with applied options.
func New(opts ...func(*Retryer)) *Retryer {
	r := &Retryer{Tries: MaxRetries}

	// setup the options
	for _, o := range opts {
		o(r)
	}

	return r
}

// Resets resets the state of the attempts to 0
func (r *Retryer) Reset() {
	r.attempts = 0
}

// Do calls the passed in function until it succeeds. The behaviour of the retry mechanism heavily relies on the config
// of the Retryer.
func (r *Retryer) Do(fn func() error) (err error) {
	r.Reset()

	// define the deferred functions
	if r.Recover{
		defer func(){
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprintf("retryer has recovered panic: %v %s", r, debug.Stack()))
			}
			}()
	}
	if r.EnsureFn != nil {
		defer r.EnsureFn(err)
	}
	if r.ErrorFn != nil {
		defer func() {
			if err != nil {
				r.ErrorFn(err)
			}
		}()
	}

	// retry the function
	for r.attempts = 0; r.attempts < r.Tries; r.attempts++ {
		err = fn()
		if r.succeeded(err) {
			return
		}
		r.trySleep()
	}

	return fmt.Errorf("max number of retries reached: %d, last error %v", r.attempts,err)
}

func (r *Retryer) succeeded(err error) bool {
	for _, e := range r.Not {
		if reflect.TypeOf(err) == reflect.TypeOf(e) {
			return true
		}
	}
	for _, e := range r.On {
		if reflect.TypeOf(err) == reflect.TypeOf(e) {
			return false
		}
	}

	return err == nil
}

func (r *Retryer) trySleep() {
	if r.SleepFn != nil {
		r.SleepFn(r.attempts)
	} else if r.SleepDur != 0 {
		time.Sleep(r.SleepDur)
	}
}