package prome

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func TestNewPlugin(t *testing.T) {
	in := `# HELP test metrics
# 	TYPE test_metrics_seconds counter
test_metrics_seconds{role="a" } 4.9351e-05
test_metrics_seconds{role="b",group="d"} 8.3835e-05
test_metrics_seconds{ role="c", group="e"} 8.3835e-05

# HELP test more metrics
# 	TYPE test_more_metrics_bytes gauge
test_more_metrics_bytes{role="a" } 256.0`

	ts := newMockServer(in)
	targets := []string{ts.URL}
	prefix := ""
	ctx := context.Background()
	p, err := NewPlugin(ctx, NewHTTPClient(), targets, prefix)
	if err != nil {
		t.Fatal(err)
	}
	g := p.GraphDefinition()
	if len(g) != 2 {
		t.Errorf("got %v want %v", len(g), 2)
	}

	m, _ := p.FetchMetrics()
	if len(m) != 4 {
		t.Errorf("got %v want %v", len(m), 4)
	}
}
func TestNewPluginGzip(t *testing.T) {
	in := `# HELP test metrics
# 	TYPE test_metrics_seconds counter
test_metrics_seconds{role="a" } 4.9351e-05
test_metrics_seconds{role="b",group="d"} 8.3835e-05
test_metrics_seconds{ role="c", group="e"} 8.3835e-05

# HELP test more metrics
# 	TYPE test_more_metrics_bytes gauge
test_more_metrics_bytes{role="a" } 256.0`

	ts := newMockGzipServer(in)
	targets := []string{ts.URL}
	prefix := ""
	ctx := context.Background()
	p, err := NewPlugin(ctx, NewHTTPClient(), targets, prefix)
	if err != nil {
		t.Fatal(err)
	}
	g := p.GraphDefinition()
	if len(g) != 2 {
		t.Errorf("got %v want %v", len(g), 2)
	}

	m, _ := p.FetchMetrics()
	if len(m) != 4 {
		t.Errorf("got %v want %v", len(m), 4)
	}
}

func newMockServer(in string) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", in)
	})
	return httptest.NewServer(handler)
}

func newMockGzipServer(in string) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		w = gzipResponseWriter{Writer: gz, ResponseWriter: w}
		fmt.Fprintf(w, "%s", in)
	})
	return httptest.NewServer(handler)
}
