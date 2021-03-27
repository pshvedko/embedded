package embedded

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrNotDirectory = errors.New("not a directory")
	ErrIsDirectory  = errors.New("is a directory")
)

type holder interface {
	http.File
	Open() (http.File, error)
}

type folder map[string]holder

func (f folder) keys() (keys []string) {
	for key := range f {
		keys = append(keys, key)
	}
	return
}

func (f folder) Open(name string) (http.File, error) {
	v, ok := f[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return v.Open()
}

type stat struct {
	t time.Time
	n string
}

func (s *stat) Name() string { return s.n }

func (s *stat) ModTime() time.Time { return s.t }

func (s *stat) Sys() interface{} { return nil }

func (s *stat) Close() error { return nil }

type file struct {
	stat
	r bytes.Reader
	p string
}

func (f *file) Open() (http.File, error) { v := *f; return &v, nil }

func (f *file) Mode() os.FileMode { return 0444 }

func (f *file) Name() string { return f.p }

func (f *file) Size() int64 { return f.r.Size() }

func (f *file) Read(p []byte) (n int, err error) { return f.r.Read(p) }

func (f *file) Seek(offset int64, whence int) (int64, error) { return f.r.Seek(offset, whence) }

func (f *file) IsDir() bool { return false }

func (f *file) Readdir(int) ([]os.FileInfo, error) { return nil, ErrNotDirectory }

func (f *file) Stat() (os.FileInfo, error) { return f, nil }

func (f *file) String() string { return fmt.Sprintf("%s,%d,%v", f.stat.Name(), f.Size(), f.Mode()) }

type dir struct {
	stat
	p string
	f []holder
	n int
}

func (d *dir) Open() (http.File, error) { v := *d; return &v, nil }

func (d *dir) Mode() os.FileMode { return 0555 }

func (d *dir) Read([]byte) (int, error) { return 0, ErrIsDirectory }

func (d *dir) Seek(int64, int) (int64, error) { return 0, ErrIsDirectory }

func (d *dir) Name() string { return d.p }

func (d *dir) Size() int64 { return 0 }

func (d *dir) Readdir(m int) (infos []os.FileInfo, err error) { panic("implement me") }

func (d *dir) IsDir() bool { return true }

func (d *dir) Stat() (os.FileInfo, error) { return d, nil }

func (d *dir) String() string { return fmt.Sprintf("%s:%v", d.stat.Name(), d.f) }

func (d *dir) add(h holder) {
	if d != h {
		d.f = append(d.f, h)
	}
}

func (f folder) up(d *dir) *dir {
	p, n := filepath.Split(filepath.Dir(d.p))
	_, ok := f[p]
	if !ok {
		f[p] = &dir{stat: stat{n: n, t: d.t}, p: p}
		f.up(f[p].(*dir)).add(f[p])
	}
	return f[p].(*dir)
}

func New(files map[string][]byte, timestamp int64) http.FileSystem {
	f := folder{}
	for n, b := range files {
		p, _ := filepath.Split(n)
		_, ok := f[p]
		if !ok {
			f[p] = &dir{stat: stat{t: time.Unix(timestamp, 0), n: filepath.Base(p)}, p: p}
			f.up(f[p].(*dir)).add(f[p])
		}
		f[n] = &file{stat: stat{t: time.Unix(timestamp, 0), n: filepath.Base(n)}, p: n, r: *bytes.NewReader(b)}
		f[p].(*dir).add(f[n])
	}
	return f
}
