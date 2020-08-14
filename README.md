# photo-map
This program generates a KML (or KMZ) file with geotagged photos.

## Table of Contents
- [Setup](#setup)

## Setup

[Go](https://golang.org/) and [Git](https://git-scm.com/) have to be installed.

Clone or download this repo, preferably into `$GOPATH/scr/github.com/sykoram/photo-map`.

Download and install all dependencies:
```sh
go get ./...
```

And install the photo-map:
```sh
go install
```

This creates an executable inside `$GOBIN` (usually `$GOPATH/bin`), and it photo-map should work now. You can try to run:

```sh
photo-map -h
```

