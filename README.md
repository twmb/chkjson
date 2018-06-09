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
ExtValid/large-4         156µs ± 0%       63µs ± 0%   -59.54%  (p=0.000 n=8+10)
ExtValid/canada-4       9.86ms ± 0%     5.00ms ± 0%   -49.31%  (p=0.000 n=9+10)
ExtValid/citm-4         8.29ms ± 0%     3.58ms ± 0%   -56.76%  (p=0.000 n=10+9)
ExtValid/twitter-4      3.32ms ± 0%     1.50ms ± 1%   -54.82%  (p=0.000 n=10+9)
ExtValid/small-4        1.10µs ± 0%     0.42µs ± 0%   -62.16%  (p=0.000 n=10+10)
ExtValid/medium-4       12.6µs ± 0%      5.2µs ± 1%   -58.22%  (p=0.000 n=10+9)
ExtCompact/small-4      1.92µs ± 0%     0.51µs ± 0%   -73.39%  (p=0.000 n=10+9)
ExtCompact/medium-4     22.5µs ± 0%      6.7µs ± 0%   -70.44%  (p=0.000 n=10+9)
ExtCompact/large-4       247µs ± 0%       86µs ± 0%   -65.36%  (p=0.000 n=10+10)
ExtCompact/canada-4     18.3ms ± 0%      6.4ms ± 0%   -64.85%  (p=0.000 n=10+10)
ExtCompact/citm-4       14.4ms ± 0%      3.9ms ± 0%   -72.85%  (p=0.000 n=10+10)
ExtCompact/twitter-4    5.72ms ± 0%     1.88ms ± 1%   -67.12%  (p=0.000 n=9+9)

name                  old speed      new speed       delta
ExtValid/large-4       180MB/s ± 0%    446MB/s ± 0%  +147.18%  (p=0.000 n=8+10)
ExtValid/canada-4      228MB/s ± 0%    450MB/s ± 0%   +97.27%  (p=0.000 n=9+10)
ExtValid/citm-4        208MB/s ± 0%    482MB/s ± 0%  +131.26%  (p=0.000 n=10+9)
ExtValid/twitter-4     190MB/s ± 0%    421MB/s ± 1%  +121.36%  (p=0.000 n=10+9)
ExtValid/small-4       172MB/s ± 0%    455MB/s ± 0%  +164.13%  (p=0.000 n=10+10)
ExtValid/medium-4      185MB/s ± 0%    444MB/s ± 1%  +139.33%  (p=0.000 n=10+9)
ExtCompact/small-4    99.0MB/s ± 0%  371.9MB/s ± 0%  +275.57%  (p=0.000 n=10+9)
ExtCompact/medium-4    103MB/s ± 0%    350MB/s ± 0%  +238.29%  (p=0.000 n=10+9)
ExtCompact/large-4     114MB/s ± 0%    328MB/s ± 0%  +188.72%  (p=0.000 n=10+10)
ExtCompact/canada-4    123MB/s ± 0%    351MB/s ± 0%  +184.46%  (p=0.000 n=10+10)
ExtCompact/citm-4      120MB/s ± 0%    443MB/s ± 0%  +268.38%  (p=0.000 n=10+10)
ExtCompact/twitter-4   110MB/s ± 0%    336MB/s ± 1%  +204.14%  (p=0.000 n=9+9)

name                  old alloc/op   new alloc/op    delta
ExtValid/large-4          184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtValid/canada-4         184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtValid/citm-4           184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtValid/twitter-4        312B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtValid/small-4         72.0B ± 0%       0.0B       -100.00%  (p=0.000 n=10+10)
ExtValid/medium-4         184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/small-4       72.0B ± 0%       0.0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/medium-4       184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/large-4        184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/canada-4       184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/citm-4         184B ± 0%         0B       -100.00%  (p=0.000 n=10+10)
ExtCompact/twitter-4      312B ± 0%         0B       -100.00%  (p=0.000 n=10+10)

name                  old allocs/op  new allocs/op   delta
ExtValid/large-4          5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/canada-4         5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/citm-4           5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/twitter-4        6.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/small-4          2.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtValid/medium-4         5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/small-4        2.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/medium-4       5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/large-4        5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/canada-4       5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/citm-4         5.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
ExtCompact/twitter-4      6.00 ± 0%       0.00       -100.00%  (p=0.000 n=10+10)
```
