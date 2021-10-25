![Paperboy](https://www.paperboy.email/images/banner.jpg)

A Fast & Modern Email Campaign Engine built in [Go][].

[Website](https://www.paperboy.email/) |
[Documentation](https://www.paperboy.email/docs/introduction/) |
[Installation Guide](https://www.paperboy.email/docs/installation/)

[![Version Badge](https://badge.fury.io/mdy/github.com%2Frykov%2Fpaperboy.svg)](https://melody.sh/github.com/rykov/paperboy)
[![Go Report Card](https://goreportcard.com/badge/github.com/rykov/paperboy)](https://goreportcard.com/report/github.com/rykov/paperboy)
[![Build Status](https://github.com/rykov/paperboy/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/paperboy/actions/workflows/tests.yml)

## Overview

Paperboy is complete email engine that helps you get the most out of your
campaigns. It allows you to craft shared templates, and then quickly author
and deliver multi-format campaigns.

Paperboy is command-line tool that consumes a [source directory][structure]
as input to render and send email campaigns via any SMTP service.  By placing
templates, lists, and content in a predefined [directory structure][structure],
Paperboy will render markup, inline styles, wrap layouts, and more to deliver
modern (yet legacy-compatible) newsletters and announcements.

**Complete documentation is available at [Paperboy Documentation][docs].**

## Installing binaries

Currently, we provide [pre-built binaries][releases] for Linux and macOS.
Paperboy is a single binary with no external dependencies.

Just run `./paperboy help` and you're [ready to go][quickstart].

## Installing from source

You can also build and install Paperboy from source. The only requirement is to
have a working installation of [Go][] 1.11+. With those in place, the following
commands will install Paperboy to `$GOPATH/bin`:

```bash
$ git clone https://github.com/rykov/paperboy.git
$ cd paperboy
$ make install
```

And please make sure `$GOPATH/bin` is in your `$PATH`.

If you receive a `go: modules disabled` error due to your project being inside
of $GOPATH, you will have to force-enable go modules support:

```bash
GO111MODULE=on make install
```

## Contributing to Paperboy

We welcome all contribution to Paperboy: documentation, bug reporst, feature
ideas, blog posts, promotion, etc

You can start here:

- [Report a bug](https://github.com/rykov/paperboy/issues/new)
- [Fork to contribute](https://github.com/rykov/paperboy/fork)
- [Improve documentation](https://github.com/rykov/paperboyDocs)
- [Logos & Design](https://github.com/rykov/paperboyDocs)

## Inspiration and credits

Paperboy aims to bring to email the ease of use and control we love from static
site generators. We particularly want to thank [Hugo][] authors for much of
the inspiration and a number of our [dependencies][].

The banner photo is by [Mathyas Kurmann](https://unsplash.com/@mathyaskurmann)

[Go]: https://golang.org/
[Hugo]: https://gohugo.io/
[quickstart]: https://www.paperboy.email/docs/quick-start/
[structure]: https://www.paperboy.email/docs/source-structure/
[releases]: https://github.com/rykov/paperboy/releases
[docs]: https://www.paperboy.email/docs/introduction/
[melody]: https://github.com/mdy/melody
