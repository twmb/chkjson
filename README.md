chkjson
=======

[![GoDoc](https://godoc.org/github.com/twmb/chkjson?status.svg)](https://godoc.org/github.com/twmb/chkjson) [![Build Status](https://travis-ci.org/twmb/chkjson.svg?branch=master)](https://travis-ci.org/twmb/chkjson)

This repo / package provides alternatives to Go's
[encoding/json](https://golang.org/pkg/encoding/json/) package for validating
JSON and appending it to a slice. The primary appeal for this package is its
in-place `AppendCompact` function and its potentially in-place
`AppendConcatJSONP` function.

A side benefit of this package compared to encoding/json's is that this
package's functions avoid allocating and are slightly faster. The stdlib's
`Valid` and `Compact` functions allocate at least once for every call (but they
are small allocations and generally not worth worrying about).

Full documentation can be found on [`godoc`](https://godoc.org/github.com/twmb/chkjson).
