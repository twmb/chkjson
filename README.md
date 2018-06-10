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
name                  old time/op    new time/op    delta
ExtValid/citm-4         8.31ms ± 0%    3.69ms ± 1%   -55.62%  (p=0.000 n=9+9)
ExtValid/twitter-4      3.32ms ± 0%    1.65ms ± 2%   -50.39%  (p=0.000 n=9+10)
ExtValid/small-4        1.12µs ± 1%    0.42µs ± 0%   -62.89%  (p=0.000 n=10+10)
ExtValid/medium-4       12.7µs ± 3%     5.4µs ± 1%   -57.80%  (p=0.000 n=10+10)
ExtValid/large-4         156µs ± 0%      68µs ± 1%   -56.46%  (p=0.000 n=9+10)
ExtValid/canada-4       9.91ms ± 1%    5.14ms ± 0%   -48.20%  (p=0.000 n=9+10)
ExtCompact/small-4      1.88µs ± 0%    0.55µs ± 0%   -70.84%  (p=0.000 n=9+9)
ExtCompact/medium-4     22.2µs ± 0%     7.0µs ± 0%   -68.49%  (p=0.000 n=8+8)
ExtCompact/large-4       243µs ± 0%      96µs ± 4%   -60.61%  (p=0.000 n=8+9)
ExtCompact/canada-4     17.4ms ± 0%     6.7ms ± 0%   -61.60%  (p=0.000 n=10+10)
ExtCompact/citm-4       13.9ms ± 0%     4.1ms ± 0%   -70.81%  (p=0.000 n=9+8)
ExtCompact/twitter-4    5.68ms ± 0%    1.98ms ± 0%   -65.05%  (p=0.000 n=9+8)

name                  old speed      new speed      delta
ExtValid/citm-4        208MB/s ± 0%   468MB/s ± 1%  +125.32%  (p=0.000 n=9+9)
ExtValid/twitter-4     190MB/s ± 0%   383MB/s ± 2%  +101.62%  (p=0.000 n=9+10)
ExtValid/small-4       169MB/s ± 1%   456MB/s ± 0%  +169.18%  (p=0.000 n=10+10)
ExtValid/medium-4      183MB/s ± 3%   434MB/s ± 0%  +137.04%  (p=0.000 n=10+9)
ExtValid/large-4       180MB/s ± 0%   414MB/s ± 1%  +129.66%  (p=0.000 n=9+10)
ExtValid/canada-4      227MB/s ± 1%   438MB/s ± 0%   +93.03%  (p=0.000 n=9+10)
ExtCompact/small-4     101MB/s ± 0%   347MB/s ± 0%  +242.83%  (p=0.000 n=9+9)
ExtCompact/medium-4    105MB/s ± 0%   333MB/s ± 0%  +217.30%  (p=0.000 n=8+8)
ExtCompact/large-4     116MB/s ± 0%   294MB/s ± 4%  +153.92%  (p=0.000 n=8+9)
ExtCompact/canada-4    129MB/s ± 0%   336MB/s ± 0%  +160.38%  (p=0.000 n=10+10)
ExtCompact/citm-4      124MB/s ± 0%   425MB/s ± 0%  +242.63%  (p=0.000 n=9+8)
ExtCompact/twitter-4   111MB/s ± 0%   318MB/s ± 0%  +186.10%  (p=0.000 n=9+8)

name                  old alloc/op   new alloc/op   delta
ExtValid/citm-4           184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtValid/twitter-4        312B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtValid/small-4         72.0B ± 0%      0.0B       -100.00%  (p=0.000 n=10+10)
ExtValid/medium-4         184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtValid/large-4          184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtValid/canada-4         184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/small-4       72.0B ± 0%      0.0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/medium-4       184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/large-4        184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/canada-4       184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/citm-4         184B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/twitter-4      312B ± 0%        0B       -100.00%  (p=0.000 n=10+10)

name                  old allocs/op  new allocs/op  delta
ExtValid/citm-4           5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/twitter-4        6.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/small-4          2.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/medium-4         5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/large-4          5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/canada-4         5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/small-4        2.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/medium-4       5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/large-4        5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/canada-4       5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/citm-4         5.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/twitter-4      6.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
```
