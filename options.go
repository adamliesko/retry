package retry

import "time"

// On configures the Retryer to retry function call on any of the passed in errors.
func On(errors []error) func(r *Retryer) {
	return func(r *Retryer) {
		r.On = errors
	}
}

// Not configures the Retryer to ignore all of the passed in errors and in case of them appearing doesn't retry
// function anymore.
func Not(errors []error) func( *Retryer) {
	return func(r *Retryer) {
		r.Not = errors
	}
}

// Ensure sets a deferred function to be called, regardless of Retryer succeeding in running the function with or without
// an error.
func Ensure(ensureFn func(error)) func( *Retryer) {
	return func(r *Retryer) {
		r.EnsureFn = ensureFn
	}
}

// Recover configures the Retryer to recover panics, returning an error containing the panic and it's stacktrace.
func Recover() func( *Retryer){
	return func(r *Retryer) {
		r.Recover = true
	}
}

// Tries configures to Retryer to keep calling the function until it succeeds tries-times.
func Tries(tries int) func(r *Retryer) {
	return func(r *Retryer) {
		if tries == 0 {
			tries = MaxRetries
		}
		r.Tries = tries
	}
}

// AfterEachFail configures the Retryer to call failFn function after each of the failed attempts.
func AfterEachFail(failFn func(error)) func(*Retryer) {
	return func(r *Retryer) {
		//r.AfterEachFailure = failFn()
	}
}

// Sleep configures to Retryer sleep a certain duration [ms] after each failed attempt.
func Sleep(dur int) func(*Retryer) {
	return func(r *Retryer) {
		r.SleepDur = time.Duration(dur) * time.Millisecond
	}
}


// SleepFn configures the Retryer to call a custom, caller supplied function after each failed attempt. SleepFn takes
// precedence over a set sleep duration.
func SleepFn(sleepFn func(int)) func(*Retryer) {
	return func(r *Retryer) {
		r.SleepFn = sleepFn
	}
}
