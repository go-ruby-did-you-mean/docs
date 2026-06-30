# go-ruby-did-you-mean documentation

**Ruby's `did_you_mean` matcher in pure Go — Jaro–Winkler & Levenshtein spelling suggestions, MRI byte-exact, no cgo.**

`go-ruby-did-you-mean/did-you-mean` is a faithful, pure-Go (zero cgo) reimplementation of Ruby's [`did_you_mean`](https://docs.ruby-lang.org/en/master/DidYouMean.html),
matching reference Ruby (MRI). The module path is `github.com/go-ruby-did-you-mean/did-you-mean`.

It is the backend bound into [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)
by `rbgo` as a native module — just like
[go-ruby-regexp](https://github.com/go-ruby-regexp) and
[go-ruby-erb](https://github.com/go-ruby-erb). The dependency runs the other way:
this library has **no dependency on the Ruby runtime**.

!!! success "Status: complete"
    **Complete — MRI byte-exact.** Faithful, line-for-line port of Ruby's `DidYouMean::SpellChecker#correct`: **normalization**, the **Jaro–Winkler** filter and ranking, the **Levenshtein** mistype filter, and the **misspell fallback**, with both metrics comparing **Unicode code points**. Validated by a **differential oracle** against the system `ruby` — every ranked array reproduced byte-for-byte — at 100% coverage, `gofmt` + `go vet` clean, CI green across the six 64-bit Go targets and three OSes.

## Quick taste

```go
methods := []string{"map", "select", "reject", "collect", "flatten"}

didyoumean.Correct("collekt", methods) // [collect]
didyoumean.Correct("rerject", methods) // [reject]
didyoumean.Correct("xyz", methods)     // []
```

## Repositories

| Repo | What it is |
| --- | --- |
| [`did-you-mean`](https://github.com/go-ruby-did-you-mean/did-you-mean) | the library — Ruby's did_you_mean in pure Go |
| [`docs`](https://github.com/go-ruby-did-you-mean/docs) | this documentation site (MkDocs Material, versioned with mike) |
| [`go-ruby-did-you-mean.github.io`](https://github.com/go-ruby-did-you-mean/go-ruby-did-you-mean.github.io) | the organization landing page (Hugo) |
| [`brand`](https://github.com/go-ruby-did-you-mean/brand) | logo and brand assets |

## Principles

- **Pure Go, `CGO_ENABLED=0`** — trivial cross-compilation, a single static
  binary, no C toolchain.
- **MRI byte-exact.** Output matches reference Ruby, validated by a differential
  oracle against the `ruby` binary.
- **Standalone & reusable.** A standalone module bound by `rbgo`; no dependency on
  the Ruby runtime — the dependency runs the other way.
- **100% test coverage** is the target, enforced as a CI gate, across 6 arches
  and 3 OSes.

## Where to go next

- [Why pure Go](why.md) — why this slice of Ruby is deterministic enough to live
  as a standalone, interpreter-independent Go library.
- [Usage & API](api.md) — the public surface and worked examples.
- [Roadmap](roadmap.md) — what is done and what is downstream by design.

Source lives at [github.com/go-ruby-did-you-mean/did-you-mean](https://github.com/go-ruby-did-you-mean/did-you-mean).
