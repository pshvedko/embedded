package embedded

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
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
}

func (f *file) Open() (http.File, error) { v := *f; return &v, nil }

func (f *file) Mode() os.FileMode { return 0444 }

func (f *file) Size() int64 { return f.r.Size() }

func (f *file) Read(p []byte) (n int, err error) { return f.r.Read(p) }

func (f *file) Seek(offset int64, whence int) (int64, error) { return f.r.Seek(offset, whence) }

func (f *file) IsDir() bool { return false }

func (f *file) Readdir(int) ([]os.FileInfo, error) { panic("implement me") }

func (f *file) Stat() (os.FileInfo, error) { return f, nil }

func (f *file) String() string { return fmt.Sprintf("%d,%v", f.Size(), f.Mode()) }

type dir struct {
	stat
	f folder
}

func (d *dir) Open() (http.File, error) { v := *d; return &v, nil }

func (d *dir) Mode() os.FileMode { return 0555 }

func (d *dir) Read([]byte) (n int, err error) { panic("implement me") }

func (d *dir) Seek(int64, int) (int64, error) { panic("implement me") }

func (d *dir) Size() int64 { panic("implement me") }

func (d *dir) Readdir(int) ([]os.FileInfo, error) { panic("implement me") }

func (d *dir) IsDir() bool { return true }

func (d *dir) Stat() (os.FileInfo, error) { return d, nil }

func (d *dir) String() string { return fmt.Sprintf("%v", d.f.keys()) }

func (f folder) Open(name string) (http.File, error) {
	v, ok := f[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return v.Open()
}

func New(files map[string][]byte, timestamp int64) http.FileSystem {
	f := folder{}
	for n, b := range files {
		s := stat{t: time.Unix(timestamp, 0), n: filepath.Dir(n)}
		p, q := filepath.Split(n)
		_, ok := f[p]
		if !ok {
			f[p] = &dir{stat: s, f: folder{}}
		}
		if len(q) > 0 {
			f[n] = &file{stat: s, r: *bytes.NewReader(b)}
			f[p].(*dir).f[q] = f[n]
		}
	}
	return f
}
