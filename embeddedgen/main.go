package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type generator struct {
	root  string
	name  string
	files []string
	max   int
}

func (g *generator) walk(path string, info os.FileInfo, err0 error) (err error) {
	if err0 != nil {
		return
	}
	if info.IsDir() {
		if path == g.name {
			return filepath.SkipDir
		}
		return
	}
	var in *os.File
	in, err = os.Open(path)
	if err != nil {
		return
	}
	defer in.Close()
	n := len(g.files)
	g.files = append(g.files, strings.TrimPrefix(path, g.root))
	m := len(g.files[n])
	if g.max < m {
		g.max = m
	}
	path = fmt.Sprintf("%d.go", n)
	path = filepath.Join(g.name, path)
	var out *os.File
	out, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer out.Close()
	_, err = fmt.Fprintf(out, "package %s\n\nvar b%d = [...]byte{\n", filepath.Base(g.name), n)
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

func (g *generator) generate() (err error) {
	err = os.MkdirAll(g.name, 0755)
	if err != nil {
		return
	}
	err = filepath.Walk(g.root, g.walk)
	if err != nil {
		return
	}
	var out *os.File
	path := filepath.Join(g.name, "main.go")
	out, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer out.Close()
	_, err = fmt.Fprintf(out, "package %s\n\nimport (\n\t\"net/http\"\n\n\t\"github.com/pshvedko/embedded\"\n)\n\n"+
		"func Dir() http.FileSystem {\n\treturn embedded.New(\n\t\tmap[string][]byte{\n", filepath.Base(g.name))
	if err != nil {
		return
	}
	for i, f := range g.files {
		_, err = fmt.Fprintf(out, "\t\t\t%q: %*sb%d[:],\n", f, g.max-len(f), "", i)
		if err != nil {
			return
		}
	}
	_, err = fmt.Fprintf(out, "\t\t}, %v)\n}\n", time.Now().Unix())
	return
}

func generate(root, name string) (err error) {
	g := generator{root: root, name: name}
	return g.generate()
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage:\n\t%s source package\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	r, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for i, a := range os.Args[1:] {
		if !filepath.IsAbs(a) {
			a = filepath.Join(r, a)
		}
		os.Args[1+i] = filepath.Clean(a)
	}
	err = generate(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
