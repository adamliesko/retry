# retry
[![Build Status](https://secure.travis-ci.org/adamliesko/retry.svg)](http://travis-ci.org/adamliesko/retry)
[![Go Report Card](https://goreportcard.com/badge/github.com/adamliesko/retry)](https://goreportcard.com/report/github.com/adamliesko/retry)
[![GoDoc](https://godoc.org/github.com/adamliesko/retry?status.svg)](https://godoc.org/github.com/adamliesko/retry)
[![Coverage Status](https://img.shields.io/coveralls/adamliesko/retry.svg)](https://coveralls.io/r/adamliesko/retry?branch=master)

Retry is a Go packag which wraps a function and retries it until it succeeds, not returning na error. Multiple retry-able
options are provided, based on number of attempts, sleep after failed attempt, errors to retry on or skip, post attempts
callback etc. Usable for interaction with flake-y web services and similar unreliable sources of frustration.

## Installation

```
go get -u github.com/adamliesko/retry
```

## Usage

#### Sleeping constant duration after each failed attempt
```

```


#### Using exponential back off (or any other custom function) after each failed attempt
```
```


#### Retry allows to combine multiple options in one Retryer. The code block below will enable:

- recovery of panics
- attempting to call the function up to 15 times
- sleeping for 200 ms after each failed attempt
- ignoring errors of certain type

```

```
