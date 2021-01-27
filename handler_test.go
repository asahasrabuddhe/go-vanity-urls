package go_vanity_urls

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name   string
		config string
		path   string

		goImport     string
		goSource     string
		cacheControl string
	}{
		{
			name: "display default",
			config: `
host: go.example.com
paths:
  /package:
    repo: https://github.com/example/package`,
			path:         "/package",
			goImport:     "go.example.com/package git https://github.com/example/package",
			goSource:     "go.example.com/package https://github.com/example/package https://github.com/example/package/tree/master{/dir} https://github.com/example/package/blob/master{/dir}/{file}#L{line}",
			cacheControl: "public, max-age=86400",
		},
		{
			name: "display alt branch",
			config: `
host: go.example.com
cache_max_age: 60
paths:
  /package:
    repo: https://github.com/example/package
    branch: main`,
			path:         "/package",
			goImport:     "go.example.com/package git https://github.com/example/package",
			goSource:     "go.example.com/package https://github.com/example/package https://github.com/example/package/tree/main{/dir} https://github.com/example/package/blob/main{/dir}/{file}#L{line}",
			cacheControl: "public, max-age=60",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h, err := NewHandler([]byte(test.config))
			if err != nil {
				t.Errorf("NewHandler: %v", err)
				return
			}

			s := httptest.NewServer(h)
			res, err := http.Get(s.URL + test.path)
			if err != nil {
				s.Close()
				t.Errorf("http.Get: %v", err)
				return
			}

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("ioutil.ReadAll: %v", err)
				return
			}

			res.Body.Close()
			s.Close()

			if res.StatusCode != http.StatusOK {
				t.Errorf("StatusCode: want %d, got %d", http.StatusOK, res.StatusCode)
				return
			}

			if got := res.Header.Get("Cache-Control"); got != test.cacheControl {
				t.Errorf("got Cache-Control header: %s; want %s", got, test.cacheControl)
				return
			}

			if got := findMeta(data, "go-import"); got != test.goImport {
				t.Errorf("got meta go-import: %s, want: %s", got, test.goImport)
				return
			}

			if got := findMeta(data, "go-source"); got != test.goSource {
				t.Errorf("got meta go-source: %s, want: %s", got, test.goSource)
				return
			}
		})
	}
}

func findMeta(data []byte, name string) string {
	var buf bytes.Buffer

	buf.WriteString("<meta name=\"")
	buf.WriteString(name)
	buf.WriteString("\" content=\"")

	contentIndex := bytes.Index(data, buf.Bytes())
	if contentIndex == -1 {
		return ""
	}

	content := data[contentIndex+buf.Len():]

	contentValueIndex := bytes.IndexByte(content, '"')
	if contentValueIndex == -1 {
		return ""
	}

	return string(content[:contentValueIndex])
}
