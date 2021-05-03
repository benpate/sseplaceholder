## SSE Placeholder

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/benpate/sseplaceholder)
[![Go Report Card](https://goreportcard.com/badge/github.com/benpate/sseplaceholder?style=flat-square)](https://goreportcard.com/report/github.com/benpate/sseplaceholder)
[![Build Status](http://img.shields.io/travis/com/benpate/sseplaceholder.svg?style=flat-square)](https://travis-ci.com/benpate/sseplaceholder)
[![Codecov](https://img.shields.io/codecov/c/github/benpate/sseplaceholder.svg?style=flat-square)](https://codecov.io/gh/benpate/sseplaceholder)


SSE Placeholder is a simple service that generates Server Sent Events for your test pages to read.  It streams fake data from [jsonplaceholder](https://jsonplaceholder.typicode.com) to your website on a semi-regular schedule.

## JSON Event Streams

Streams random JSON records every second (or so) to your client.

* `/posts.json`
* `/comments.json`
* `/albums.json`
* `/photos.json`
* `/todos.json`
* `/users.json`

## HTML Event Streams

Streams random HTML fragments every second (or so) to your client.

* `/posts.html`
* `/comments.html`
* `/albums.html`
* `/photos.html`
* `/todos.html`
* `/users.html`

## HTMX Demos

Includes several other demo pages that use [HTMX](https://htmx.org) to load remote data.

---

It is inspired by [jsonplaceholder](https://jsonplaceholder.typicode.com) -- *"a free online REST API that you can use whenever you need some fake data."*

Also, thanks to [HTMX](https://htmx.org) working and testing this code was the original reason for putting this code together.
