# Roadmap

`go-ruby-did-you-mean/did-you-mean` is grown **test-first**, each capability differential-tested against MRI
rather than built in isolation. Ruby's did_you_mean — the deterministic,
interpreter-independent slice — is **complete**.

| Stage | What | Status |
| --- | --- | --- |
| Normalization | Downcase and strip `@` sigils from input and every candidate, so `@ivar` and `IVar` compare alike — the first step of MRI's `SpellChecker#correct`. | **Done** |
| Jaro–Winkler filter & ranking | Keep candidates whose Jaro–Winkler similarity clears the threshold (`0.834` for inputs longer than 3 characters, else `0.77`), drop an exact match of the original, and rank survivors by similarity descending — a stable sort then reverse, matching Ruby's tie order. | **Done** |
| Levenshtein mistype filter | Keep the ranked candidates within an edit distance of `⌈len/4⌉` of the input, exactly as MRI's mistype filter does. | **Done** |
| Misspell fallback | If nothing survives, return the first ranked candidate whose Levenshtein distance is strictly less than the shorter of the two word lengths — Ruby's stable-sorted, input-order fallback. | **Done** |
| Unicode-correct metrics | `Jaro` / `JaroWinkler.distance` and `Levenshtein.distance` ported exactly, comparing Unicode code points (not bytes) like Ruby's `String#each_codepoint`, so multibyte words (`café`, `naïve`) rank identically. | **Done** |
| Differential oracle & coverage | Golden vectors captured from MRI plus a broad corpus of typos, transpositions, case variants, multibyte words and no-match inputs fed to the system `ruby`'s `SpellChecker#correct`, reproduced byte-for-byte. 100% coverage, gofmt + go vet clean, green across all six 64-bit Go arches and three OSes. | **Done** |

## Documented out-of-scope boundaries

These are **deliberate**, recorded so the module's surface is unambiguous:

- **No interpreter.** The library implements the deterministic algorithm; it
  never runs arbitrary Ruby. Anything that needs a live binding or evaluation is
  the consumer's job — that is why `rbgo` binds this module rather than the
  reverse.
- **Reference is reference Ruby (MRI).** Conformance targets MRI's behaviour;
  differences across MRI releases are matched to the reference used by the
  differential oracle.
- **Standalone & reusable.** The module has no dependency on the Ruby runtime;
  the dependency runs the other way.

See [Usage & API](api.md) for the surface and [Why pure Go](why.md) for the
deterministic/interpreter split.
