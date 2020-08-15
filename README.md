# photo-map

An image gallery placed on a map!

## Table of Contents
- [Setup](#setup)
- [Usage](#usage)
  - [Flags](#flags)
  - [Modes](#modes)
  - [JSON file](#json-file)


## Setup

[Go](https://golang.org/) and [Git](https://git-scm.com/) have to be installed.

Clone or download this repo, preferably into `$GOPATH/src/github.com/sykoram/photo-map`.

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

Using the JSON file, you can specify some information about the images. This will overwrite information extracted from the EXIF.

Example:
```json
{"items": [
  {
    "file": "path/to/image.jpg",
    "external": "https://example.com/path/to/image.jpg",
    "dateTime": "2006:01:02 15:04:05",
    "timeZone": "UTC",
    "latitude": 50.09,
    "longitude": 14.4
  },
  {
    "file": "image2.jpg",
    "latitude": 50.087,
    "longitude": 14.42
  },
  {
    "file": "path/to/image3.jpg",
    "external": "https://example.com/path/to/image3.jpg"
  }
]}
```

The `items` key inside the main object is required, and its value is an array of objects.

Each of these objects contains information about an image:

`file` specifies the file (image). The path should be relative to the input directory (containing images), so it might be a good idea to put the JSON file also in this directory.

`external` specifies the absolute path to the corresponding image that is somewhere else (eg. on a website) and is not included in the KMZ file.

`dateTime` sets the date and time using the EXIF format: `"2006:01:02 15:04:05"`. Any trailing spaces or null characters are trimmed.

`timeZone` sets the time zone. It has to be either a valid [tz database name](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) or `"Local"`. If the `dateTime` is not specified in JSON for an image, the `dateTime` from EXIF will be recalculated using the difference between the EXIF time zone (if exists) and the specified zone. This can be used to correct the time and date.

`latitude` and `longitude` define the GPS coordinations of the image. Positive for north and east, and negative for south and west.

If a field is left out, the data from EXIF will not be overwritten. Unknown keys are ignored.


