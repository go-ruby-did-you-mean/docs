<!-- SPDX-License-Identifier: BSD-3-Clause -->
# `go-ruby-did-you-mean` library-level benchmark harness

Reproducible, cross-runtime benchmark of the **pure-Go
`go-ruby-did-you-mean/did-you-mean` library** against the reference Ruby runtimes
(MRI, MRI + YJIT, JRuby, TruffleRuby). It measures the `did_you_mean` spell-
suggestion primitive through the Go API, isolated from the rbgo interpreter, so
the numbers answer: *is the pure-Go matcher as fast as the reference runtime's
own stdlib `did_you_mean` — and does it beat MRI + YJIT?*

## Scope: the pure-compute matcher

`did_you_mean` splits into two halves. The interpreter-tied half — the
`NameError` / `NoMethodError` / `KeyError` hooks, the per-error checkers that
gather candidate names off a live object, and the message formatter — belongs to
the Ruby host. The other half, ranking spelling suggestions for a misspelled name
against a dictionary, is **pure, deterministic computation**, and that is exactly
what this library ports and what this harness benchmarks:
`DidYouMean::SpellChecker#correct`, the sole public entry point, which internally
runs the Jaro–Winkler and Levenshtein metrics over every candidate. No object
introspection and no formatting are involved, so every op is reproducible.

## Layout

- `go/`               — self-contained Go driver; `go.mod` pins the **published**
  library by pseudo-version (no `replace`).
- `ruby/did_you_mean.rb` — the equivalent workload; `ruby/_harness.rb` is the
  shared timer.
- `run.sh`            — runs every available runtime and prints one Markdown
  table per operation (ns/op + ratio vs MRI).

## Run

```sh
bash benchmarks/run.sh
```

Environment knobs: `OUTER` (timed passes, default 25), `WARM` (untimed warm-up
passes, default 3), and `RUBY`/`JRUBY`/`TRUFFLERUBY` to select runtime binaries.
If a runtime does not ship `did_you_mean`, `gem install did_you_mean` installs
the MRI default-gem version.

## Method

Each process runs `WARM` untimed passes (to let the JVM/GraalVM JITs warm up),
then `OUTER` timed passes of a fixed inner loop, timed with a monotonic clock;
the **best** pass is reported as **ns/op**. Interpreter start-up is outside the
timed region. The Go driver and the Ruby script build the **identical** fixed
dictionary and the identical three input batches, and each op's full ranked
suggestion output is verified **byte-identical** across all four runtimes and the
Go driver (`CHECK=1 go run .` / `CHECK=1 ruby ruby/did_you_mean.rb`) before
timing — a strong equality proof, since the string encodes every suggestion and
its rank for every input. Published, dated results are in
[`../docs/performance.md`](../docs/performance.md).

## The three ops

| Op | What it exercises |
| --- | --- |
| `correct-hit` | 16 near-typos that each land on a suggestion — the full pipeline: Jaro–Winkler ranking **plus** the Levenshtein mistype/fallback filter. |
| `correct-miss` | 16 inputs far from every candidate — a **Jaro–Winkler-only** sweep (all rejected below threshold, so no Levenshtein runs). |
| `correct-batch` | the realistic mixed workload: the 16 hits and 16 misses interleaved (32 lookups). |
