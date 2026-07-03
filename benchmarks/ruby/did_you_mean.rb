# frozen_string_literal: true
# Copyright (c) the go-ruby-did-you-mean/did-you-mean authors
# SPDX-License-Identifier: BSD-3-Clause
#
# Reference did_you_mean workload, mirroring benchmarks/go/main.go op-for-op over
# an identical, deterministic corpus. It drives the runtime's own stdlib
# DidYouMean::SpellChecker#correct — the exact public primitive the pure-Go
# Correct ports — against a fixed dictionary of Ruby-ish identifier names and
# three fixed input batches. No interpreter-tied did_you_mean machinery (the
# NameError/NoMethodError hooks, per-error checkers, formatter) is touched: only
# the deterministic, reproducible matcher is measured.
#
# Run normally it reports ns/op per op through the shared harness; run with
# CHECK=1 it prints one "CHECK\t<label>\t<value>" line per op so the Go output
# can be proven byte-identical to MRI (the oracle) before any timing is trusted.
require "did_you_mean"
require_relative "_harness"

# Byte-for-byte the same dictionary the Go driver builds; order is significant.
DICT = %w[
  map flat_map collect select filter_map reject
  detect find find_all each each_with_index each_with_object
  inject reduce group_by partition sort sort_by
  min max min_by max_by sum count
  tally zip take take_while drop drop_while
  first last length size include index
].freeze

# Near-typos that each land on a suggestion (full pipeline).
HIT_INPUTS = %w[
  collekt selekt rejekt detct injekt
  reduse sord lenght grup_by parttion
  flat_mp cont firts include idnex sizze
].freeze

# Inputs far from every candidate: a pure Jaro–Winkler sweep, all rejected.
MISS_INPUTS = %w[
  xyzzy qwerty zzzzzz asdfgh plugh frobnicate
  wibble quux 0xdeadbeef supercalifragilistic hjkl vmware
  kubernetes photosynthesis abcdefg mmmmmm
].freeze

# Realistic mixed workload: hits and misses interleaved.
BATCH_INPUTS = begin
  out = []
  HIT_INPUTS.each_with_index do |h, i|
    out << h
    out << MISS_INPUTS[i] if i < MISS_INPUTS.length
  end
  out.freeze
end

# One SpellChecker over the fixed dictionary, reused across ops exactly as the
# Go driver reuses the fixed dict slice.
CHECKER = DidYouMean::SpellChecker.new(dictionary: DICT)

# correct_all runs #correct over every input and joins every ranked suggestion
# into a total, order-sensitive fingerprint of the whole batch's output.
def correct_all(inputs)
  s = +""
  inputs.each do |input|
    s << input << "=" << CHECKER.correct(input).join(",") << ";"
  end
  s
end

OPS = [
  ["correct-hit",   HIT_INPUTS],
  ["correct-miss",  MISS_INPUTS],
  ["correct-batch", BATCH_INPUTS],
].freeze

if ENV["CHECK"] && !ENV["CHECK"].empty?
  OPS.each { |label, inputs| printf("CHECK\t%s\t%s\n", label, correct_all(inputs)) }
else
  INNER = 100
  OPS.each { |label, inputs| bench(label, INNER) { correct_all(inputs) } }
end
