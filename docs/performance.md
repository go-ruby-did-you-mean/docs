# Performance

`go-ruby-did-you-mean/did-you-mean` is the pure-Go, CGO-free library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's
`did_you_mean`. This page records a **real, library-level** benchmark of that
module's Go API against every reference runtime's own stdlib `did_you_mean`
SpellChecker, one row per representative spell-check workload. It is part of the
ecosystem-wide per-module parity suite, and **the bar is beating MRI + YJIT**,
not just plain MRI.

## Scope: the pure-compute matcher

`did_you_mean` has two halves. The **interpreter-tied** half — the `NameError` /
`NoMethodError` / `KeyError` hooks, the per-error checkers that gather candidate
names off a live object, and the message formatter — is the Ruby host's job. The
other half, **ranking spelling suggestions for a misspelled name against a
dictionary**, is pure, deterministic computation; that is what this library ports
and what this benchmark measures: `DidYouMean::SpellChecker#correct`, the module's
sole public entry point, which internally runs the **Jaro–Winkler** and
**Levenshtein** metrics over every candidate. No object introspection and no
message formatting are timed, which is exactly what makes every number
reproducible.

Because the Go module exports only `Correct` (the two distance metrics are
internal, just as they are effectively private in MRI's `SpellChecker`), the
benchmark drives them **through `Correct`** — the same observable operation every
Ruby runtime exposes — rather than fabricating a distance-only entry point.

## What is measured

Three workloads run over one **fixed, deterministic corpus**: a 36-word
dictionary of Ruby `Enumerable`-style identifier names (`map`, `flat_map`,
`collect`, `reduce`, `group_by`, …) and two fixed 16-input batches of query words.

| Op | What it exercises |
| --- | --- |
| `correct-hit` | 16 near-typos (`collekt`, `lenght`, `grup_by`, …) that each land on a suggestion — the **full pipeline**: Jaro–Winkler ranking **plus** the Levenshtein mistype/fallback filter |
| `correct-miss` | 16 inputs far from every candidate (`xyzzy`, `frobnicate`, `kubernetes`, …) — a **Jaro–Winkler-only sweep** (all rejected below threshold, so no Levenshtein runs), isolating the metric that runs on every candidate |
| `correct-batch` | the realistic mixed workload: the 16 hits and 16 misses interleaved (32 lookups), the shape a real `did_you_mean` lookup faces |

The **go-ruby** column drives this pure-Go library through its Go API; every other
column is that interpreter's own stdlib `DidYouMean::SpellChecker#correct`. The Go
and Ruby drivers build the **identical** dictionary and input batches and, before
any timing, each op's **full ranked suggestion output** is verified **byte-
identical across all four runtimes and the Go driver** (e.g. `correct-hit` maps
`collekt→collect`, `lenght→length`, `grup_by→group_by`, and drops `include` as an
exact match; `correct-miss` returns empty for all 16). The check string encodes
every suggestion and its rank for every input, so it only matches if the Jaro–
Winkler ordering, the two thresholds, and the Levenshtein filter all agree —
making it a strong ranked-output equality proof. So the comparison is the same
observable operation, apples-to-apples.

- **Host:** Apple M4 Max, macOS (`arm64-darwin`). **Date:** 2026-07-03.
- **Runtimes:** Go 1.26.4; `ruby 4.0.5 +PRISM` (MRI, the oracle) and
  `ruby --yjit`; `jruby 10.1.0.0` (OpenJDK 25); `truffleruby 34.0.1`
  (GraalVM CE Native). `did_you_mean` ships as a default gem in every one.
- **Method:** each process runs 3 untimed warm-up passes then 25 timed passes of
  a fixed inner loop, timed with a monotonic clock; the **best** pass is reported
  as **ns/op**. Interpreter start-up is outside the timed region, so the number is
  the operation's own cost, not `ruby file.rb` process cost. Numbers were stable
  to within a few percent across repeated runs; treat the gaps as approximate.
- Harness and drivers live in this repo under
  [`benchmarks/`](https://github.com/go-ruby-did-you-mean/docs/tree/main/benchmarks)
  (`go/`, `ruby/did_you_mean.rb`, `run.sh`). Reproduce: `bash benchmarks/run.sh`.

## Results (ns/op, best of 25)

| Op | go-ruby (pure Go) | MRI | MRI + YJIT | JRuby | TruffleRuby | **go vs YJIT** |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `correct-batch` | **67 860** | 2 504 580 | 716 170 | 882 336 | 158 404 | **10.6× faster** ✅ |
| `correct-hit` | **33 880** | 1 187 380 | 345 080 | 387 919 | 89 446 | **10.2× faster** ✅ |
| `correct-miss` | **33 732** | 1 298 080 | 344 380 | 437 305 | 81 339 | **10.2× faster** ✅ |

## The go-vs-YJIT verdict, per op

**The pure-Go library beats MRI + YJIT on every one of the three workloads** —
there is no op where YJIT wins, and the margin is a remarkably uniform **~10×**:

- **`correct-batch` — 10.6× faster** (67 860 ns vs 716 170 ns).
- **`correct-hit` — 10.2× faster** (33 880 ns vs 345 080 ns).
- **`correct-miss` — 10.2× faster** (33 732 ns vs 344 380 ns).

The reason the margin is so flat across all three is that `SpellChecker#correct`
is dominated by the **Jaro–Winkler sweep over every candidate**: it computes the
metric for all 36 dictionary words on every lookup, whether or not any survive.
`correct-miss` (Jaro–Winkler only, zero Levenshtein) and `correct-hit` (Jaro–
Winkler plus Levenshtein on the survivors) therefore cost almost the same —
Levenshtein runs on only a handful of survivors and is cheap next to the full
sweep. In MRI this sweep is per-character Ruby with `Array#each` closures and
match-flag arrays; YJIT strips interpreter overhead but not the method-dispatch
and allocation cost of the inner loops. The Go port runs the same code-point
comparisons over `[]rune` with no interpreter in the loop, so it pulls ~10× ahead
— and this is the exact shape of the payload `rbgo` binds.

The pure-Go library is also the **fastest implementation outright**: it beats even
TruffleRuby (the quickest interpreter here) by 2.3×–2.6× on all three ops, and
JRuby and plain MRI by far wider margins.

## Caveats

- **Cold-JIT framing.** JRuby and TruffleRuby are timed after the same 3 warm-up
  passes as everyone else, but 3 passes do **not** bring the JVM/GraalVM JITs to
  full steady state; read their columns as lightly-warmed, not as peak
  throughput. TruffleRuby in particular is still climbing here — its steady-state
  number would be lower. MRI, YJIT and Go reach their representative speed almost
  immediately, so the **go-vs-MRI and go-vs-YJIT columns are the load-bearing
  comparison**.
- **Matcher scope.** This measures the pure-compute suggestion matcher only; the
  interpreter-tied `did_you_mean` machinery (error hooks, per-error checkers,
  message formatter) is the Ruby host's job and is not measured — it is not part
  of this library.
- No number here is fabricated: all figures are measured on the host and date
  named above and reproduce with `bash benchmarks/run.sh`.
