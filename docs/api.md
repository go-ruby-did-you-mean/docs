# Usage & API

The public API lives at the module root (`github.com/go-ruby-did-you-mean/did-you-mean`). It is **Ruby-shaped but
Go-idiomatic**: the surface mirrors MRI's `did_you_mean`, while following Go conventions —
an explicit `error` where Ruby raises, value types, no global state.

!!! success "Status: implemented"
    The library is built and importable as `github.com/go-ruby-did-you-mean/did-you-mean`, bound into
    `rbgo` as a native module; see [Roadmap](roadmap.md).

## Install

```sh
go get github.com/go-ruby-did-you-mean/did-you-mean
```

## Worked example

```go
package main

import (
	"fmt"

	didyoumean "github.com/go-ruby-did-you-mean/did-you-mean"
)

func main() {
	methods := []string{"map", "select", "reject", "collect", "flatten"}

	didyoumean.Correct("collekt", methods) // [collect]
	didyoumean.Correct("rerject", methods) // [reject]
	didyoumean.Correct("xyz", methods)     // []
	_ = fmt.Sprint
}
```

## Shape

```go
// Correct returns the ranked spelling suggestions for input drawn from
// dictionary, exactly as Ruby's
// DidYouMean::SpellChecker.new(dictionary:).correct(input).
func Correct(input string, dictionary []string) []string
```

## MRI conformance

Correctness is defined by reference Ruby. A **differential oracle** runs a wide
corpus through both the system `ruby` and this library and compares the results
**byte-for-byte** — not approximated from memory. The oracle tests skip
themselves where `ruby` is not on `PATH` (e.g. the qemu arch lanes), so the
cross-arch builds still validate the library.

## Relationship to Ruby

`go-ruby-did-you-mean/did-you-mean` is **standalone and reusable**, and is the backend bound into
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) by `rbgo` as a
native module — the same way [go-ruby-regexp](https://github.com/go-ruby-regexp)
and [go-ruby-erb](https://github.com/go-ruby-erb) are bound. The dependency runs
the other way: this library has no dependency on the Ruby runtime.
