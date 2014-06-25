// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jteeuwen/go-bindata"
)

func die(err error) {
	fmt.Fprintf(os.Stderr, "bindata: %v\n", err)
	os.Exit(1)
}

var data = string(os.PathSeparator) + "data" + string(os.PathSeparator)

func log(c *bindata.Config, d time.Duration, err error) {
	prefix := c.Input[0].Path
	if i := strings.Index(prefix, data); i != -1 {
		prefix = prefix[i+len(data):]
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail\t%s\t(%s)\t%.3fs\n\terror: %v\n",
			prefix, c.Output, d.Seconds(), err)
	} else {
		fmt.Printf("ok\t%s\t(%s)\t%.3fs\n", prefix, c.Output, d.Seconds())
	}
}

func copycfg(dst, src *bindata.Config) {
	dst.Tags = src.Tags
	dst.NoMemCopy = src.NoMemCopy
	dst.NoCompress = src.NoCompress
	dst.Debug = src.Debug
	dst.Ignore = src.Ignore
}

func main() {
	c, auto := parseArgs()
	if auto {
		cfgs, err := bindata.Glob(os.Getenv("GOPATH"))
		if err != nil {
			die(err)
		}
		for _, cfg := range cfgs {
			copycfg(cfg, c)
		}
		if !bindata.GlobTranslate(cfgs, log) {
			os.Exit(1)
		}
		return
	}
	if err := bindata.Translate(c); err != nil {
		die(err)
	}
}

// parseArgs creates a new, filled configuration instance
// by reading and parsing command line options.
//
// This function exits the program with an error, if
// any of the command line options are incorrect.
func parseArgs() (c *bindata.Config, auto bool) {
	var version bool

	c = bindata.NewConfig()

	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <input directories>\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVar(&c.Debug, "debug", c.Debug, "Do not embed the assets, but provide the embedding API. Contents will still be loaded from disk.")
	flag.StringVar(&c.Tags, "tags", c.Tags, "Optional set of build tags to include.")
	flag.StringVar(&c.Prefix, "prefix", c.Prefix, "Optional path prefix to strip off asset names.")
	flag.StringVar(&c.Package, "pkg", c.Package, "Package name to use in the generated code.")
	flag.BoolVar(&c.NoMemCopy, "nomemcopy", c.NoMemCopy, "Use a .rodata hack to get rid of unnecessary memcopies. Refer to the documentation to see what implications this carries.")
	flag.BoolVar(&c.NoCompress, "nocompress", c.NoCompress, "Assets will *not* be GZIP compressed when this flag is specified.")
	flag.StringVar(&c.Output, "o", c.Output, "Optional name of the output file to be generated.")
	flag.BoolVar(&version, "version", false, "Displays version information.")

	ignore := make([]string, 0)
	flag.Var((*AppendSliceValue)(&ignore), "ignore", "Regex pattern to ignore")

	flag.Parse()

	for _, pattern := range ignore {
		c.Ignore = append(c.Ignore, regexp.MustCompile(pattern))
	}

	if version {
		fmt.Printf("%s\n", Version())
		os.Exit(0)
	}

	// No input directories provided, assuming automatic mode.
	if flag.NArg() == 0 {
		auto = true
		return
	}

	// Create input configurations.
	c.Input = make([]bindata.InputConfig, flag.NArg())
	for i := range c.Input {
		c.Input[i] = parseInput(flag.Arg(i))
	}

	return
}

// parseRecursive determines whether the given path has a recrusive indicator and
// returns a new path with the recursive indicator chopped off if it does.
//
//  ex:
//      /path/to/foo/...    -> (/path/to/foo, true)
//      /path/to/bar        -> (/path/to/bar, false)
func parseInput(path string) bindata.InputConfig {
	if strings.HasSuffix(path, "/...") {
		return bindata.InputConfig{
			Path:      filepath.Clean(path[:len(path)-4]),
			Recursive: true,
		}
	} else {
		return bindata.InputConfig{
			Path:      filepath.Clean(path),
			Recursive: false,
		}
	}

}
