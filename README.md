chkjson
=======

[![GoDoc](https://godoc.org/github.com/twmb/chkjson?status.svg)](https://godoc.org/github.com/twmb/chkjson) [![Build Status](https://travis-ci.org/twmb/chkjson.svg?branch=master)](https://travis-ci.org/twmb/chkjson)

This repo / package provides alternatives to Go's
[encoding/json](https://golang.org/pkg/encoding/json/) package for validating
JSON and appending it to a slice. The primary appeal for this package is its
in-place `AppendCompact` function, and its potentially in-place
`AppendConcatJSONP` function.

A side benefit of this package compared to encoding/json's is that this
package's functions avoid allocating and are very fast (the stdlib's `Valid`
and `Compact` functions allocate at least once for every call [but they are
small allocations and generally not worth worrying about]).

Full documentation can be found on [`godoc`](https://godoc.org/github.com/twmb/chkjson).

## Benchmarks

What follows is `benchstat` output for JSON files taken from [valyala/fastjson](https://github.com/valyala/fastjson)
comparing stdlib against my code.

```
name                  old time/op    new time/op     delta
ExtCompact/canada-4     18.3ms ± 0%      6.4ms ± 0%   -64.96%  (p=0.000 n=9+9)
ExtCompact/citm-4       14.4ms ± 0%      3.9ms ± 0%   -72.50%  (p=0.000 n=8+9)
ExtCompact/large-4       248µs ± 0%       88µs ±10%   -64.47%  (p=0.000 n=10+10)
ExtCompact/medium-4     22.6µs ± 2%      6.6µs ± 0%   -70.85%  (p=0.000 n=10+8)
ExtCompact/small-4      1.92µs ± 0%     0.50µs ± 0%   -73.74%  (p=0.000 n=8+9)
ExtCompact/twitter-4    5.72ms ± 1%     1.88ms ± 0%   -67.06%  (p=0.000 n=8+8)
ExtValid/canada-4       10.5ms ±10%      5.0ms ± 0%   -52.63%  (p=0.000 n=10+8)
ExtValid/citm-4         8.50ms ± 7%     3.75ms ± 0%   -55.92%  (p=0.000 n=10+9)
ExtValid/large-4         156µs ± 0%       63µs ±15%   -59.84%  (p=0.000 n=8+9)
ExtValid/medium-4       12.6µs ± 1%      5.1µs ± 0%   -59.47%  (p=0.000 n=9+9)
ExtValid/small-4        1.10µs ± 0%     0.41µs ±10%   -62.58%  (p=0.000 n=10+9)
ExtValid/twitter-4      3.32ms ± 0%     1.45ms ± 1%   -56.28%  (p=0.000 n=9+9)

name                  old speed      new speed       delta
ExtCompact/canada-4    123MB/s ± 0%    352MB/s ± 0%  +185.39%  (p=0.000 n=9+9)
ExtCompact/citm-4      120MB/s ± 0%    437MB/s ± 0%  +263.63%  (p=0.000 n=8+9)
ExtCompact/large-4     114MB/s ± 0%    320MB/s ± 9%  +181.93%  (p=0.000 n=10+10)
ExtCompact/medium-4    103MB/s ± 2%    353MB/s ± 0%  +243.08%  (p=0.000 n=10+8)
ExtCompact/small-4    99.0MB/s ± 0%  376.7MB/s ± 0%  +280.59%  (p=0.000 n=8+9)
ExtCompact/twitter-4   110MB/s ± 1%    335MB/s ± 0%  +203.58%  (p=0.000 n=8+8)
ExtValid/canada-4      215MB/s ±10%    453MB/s ± 0%  +110.16%  (p=0.000 n=10+8)
ExtValid/citm-4        204MB/s ± 6%    461MB/s ± 0%  +126.55%  (p=0.000 n=10+9)
ExtValid/large-4       180MB/s ± 0%    450MB/s ±13%  +149.75%  (p=0.000 n=8+9)
ExtValid/medium-4      185MB/s ± 1%    455MB/s ± 0%  +146.70%  (p=0.000 n=9+9)
ExtValid/small-4       173MB/s ± 0%    462MB/s ± 9%  +167.33%  (p=0.000 n=10+9)
ExtValid/twitter-4     190MB/s ± 0%    435MB/s ± 1%  +128.73%  (p=0.000 n=9+9)

name                  old alloc/op   new alloc/op    delta
ExtCompact/canada-4       184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/citm-4         184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/large-4        184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/medium-4       184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/small-4       72.0B ± 0%       0.0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/twitter-4      312B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtValid/canada-4         184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtValid/citm-4           184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtValid/large-4          184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtValid/medium-4         184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtValid/small-4         72.0B ± 0%       0.0B       -100.00%  (p=0.000 n=10+10)
ExtValid/twitter-4        312B ± 0%         0B       -100.00%  (p=0.000 n=10+10)

name                  old allocs/op  new allocs/op   delta
ExtCompact/canada-4       5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/citm-4         5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/large-4        5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/medium-4       5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/small-4        2.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/twitter-4      6.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/canada-4         5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/citm-4           5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/large-4          5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/medium-4         5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/small-4          2.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/twitter-4        6.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
```
