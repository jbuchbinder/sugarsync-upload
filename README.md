# SUGARSYNC UPLOADER

[![Build Status](https://secure.travis-ci.org/jbuchbinder/sugarsync-upload.png)](http://travis-ci.org/jbuchbinder/sugarsync-upload)

[![Gobuild Download](http://gobuild.io/badge/github.com/jbuchbinder/sugarsync-upload/downloads.svg)](http://gobuild.io/github.com/jbuchbinder/sugarsync-upload)

* https://github.com/jbuchbinder/sugarsync-upload
* Twitter: [@jbuchbinder](https://twitter.com/jbuchbinder)

## OVERVIEW

This client allows multiple files to be uploaded to arbitrary
Sugarsync folders.

Use '-h' with the binary to get an explanation of syntax.

## USAGE

```
Usage of sugarsync-upload:
  -action="upload": upload|list|mkdir
  -debug=false: debug mode
  -dest="": destination folder (or 'mb' for magic briefcase, 'wa' for web archive, etc)
  -folderName="": folder name (for new folder creation)
  -password="": sugarsync password
  -username="": sugarsync email/user name
```

Files to be uploaded are passed as additional parameters.

Uploading:
```
sugarsync-upload -username=user@my.net -password=pwd -action=upload -dest=wa FILE1 FILE2 FILE3
```


## COMPILING

`go build`

It requires a copy of Go 1.0.x.

## CHANGELOG

### 0.3.0

* Added MIME autodetection based on file extension. This should allow
  picture and music uploads to work properly with Sugarsync's service.

### 0.2.0

* Legacy version

