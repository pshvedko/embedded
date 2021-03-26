package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type dir struct {
	root  string
	name  string
	files []string
	max   int
}

func (d *dir) generate(path string, info os.FileInfo, err0 error) (err error) {
	if err0 != nil || info.IsDir() {
		return
	}
	var in *os.File
	in, err = os.Open(path)
	if err != nil {
		return
	}
	defer in.Close()
	n := len(d.files)
	d.files = append(d.files, strings.TrimPrefix(path, d.root))
	m := len(d.files[n])
	if d.max < m {
		d.max = m
	}
	path = fmt.Sprintf("%d.go", n)
	path = filepath.Join(d.name, path)
	var out *os.File
	out, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer out.Close()
	_, err = fmt.Fprintf(out, "package %s\n\nvar b%d = [...]byte{\n", filepath.Base(d.name), n)
	if err != nil {
		return
	}
	b := [16]byte{}
	for {
		n, err = in.Read(b[:])
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		_, err = fmt.Fprintf(out, "\t")
		if err != nil {
			return
		}
		for _, x := range b[:n] {
			_, err = fmt.Fprintf(out, "0x%x, ", x)
			if err != nil {
				return
			}
		}
		_, err = fmt.Fprintf(out, "\n")
		if err != nil {
			return
		}
	}
	_, err = fmt.Fprintf(out, "}\n")
	return
}

const (
	header = `package %s

import (
	"net/http"

	"github.com/pshvedko/embedded"
)

func Dir() http.FileSystem {
	return embedded.New(
		map[string][]byte{
`
	mapper = `			%q: %*sb%d[:],
`
	footer = `		}, %v)
}
`
)

func (d *dir) gen() (err error) {
	err = os.RemoveAll(d.name)
	if err != nil {
		return
	}
	err = os.MkdirAll(d.name, 0755)
	if err != nil {
		return
	}
	err = filepath.Walk(d.root, d.generate)
	if err != nil {
		return
	}
	var out *os.File
	path := filepath.Join(d.name, "main.go")
	out, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer out.Close()
	_, err = fmt.Fprintf(out, header, filepath.Base(d.name))
	if err != nil {
		return
	}
	for i, f := range d.files {
		_, err = fmt.Fprintf(out, mapper, f, d.max-len(f), "", i)
		if err != nil {
			return
		}
	}
	_, err = fmt.Fprintf(out, footer, time.Now().Unix())
	return
}

var source string

func init() {
	flag.StringVar(&source, "source", ".", "path to source dir")
}

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage:\n\t%s [options] package\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	c := dir{root: filepath.Clean(source), name: filepath.Clean(flag.Arg(0))}
	err := c.gen()
	if err != nil {
		log.Fatal(err)
	}
}
