# retry
![Build Status](https://secure.travis-ci.org/adamliesko/retry.svg)
![Go Report Card](https://goreportcard.com/badge/github.com/adamliesko/retry)
![GoDoc](https://godoc.org/github.com/adamliesko/retry?status.svg)
![Coverage Status](https://img.shields.io/coveralls/adamliesko/retry.svg)

Retry is a Go package which wraps a function and retries it until it succeeds, not returning na error. Multiple retry-able
options are provided, based on number of attempts, sleep after failed attempt, errors to retry on or skip, post attempts
callback etc. Usable for interaction with flake-y web services and similar unreliable sources of frustration.

![Lego Batman gif retry](https://media.giphy.com/media/JJhiRdcYfcokU/giphy.gif)
## Installation

```
go get -u github.com/adamliesko/retry
```

## Usage

In the simplest and default configuration, with 10 retries it's only about creating a new Retryer and calling Do, with
the desired function. If the failed functions failes after 10 retries a custom error of Max Attempts reached is returned.
Befor each call of Retryer's Do() it's state is reset, which makes the Retryer reusable.
```go
func poll() error{
    return external.IsItDone() 
}
    
result := retry.New().Do(poll)
```

The usual usage would be to use either directly function, which return an error or wrap the function call with a function,
which sets the error according to the inner function output.

```go
// can be used directly
func poll() error{
    return external.IsItDone() 
}

// has to be wrapped
func pollNoError bool{
	return external.HasItSucceeded()
}

func wrappedPoll() error{
	if pollNoError{
		return nil
	}
	
	return errors.New("pollNoError has failed")
}
```

Options on Retryer (listed below in greater detail):
- constant sleep delay after a failure
- custom function sleep delay (e.g. exponential backoff)
- recovery of panics
- calling an ensure function, regardless of the Retryer's work inside, once that it finishes
- calling a custom function after each failure
- ignoring certain errors
- retrying only on certain errors

#### Sleeping constant duration of 100ms after each failed attempt
```go
func poll() error {
    return external.IsItDone() 
}
    
err := retry.New(retry.Sleep(100))
result := r.Do(poll)
```


#### Using an exponential back off (or any other custom function) after each failed attempt
```go
func poll() error {
    return external.IsItDone()
}
        
sleepFn := func(attempts int) {
    sleep := time.Duration(2^attempts) * time.Millisecond
    time.Sleep(sleep)
}

err := retry.New(retry.SleepFn(sleepFn)).Do(poll)
```

#### Calling an ensure function, which is called after whole Retryer execution
```go
func poll() error {
    return external.IsItDone()
}
        
func ensure(err error){
	fmt.Println("ensure will be called regardless of err value")
}

err := retry.New(retry.Ensure(ensure)).Do(poll)
```

#### Ignoring failures with errors of listed types (whitelist) and considering them as success
```go
type MyError struct {
}

func (e MyError) Error() string {
	return "this is my custom error"
}

func poll() error {
    return external.IsItDone()
}
        
err := retry.New(Not([]errors{MyError{}})).Do(poll)
```

#### Retrying only on listed error types (blacklist), other errors will be considered as success
```go
type MyError struct {
}

func (e MyError) Error() string {
	return "this is my custom error"
}

func poll() error {
    return external.IsItDone()
}
        
err := retry.New(On([]errors{MyError{}})).Do(poll)
```

#### Retry allows to combine many options in one Retryer. The code block below will enable:

- recovery of panics
- attempting to call the function up to 15 times
- sleeping for 200 ms after each failed attempt
- printing the failures to the Stdout

```go
func poll() error {
    return external.IsItDone()
}
     
failCallback := func(err error){ fmt.Println("failed with error",error) }

r := retry.New(retry.Sleep(200), retry.Tries(15), retry.Recover(), retry.AfterEachFail(failCallback)
err := r.Do(poll)
```

## License
See [LICENSE](LICENSE).