# spec

[![Build Status](https://travis-ci.org/sclevine/spec.svg?branch=master)](https://travis-ci.org/sclevine/spec)
[![GoDoc](https://godoc.org/github.com/sclevine/spec?status.svg)](https://godoc.org/github.com/sclevine/spec)

Spec is a simple BDD test organizer for Go. It minimally extends the standard
library `testing` package by facilitating easy organization of Go 1.7+
[subtests](https://blog.golang.org/subtests).

Spec differs from other BDD libraries for Go in that it:
- Does not reimplement or replace any functionality of the `testing` package
- Does not provide an alternative test parallelization strategy to the `testing` package
- Does not provide assertions
- Does not encourage the use of dot-imports
- Does not reuse any closures between test runs (to avoid test pollution)
- Does not use global state, excessive interface types, or reflection

Spec is intended for gophers who want to write BDD tests in idiomatic Go using
the standard library `testing` package. Spec aims to do "one thing right,"
and does not provide a wide DSL or any functionality outside of test
organization.

### Features

- Clean, simple syntax
- Supports focusing and pending tests
- Supports sequential, random, reverse, and parallel test order
- Provides granular control over test order and subtest nesting
- Provides a test writer to manage test output
- Provides a generic, asynchronous reporting interface
- Provides multiple reporter implementations

### Notes

- Use `go test -v` to see individual subtests.

### Examples

[Most functionality is demonstrated here.](spec_test.go#L238)

Quick example:

```go
func TestObject(t *testing.T) {
    spec.Run(t, "object", func(t *testing.T, when spec.G, it spec.S) {
        var someObject *myapp.Object

        it.Before(func() {
            someObject = myapp.NewObject()
        })

        it.After(func() {
            someObject.Close()
        })

        it("should have some default", func() {
            if someObject.Default != "value" {
                t.Error("bad default")
            }
        })

        when("something happens", func() {
            it.Before(func() {
                someObject.Connect()
            })

            it("should do one thing", func() {
                if err := someObject.DoThing(); err != nil {
                    t.Error(err)
                }
            })

            it("should do another thing", func() {
                if result := someObject.DoOtherThing(); result != "good result" {
                    t.Error("bad result")
                }
            })
        }, spec.Random())

        when("some slow things happen", func() {
            it("should do one thing in parallel", func() {
                if result := someObject.DoSlowThing(); result != "good result" {
                    t.Error("bad result")
                }
            })

            it("should do another thing in parallel", func() {
                if result := someObject.DoOtherSlowThing(); result != "good result" {
                    t.Error("bad result")
                }
            })
        }, spec.Parallel())
    }, spec.Report(report.Terminal{}))
}
```

With less nesting:

```go
func TestObject(t *testing.T) {
    spec.Run(t, "object", testObject, spec.Report(report.Terminal{}))
}

func testObject(t *testing.T, when spec.G, it spec.S) {
    ...
}
```

For focusing/reporting across multiple files in a package:

```go
var suite spec.Suite

func init() {
    suite = spec.New("my suite", spec.Report(report.Terminal{}))
    suite("object", testObject)
    suite("other object", testOtherObject)
}

func TestObjects(t *testing.T) {
	suite.Run(t)
}

func testObject(t *testing.T, when spec.G, it spec.S) {
	...
}

func testOtherObject(t *testing.T, when spec.G, it spec.S) {
	...
}
```