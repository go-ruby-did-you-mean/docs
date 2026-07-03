// Copyright (c) the go-ruby-did-you-mean/did-you-mean authors
// SPDX-License-Identifier: BSD-3-Clause
//
// Library-level benchmark driver for the pure-Go go-ruby-did-you-mean library.
// It exercises the public spell-suggestion primitive — didyoumean.Correct, the
// port of Ruby's DidYouMean::SpellChecker#correct — over an identical,
// deterministic corpus (a fixed dictionary of Ruby-ish identifier names plus
// three fixed input batches), so the ns/op numbers compare the pure-Go matcher
// against each Ruby runtime's own stdlib did_you_mean SpellChecker.
//
// Correct is the module's only exported entry point; the Jaro–Winkler and
// Levenshtein metrics are internal, so the benchmark drives them through
// Correct exactly as rbgo and every Ruby runtime do. The three batches isolate
// the algorithm's cost bands: `correct-miss` is a Jaro–Winkler-only sweep (every
// candidate is rejected below threshold, so no Levenshtein runs), while
// `correct-hit` and `correct-batch` add the Levenshtein mistype/fallback filter.
//
// With CHECK=1 it prints one "CHECK\t<label>\t<value>" line per op, where value
// is the full ranked suggestion output joined into a string, used to prove the
// Go output is byte-identical to MRI (the oracle) before any timing is trusted.
package main

import (
	"fmt"
	"os"
	"strings"

	didyoumean "github.com/go-ruby-did-you-mean/did-you-mean"
)

// dict is a fixed, realistic dictionary of Ruby Enumerable-ish identifier
// names, the kind of candidate list did_you_mean ranks a mistyped method
// against. Order is significant (Correct's stable sort and misspell fallback
// depend on it), and it is byte-for-byte the same list the Ruby workload builds.
var dict = []string{
	"map", "flat_map", "collect", "select", "filter_map", "reject",
	"detect", "find", "find_all", "each", "each_with_index", "each_with_object",
	"inject", "reduce", "group_by", "partition", "sort", "sort_by",
	"min", "max", "min_by", "max_by", "sum", "count",
	"tally", "zip", "take", "take_while", "drop", "drop_while",
	"first", "last", "length", "size", "include", "index",
}

// hitInputs are near-typos that each land on a suggestion: they clear the
// Jaro–Winkler threshold and pass (or fall back through) the Levenshtein
// mistype filter. This exercises the full pipeline.
var hitInputs = []string{
	"collekt", "selekt", "rejekt", "detct", "injekt",
	"reduse", "sord", "lenght", "grup_by", "parttion",
	"flat_mp", "cont", "firts", "include", "idnex", "sizze",
}

// missInputs are far from every candidate: none clears the Jaro–Winkler
// threshold, so Correct returns an empty list after a pure Jaro–Winkler sweep
// with no Levenshtein work — isolating the metric that runs on every candidate.
var missInputs = []string{
	"xyzzy", "qwerty", "zzzzzz", "asdfgh", "plugh", "frobnicate",
	"wibble", "quux", "0xdeadbeef", "supercalifragilistic", "hjkl", "vmware",
	"kubernetes", "photosynthesis", "abcdefg", "mmmmmm",
}

// batchInputs is the realistic mixed workload: hits and misses interleaved,
// the shape a real did_you_mean lookup faces.
var batchInputs = func() []string {
	out := make([]string, 0, len(hitInputs)+len(missInputs))
	for i := 0; i < len(hitInputs); i++ {
		out = append(out, hitInputs[i])
		if i < len(missInputs) {
			out = append(out, missInputs[i])
		}
	}
	return out
}()

// correctAll runs Correct over every input against the fixed dictionary and
// returns the concatenation of every ranked suggestion, so the returned string
// is a total, order-sensitive fingerprint of the whole batch's output — it
// matches MRI only if every suggestion and its rank agree.
func correctAll(inputs []string) string {
	var b strings.Builder
	for _, in := range inputs {
		b.WriteString(in)
		b.WriteByte('=')
		b.WriteString(strings.Join(didyoumean.Correct(in, dict), ","))
		b.WriteByte(';')
	}
	return b.String()
}

var ops = []struct {
	label  string
	inputs []string
}{
	{"correct-hit", hitInputs},
	{"correct-miss", missInputs},
	{"correct-batch", batchInputs},
}

func main() {
	if os.Getenv("CHECK") != "" {
		for _, o := range ops {
			fmt.Printf("CHECK\t%s\t%s\n", o.label, correctAll(o.inputs))
		}
		return
	}
	const inner = 100
	for _, o := range ops {
		inputs := o.inputs
		bench(o.label, inner, func() { sink = correctAll(inputs) })
	}
}
