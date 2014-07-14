bindata [![GoDoc](https://godoc.org/github.com/rjeczalik/bindata?status.png)](https://godoc.org/github.com/rjeczalik/bindata) [![Build Status](https://travis-ci.org/rjeczalik/bindata.png?branch=master)](https://travis-ci.org/rjeczalik/bindata)
=======

*`bindata` is a fork of [jteeuwen/go-bindata](https://github.com/jteeuwen/go-bindata) by [jteeuwen](https://github.com/jteeuwen) branched off at [d3feb9534c](https://github.com/rjeczalik/bindata/commit/d3feb9534ca8703000a19f08ffae766d2958d7d6) with changes not meant to be pushed to the upstream repository.*

This package converts any file into managable Go source code. Useful for
embedding binary data into a Go program. The file data is optionally gzip
compressed before being converted to a raw byte slice.

It comes with a command line tool in the `bindata` sub directory.
This tool offers a set of command line options, used to customize the
output being generated.


*Installation*

```bash
~ $ go get -u github.com/rjeczalik/bindata
```

*Documentation*

[godoc.org/github.com/rjeczalik/bindata](http://godoc.org/github.com/rjeczalik/bindata)

## cmd/bindata [![GoDoc](https://godoc.org/github.com/rjeczalik/bindata/cmd/bindata?status.png)](https://godoc.org/github.com/rjeczalik/bindata/cmd/bindata)

*Installation*

To install the library and command line program, use the following:

```bash
~ $ go get -u github.com/rjeczalik/bindata
~ $ go install github.com/rjeczalik/bindata/cmd/bindata
```

*Documentation*

[godoc.org/github.com/rjeczalik/bindata/cmd/bindata](http://godoc.org/github.com/rjeczalik/bindata/cmd/bindata)


### Automagic conversion within `$GOPATH` workspace

When no input files nor directories are provided via command line flags,
bindata reads `$GOPATH` workspaces and attempts to convert all files it
finds recursively in `$GOPATH/data`, generating a bindata.go file in a matching
`$GOPATH/src` directory. The match is always the longest path diff between
`$GOPATH/data` and `$GOPATH/src`, in order to avoid having assets which content
overlaps. For example, running:

```bash
  ~ $ GOPATH=/home/user bindata
```

over the following $GOPATH workspace:

```bash
  /home/user
  ├── data
  │   ├── bitbucket.org
  │   │   └── user
  │   │       └── hidden
  │   │           └── subpackage
  │   │               ├── aws.txt
  │   │               └── pass.txt
  │   └── github.com
  │       └── user
  │           └── example
  │               └── assets
  │                   ├── css
  │                   │   └── default.css
  │                   └── js
  │                       ├── app.js
  │                       └── link.js
  └── src
      ├── bitbucket.org
      │   └── user
      │       └── hidden
      │           ├── hidden.go
      │           └── subpackage
      │               └── subpackage.go
      └── github.com
          └── user
              └── example
                  └── example.go
```

will create two asset files under the following paths:

```bash
  ~ $ GOPATH=/home/user bindata
  ok      bitbucket.org/user/hidden/subpackage    (/home/user/src/bitbucket.org/user/hidden/subpackage/bindata.go)       0.001s
  ok      github.com/user/example (/home/user/src/github.com/user/example/bindata.go)    0.002s
```

Running bindata in this mode will ignore any values passed by `-o`, `-pkg` and
`-prefix` flags.
