# README.md

# Status

[![GitHub release](https://img.shields.io/github/release/keltia/archive.svg)](https://github.com/keltia/archive/releases)
[![GitHub issues](https://img.shields.io/github/issues/keltia/archive.svg)](https://github.com/keltia/archive/issues)
[![Go Version](https://img.shields.io/badge/go-1.10-blue.svg)](https://golang.org/dl/)
[![Build Status](https://travis-ci.org/keltia/archive.svg?branch=master)](https://travis-ci.org/keltia/archive)
[![GoDoc](http://godoc.org/github.com/keltia/archive?status.svg)](http://godoc.org/github.com/keltia/archive)
[![SemVer](http://img.shields.io/SemVer/2.0.0.png)](https://semver.org/spec/v2.0.0.html)
[![License](https://img.shields.io/pypi/l/Django.svg)](https://opensource.org/licenses/BSD-2-Clause)
[![Go Report Card](https://goreportcard.com/badge/github.com/keltia/archive)](https://goreportcard.com/report/github.com/keltia/archive)

# Installation

As with many Go utilities, a simple

    go get github.com/keltia/archive

is enough to fetch, build and install.  Most of the time you will not need to do it explicitely, other applications will fetch it automatically as a dependency.

# Dependencies

* Go >= 1.10

Only standard Go modules are used.  I use Semantic Versioning for all my modules.

# Usage

The module currently supports the following "archives":

- plain text
- gzip files (one file per stream, only first stream)
- zip files
- GPG files (either .asc or .gpg)
- Tar files

SYNOPSIS
``` go
    a, err := archive.New("foo.txt")
    content, err := a.Extract(".txt")
    
    a, err := archive.New("bar.zip")
    content, err := a.Extract(".txt")       // extract the first .txt file
    
    a, err := archive.New("baz.txt.gz")
    content, err := a.Extract(".txt")       // extracts baz.txt
    
    // Gpg is a bit special
    a, err := archive.New("xyz.zip.asc")
    unencrypted, err := a.Extract(".zip")   // decrypt into variable
                                            // if you want to use archive.New() there too
                                            // you need to save into a temp file.
    var buf bytes.Buffer
    
    fmt.Fprintf(&buf, "%s", unencrypted)                                        
    fh, err := os.Create("xyz.zip")
    _, err := io.Copy(fh, &buf)
    fh.Close()
    
    a1, err := archive.New("xyz.zip")
    ...
    You can have more verbose output and debug by using these functions:
    
    archive.SetVerbose()
    archive.SetDebug()      // implies verbose
    ...
    archive.Reset()         // both flags are cleared
```

# Limitations

I wrote this both to simplify and my own code in `dmarc-cat` (that's also how `sandbox` got created) and to play with interfaces.  It is currently only trying to extract one file at a time matching the extension provided.  It will probably evolve into a more general code later.

# Tests

I'm trying to get to 100% coverage but some error cases are more difficult to create.

## License

This is released under the BSD 2-Clause license.  See `LICENSE.md`.

# Contributing

This project is an open Open Source project, please read `CONTRIBUTING.md`.

# Feedback

We welcome pull requests, bug fixes and issue reports.

Before proposing a large change, first please discuss your change by raising an issue.

I use Git Flow for this package so please use something similar or the usual github workflow.

1. Fork it ( https://github.com/keltia/archive/fork )
2. Checkout the develop branch (`git checkout develop`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Commit your changes (`git commit -am 'Add some feature'`)
5. Push to the branch (`git push origin my-new-feature`)
6. Create a new Pull Request
