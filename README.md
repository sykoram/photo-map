# photo-map
This program generates a KML (or KMZ) file with geotagged photos.

## Table of Contents
- [Setup](#setup)
- [Usage](#usage)
  - [Flags](#flags)
  - [Modes](#modes)
  - [JSON file](#json-file)


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

## Usage

Use flag `-h` or `--help` to display the help.

The most basic command would be:
```sh
photo-map -i IMAGE_DIR -o OUTPUT_DIR
```


### Flags

`-m MODE` sets a [mode](#modes) of an image representation.

`-json JSON_FILE` defines path to a [JSON file](#json-file) with user-specified image properties.

`-timesort` orders the images by timestamp.

`-path` draws lines between the images (`-timesort` is recommended).

`-include-zero-location` includes images with [0,0] location into the path (this will not work without `-path`).

`-kmz` zips the output directory into a one KMZ file.


### Modes

`html-image`: The image is an `<img>` tag inside a HTML balloon style (it is not in the description field).

`description-image`: The image is a HTML `<img>` tag in the description.

`photo-overlay`: The image is placed above the map using PhotoOverlay.


### JSON file

See comments in [data.json](./data.json).

