chkjson
=======

[![GoDoc](https://godoc.org/github.com/twmb/chkjson?status.svg)](https://godoc.org/github.com/twmb/chkjson) [![Build Status](https://travis-ci.org/twmb/chkjson.svg?branch=master)](https://travis-ci.org/twmb/chkjson)

This repo / package provides alternatives to Go's
[encoding/json](https://golang.org/pkg/encoding/json/) package for validating
JSON and compacting it to a slice. A great appeal for this package is the
in-place `AppendCompact` function and the potentially in-place
`AppendConcatJSONP` function (and string variants).

A minor appeal is the quick and easy string and slice escaping.

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
implementations. The code, while _supremely_ ugly and repetitive,
is written for the compiler.
The implementation was guided by profiling sections of code individually with
various implementations.

This is the type of code that truly is write once, maintain rarely. It
implements validating (and compacting) to spec. The current tests are
comprehensive, the parsers have been fuzzed, and the implementation is nearly
as good as it gets without dropping into assembly. Validating passes all
required `y_` and `n_` files in [JSONTestSuite](https://github.com/nst/JSONTestSuite),
and compacting followed by unmarshalling for all of these files matches stdlib
unmarshalling.


## Why not this package?

The parsing implementation is recursive. This is much faster than a stack based
parsing implementation but does consume more memory. The implementation only
recurses when necessary and this consumes only about 2.2x the memory a simple
(not extremely memory optimized) stack based implementation would. For nearly
all cases, this extra memory consumption is not an issue, especially so since
the CPU is freed up more for actual important work.

The implementation is extremely repetitive and ugly, making it difficult to
maintain. This tradeoff was made due to ideally not ever _needing_ changes.

If you are only validating or compacting a few (countable) amount of times,
this package is overkill.

## Documentation

Full documentation can be found on [`godoc`](https://godoc.org/github.com/twmb/chkjson).

## Benchmarks

What follows is `benchstat` output for JSON files taken from [valyala/fastjson](https://github.com/valyala/fastjson)
comparing stdlib against my code.

Memory allocation savings are elided for brevity (normally 72 to 312 bytes).

```
name                  old time/op    new time/op    delta
ExtCompact/canada-4     16.8ms ± 1%     4.0ms ± 0%   -76.06%  (p=0.000 n=8+9)
ExtCompact/citm-4       13.9ms ± 1%     2.6ms ± 0%   -81.50%  (p=0.000 n=10+9)
ExtCompact/large-4       234µs ± 0%      70µs ± 0%   -70.11%  (p=0.000 n=10+9)
ExtCompact/medium-4     21.0µs ± 0%     5.2µs ± 1%   -75.11%  (p=0.000 n=8+10)
ExtCompact/small-4      1.80µs ± 0%    0.45µs ± 0%   -74.86%  (p=0.000 n=10+10)
ExtCompact/twitter-4    5.34ms ± 0%    1.43ms ± 0%   -73.24%  (p=0.000 n=9+9)
ExtValid/canada-4       10.9ms ± 0%     3.0ms ± 0%   -72.12%  (p=0.000 n=9+10)
ExtValid/citm-4         8.36ms ± 0%    2.31ms ± 0%   -72.36%  (p=0.000 n=9+10)
ExtValid/large-4         153µs ± 0%      48µs ± 0%   -68.83%  (p=0.000 n=9+10)
ExtValid/medium-4       12.4µs ± 0%     3.8µs ± 0%   -69.70%  (p=0.000 n=10+10)
ExtValid/small-4        1.09µs ± 0%    0.31µs ± 0%   -71.79%  (p=0.000 n=10+10)
ExtValid/twitter-4      3.30ms ± 0%    1.06ms ± 0%   -67.76%  (p=0.000 n=10+9)

name                  old speed      new speed      delta
ExtCompact/canada-4    134MB/s ± 1%   561MB/s ± 0%  +317.73%  (p=0.000 n=8+9)
ExtCompact/citm-4      124MB/s ± 1%   672MB/s ± 0%  +440.60%  (p=0.000 n=10+9)
ExtCompact/large-4     120MB/s ± 0%   402MB/s ± 0%  +234.53%  (p=0.000 n=10+9)
ExtCompact/medium-4    111MB/s ± 0%   445MB/s ± 1%  +301.72%  (p=0.000 n=8+10)
ExtCompact/small-4     106MB/s ± 0%   419MB/s ± 0%  +297.49%  (p=0.000 n=10+10)
ExtCompact/twitter-4   118MB/s ± 0%   442MB/s ± 0%  +273.75%  (p=0.000 n=9+9)
ExtValid/canada-4      206MB/s ± 0%   739MB/s ± 0%  +258.67%  (p=0.000 n=9+10)
ExtValid/citm-4        207MB/s ± 0%   747MB/s ± 0%  +261.84%  (p=0.000 n=9+10)
ExtValid/large-4       184MB/s ± 0%   590MB/s ± 0%  +220.77%  (p=0.000 n=9+10)
ExtValid/medium-4      187MB/s ± 0%   618MB/s ± 0%  +230.03%  (p=0.000 n=10+10)
ExtValid/small-4       175MB/s ± 0%   618MB/s ± 0%  +254.06%  (p=0.000 n=10+10)
ExtValid/twitter-4     191MB/s ± 0%   593MB/s ± 0%  +210.19%  (p=0.000 n=10+9)
```

In place compacting directly using the `Compact` function is even faster.

```
name                         old time/op    new time/op    delta
ExtCompactInplace/canada-4     16.8ms ± 0%     3.5ms ± 1%    -78.88%  (p=0.000 n=8+10)
ExtCompactInplace/citm-4       13.9ms ± 1%     0.9ms ± 1%    -93.36%  (p=0.000 n=10+10)
ExtCompactInplace/large-4       235µs ± 1%      56µs ± 1%    -76.35%  (p=0.000 n=10+10)
ExtCompactInplace/medium-4     21.1µs ± 1%     3.2µs ± 1%    -84.73%  (p=0.000 n=10+10)
ExtCompactInplace/small-4      1.85µs ± 3%    0.28µs ± 2%    -84.83%  (p=0.000 n=10+10)
ExtCompactInplace/twitter-4    5.35ms ± 0%    0.88ms ± 1%    -83.58%  (p=0.000 n=10+10)

name                         old speed      new speed      delta
ExtCompactInplace/canada-4    134MB/s ± 0%   636MB/s ± 1%   +373.47%  (p=0.000 n=8+10)
ExtCompactInplace/citm-4      124MB/s ± 1%  1872MB/s ± 1%  +1406.08%  (p=0.000 n=10+10)
ExtCompactInplace/large-4     120MB/s ± 1%   506MB/s ± 1%   +322.89%  (p=0.000 n=10+10)
ExtCompactInplace/medium-4    110MB/s ± 1%   723MB/s ± 0%   +554.90%  (p=0.000 n=10+10)
ExtCompactInplace/small-4     103MB/s ± 3%   677MB/s ± 2%   +558.42%  (p=0.000 n=10+10)
ExtCompactInplace/twitter-4   118MB/s ± 0%   719MB/s ± 1%   +508.94%  (p=0.000 n=10+10)
```
