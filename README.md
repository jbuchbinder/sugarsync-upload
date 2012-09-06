SUGARSYNC UPLOADER
==================

* https://github.com/jbuchbinder/sugarsync-uploader
* Twitter: [@jbuchbinder](https://twitter.com/jbuchbinder)

OVERVIEW
--------

This client allows multiple files to be uploaded to arbitrary
Sugarsync folders.

Use '-h' with the binary to get an explanation of syntax.

USAGE
-----

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


COMPILING
---------

`go build`

It requires a copy of Go 1.0.x.

