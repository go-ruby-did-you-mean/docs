# Performance

`go-ruby-did-you-mean/did-you-mean` is the pure-Go library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's `did-you-mean`. This
page records the **methodology** for the comparative benchmark of that module
against the reference Ruby runtimes, part of the ecosystem-wide per-module parity
suite.

## What is measured

The **same** Ruby script — a representative `did_you_mean` workload — is run under every
runtime. `rbgo`'s number reflects **this pure-Go library doing the work**; every
other column is that interpreter's own `did-you-mean` stdlib. So the comparison is the
**Ruby-visible operation**, apples-to-apples across interpreters. The script
prints a deterministic checksum and its output is checked **byte-identical to
MRI** before timing.

- **Method:** best-of-N wall time (best, not mean, to suppress scheduler noise);
  single-shot processes, no warm-up beyond the script's own loop.
- **Runtimes:** `ruby` (MRI, the oracle) and `ruby --yjit`; `jruby` (OpenJDK);
  `truffleruby` (GraalVM CE Native).
- The benchmark script and harness live in rbgo's repo under
  [`bench/modules/`](https://github.com/go-embedded-ruby/ruby/tree/main/bench/modules)
  (`did-you-mean.rb` + `run.sh`). Reproduce:
  `RBGO=./rbgo TRUFFLE=truffleruby bash bench/modules/run.sh 5`.

## Result (best of 5, ms)

| Runtime | time | vs MRI |
| --- | ---: | ---: |
| **rbgo** (go-ruby-did-you-mean) | 40 | 0.03× |
| MRI (ruby 4.0.5) | 1450 | 1.00× |
| MRI + YJIT | 400 | 0.28× |
| JRuby 10.1.0.0 | 2660 | 1.83× |
| TruffleRuby 34.0.1 | 480 | 0.33× |

rbgo runs on **go-ruby-did-you-mean** and is **~33x faster than MRI** here (0.03x): MRI's `DidYouMean::SpellChecker` edit-distance search is Ruby-coded, so the compiled pure-Go spell-checker dominates. The standout win of the wave-3 suite.

!!! note "Honest framing"
    JRuby and TruffleRuby are timed **cold, single-shot**, so they carry JVM /
    Graal startup on every run — read them as one-shot `ruby file.rb` costs, the
    same way `rbgo` and MRI are measured, not as steady-state JIT numbers. Rows
    that complete in well under ~200 ms carry the most relative noise; treat
    their ratios as order-of-magnitude. These are **real measured numbers** from
    the 2026-06-30 run (Apple M-series; `ruby 4.0.5 +PRISM`, `jruby 10.1.0.0`,
    `truffleruby 34.0.1`) — nothing is fabricated or cherry-picked.
