// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package bindata

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const sep = string(os.PathListSeparator)

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func readdir(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errors.New("not a directory")
	}
	fis, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}
	if len(fis) == 0 {
		return nil, errors.New("empty directory")
	}
	dirs := make([]string, 0, len(fis))
	for _, fi = range fis {
		if fi.IsDir() {
			dirs = append(dirs, fi.Name())
		}
	}
	if len(dirs) == 0 {
		return nil, errors.New("leaf directory")
	}
	return dirs, nil
}

func countdir(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	// TODO(rjeczalik): Return 0 if path recursively contains only empty
	//                  directories.
	names, err := f.Readdirnames(0)
	if err != nil {
		return 0
	}
	return len(names)
}

var vcs = map[string]struct{}{
	".bzr": {},
	".git": {},
	".hg":  {},
	".svn": {},
}

func isvcs(name string) (ok bool) {
	_, ok = vcs[name]
	return ok
}

// Glob searches paths provided by a os.PathListSeparator-separated list. It looks
// for longest matching path between $GOPATH/data and $GOPATH/src directories -
// the former is used as an input source, the latter - an output one.
// The matching part of the paths is treated as a prefix. Glob uses BFS-like
// lookup.
//
// Example
//
// For the following $GOPATH workspace:
//
//   .
//   ├── data
//   │   └── github.com
//   │       └── user
//   │           └── example
//   │               └── assets
//   │                   ├── css
//   │                   └── js
//   └── src
//       └── github.com
//           └── user
//               └── example
//
// Glob will create single Config, where the prefix would be "github.com/user/example",
// the files would get read recursively from the "assets" directory and outputted
// to the "./src/github.com/user/example/bindata.go" file.
func Glob(list string) ([]*Config, error) {
	var paths = make([]string, 0, strings.Count(list, sep)+1)
	for _, path := range strings.Split(list, sep) {
		if path != "" {
			paths = append(paths, path)
		}
	}
	type inout struct{ gopath, dir string }
	var (
		glob, data, src []string
		inouts          []inout
		dir             string
		err             error
	)
	for _, path := range paths {
		glob = append(glob, "")
		for len(glob) > 0 {
			dir, glob = glob[len(glob)-1], glob[:len(glob)-1]
			if data, err = readdir(filepath.Join(path, "data", dir)); err != nil {
				inouts = append(inouts, inout{gopath: path, dir: dir})
				continue
			}
			if src, err = readdir(filepath.Join(path, "src", dir)); err != nil {
				inouts = append(inouts, inout{gopath: path, dir: dir})
				continue
			}
			m := make(map[string]int, min(len(data), len(src)))
			for _, dir := range data {
				if !isvcs(dir) {
					m[dir]++
				}
			}
			for _, dir := range src {
				if !isvcs(dir) {
					m[dir]++
				}
			}
			for name, n := range m {
				if n > 1 {
					glob = append(glob, filepath.Join(dir, name))
				}
			}
		}
	}
	var cfgs = make([]*Config, 0, len(inouts))
	for _, inout := range inouts {
		input := filepath.Join(inout.gopath, "data", inout.dir)
		output := filepath.Join(inout.gopath, "src", inout.dir, "bindata.go")
		// TODO(rjeczalik): ignore input if max(ModTime in [input/...]) > ModTime(output)
		if countdir(input) > 0 {
			cfg := NewConfig()
			cfg.Package = filepath.Base(inout.dir)
			cfg.Prefix = filepath.Join(inout.gopath, "data", inout.dir)
			cfg.Output = output
			cfg.Input = []InputConfig{{Path: input, Recursive: true}}
			cfgs = append(cfgs, cfg)
		}
	}
	if len(cfgs) == 0 {
		return nil, errors.New("bindata: no matching $GOPATH/data directories found or no input ones provided")
	}
	return cfgs, nil
}

// GlobGenerate runs Generate concurrently over cfgs configuration list.
// It logs execution time and eventual errors via user-provided log function.
// It returns true when all executions of Generate were successful,
// false otherwise.
func GlobGenerate(cfgs []*Config, log func(*Config, time.Duration, error)) bool {
	ch, ret, ok := make(chan *Config, len(cfgs)), make(chan error), true
	for _, cfg := range cfgs {
		ch <- cfg
	}
	defer close(ch)
	for n := min(runtime.GOMAXPROCS(-1), len(cfgs)); n > 0; n-- {
		go func() {
			for c := range ch {
				begin := time.Now()
				err := Generate(c)
				log(c, time.Now().Sub(begin), err)
				ret <- err
			}
		}()
	}
	for i := 0; i < len(cfgs); i++ {
		if err := <-ret; err != nil {
			if err != nil {
				ok = false
			}
		}
	}
	return ok
}
