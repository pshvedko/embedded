package embedded_test

import (
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/pshvedko/embedded"
)

func TestNew(t *testing.T) {
	type args struct {
		files     map[string][]byte
		timestamp int64
	}
	tests := []struct {
		name string
		args args
		want http.FileSystem
	}{
		// TODO: Add test cases.
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := embedded.New(tt.args.files, tt.args.timestamp); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleNew() {
	files := embedded.New(map[string][]byte{"/index.html": []byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Example</title>
</head>
<body>
    <div></div>
</body>
</html>
`)}, time.Now().Unix())
	err := http.ListenAndServe(":8080", http.FileServer(files))
	if err != nil {
		log.Fatal(err)
	}
}
