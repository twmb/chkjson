chkjson
=======

[![GoDoc](https://godoc.org/github.com/twmb/chkjson?status.svg)](https://godoc.org/github.com/twmb/chkjson) [![Build Status](https://travis-ci.org/twmb/chkjson.svg?branch=master)](https://travis-ci.org/twmb/chkjson)

This repo / package provides alternatives to Go's
[encoding/json](https://golang.org/pkg/encoding/json/) package for validating
JSON and compacting it to a slice. A great appeal for this package is the
in-place `AppendCompact` function and the potentially in-place
`AppendConcatJSONP` function (and string variants).

## Why would I need this?

This package was designed for streaming untrusted JSON input to endpoints that
specifically take JSON. Because the input is untrusted, it should be validated
before being sent off. This is even more important if the JSON is being batched
before sending—you do not want to mix good JSON with bad and have the whole
batch rejected.

## Why this package?

If you are streaming _lots_ of JSON _all of the time_, the CPU and memory
savings this package provides add up.

This package outperforms any other JSON validating or compacting
implementations. The code, while _supremely_ ugly, is written for the compiler.
The implementation was guided by profiling sections of code individually with
various implementations.

This is the type of code that truly is write once, maintain rarely. It
implements validating (and compacting) to spec. The current tests are
comprehensive, the parsers have been fuzzed, and the implementation is nearly
as good as it gets without dropping into assembly.

## Why not this package?

The parsing implementation is recursive. This is much faster than a stack based
parsing implementation but does consume more memory. The implementation only
recurses when necessary and this consumes only about 2.2x the memory a simple
(not extremely memory optimized) stack based implementation would. For nearly
all cases, this extra memory consumption is not an issue, especially so since
the CPU is freed up more for actual important work.

If you are only validating or compacting a few (countable) amount of times,
this package is overkill.

## Documentation

Full documentation can be found on [`godoc`](https://godoc.org/github.com/twmb/chkjson).

## Benchmarks

What follows is `benchstat` output for JSON files taken from [valyala/fastjson](https://github.com/valyala/fastjson)
comparing stdlib against my code.

```
name                  old time/op    new time/op    delta
ExtCompact/canada-4     17.5ms ± 0%     3.9ms ± 0%   -77.81%  (p=0.000 n=9+8)
ExtCompact/citm-4       14.0ms ± 0%     2.5ms ± 0%   -82.10%  (p=0.000 n=9+9)
ExtCompact/large-4       243µs ± 0%      73µs ± 0%   -69.98%  (p=0.000 n=8+9)
ExtCompact/medium-4     22.3µs ± 0%     5.4µs ± 0%   -75.58%  (p=0.000 n=9+9)
ExtCompact/small-4      1.88µs ± 0%    0.45µs ± 0%   -75.89%  (p=0.000 n=10+9)
ExtCompact/twitter-4    5.69ms ± 0%    1.52ms ± 0%   -73.20%  (p=0.000 n=9+9)
ExtValid/canada-4       9.91ms ± 0%    3.34ms ± 0%   -66.32%  (p=0.000 n=9+10)
ExtValid/citm-4         8.38ms ± 1%    2.31ms ± 1%   -72.38%  (p=0.000 n=10+10)
ExtValid/large-4         156µs ± 0%      47µs ± 1%   -70.25%  (p=0.000 n=8+10)
ExtValid/medium-4       12.6µs ± 0%     3.5µs ± 1%   -72.03%  (p=0.000 n=9+10)
ExtValid/small-4        1.13µs ± 0%    0.30µs ± 4%   -73.35%  (p=0.000 n=9+10)
ExtValid/twitter-4      3.33ms ± 0%    1.07ms ± 0%   -67.92%  (p=0.000 n=10+9)

name                  old speed      new speed      delta
ExtCompact/canada-4    129MB/s ± 0%   579MB/s ± 0%  +350.64%  (p=0.000 n=9+8)
ExtCompact/citm-4      123MB/s ± 0%   689MB/s ± 0%  +458.71%  (p=0.000 n=9+9)
ExtCompact/large-4     116MB/s ± 0%   385MB/s ± 0%  +233.01%  (p=0.000 n=9+9)
ExtCompact/medium-4    105MB/s ± 0%   428MB/s ± 0%  +309.41%  (p=0.000 n=9+9)
ExtCompact/small-4     101MB/s ± 0%   418MB/s ± 0%  +314.42%  (p=0.000 n=10+9)
ExtCompact/twitter-4   111MB/s ± 0%   415MB/s ± 0%  +273.21%  (p=0.000 n=9+9)
ExtValid/canada-4      227MB/s ± 0%   675MB/s ± 0%  +196.95%  (p=0.000 n=9+10)
ExtValid/citm-4        206MB/s ± 1%   746MB/s ± 1%  +262.01%  (p=0.000 n=10+10)
ExtValid/large-4       180MB/s ± 0%   605MB/s ± 1%  +236.13%  (p=0.000 n=8+10)
ExtValid/medium-4      184MB/s ± 0%   658MB/s ± 1%  +257.53%  (p=0.000 n=9+10)
ExtValid/small-4       168MB/s ± 0%   630MB/s ± 4%  +275.01%  (p=0.000 n=9+10)
ExtValid/twitter-4     190MB/s ± 0%   591MB/s ± 0%  +211.75%  (p=0.000 n=10+9)

name                  old alloc/op   new alloc/op   delta
ExtCompact/canada-4       184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/citm-4         184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/large-4        184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/medium-4       184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/small-4       72.0B ± 0%      0.0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/twitter-4      312B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtValid/canada-4         184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtValid/citm-4           184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtValid/large-4          184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtValid/medium-4         184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtValid/small-4         72.0B ± 0%      0.0B       -100.00%  (p=0.000 n=10+10)
ExtValid/twitter-4        312B ± 0%        0B       -100.00%  (p=0.000 n=10+10)

name                  old allocs/op  new allocs/op  delta
ExtCompact/canada-4       5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/citm-4         5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/large-4        5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/medium-4       5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/small-4        2.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/twitter-4      6.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/canada-4         5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/citm-4           5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/large-4          5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/medium-4         5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/small-4          2.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/twitter-4        6.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
```
