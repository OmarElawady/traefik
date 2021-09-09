package metrics

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"

	gokitmetrics "github.com/go-kit/kit/metrics"
)

type BodyWrapper struct {
	io.ReadCloser
	counter gokitmetrics.Counter
}

func NewBodyWrapper(body io.ReadCloser, counter gokitmetrics.Counter) *BodyWrapper {
	return &BodyWrapper{
		body,
		counter,
	}
}

func (b *BodyWrapper) Read(p []byte) (int, error) {
	r, e := b.ReadCloser.Read(p)
	b.add(r)
	return r, e
}

func (b *BodyWrapper) add(v int) {
	b.counter.Add(float64(v))
}

type ResponseWriterWrapper struct {
	http.ResponseWriter
	counter gokitmetrics.Counter
}

func NewResponseWritrWrapper(rw http.ResponseWriter, counter gokitmetrics.Counter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		rw,
		counter,
	}
}

func (r *ResponseWriterWrapper) Write(d []byte) (int, error) {
	r.add(len(d))
	return r.ResponseWriter.Write(d)
}

// Hijack hijacks the connection.
func (r *ResponseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush sends any buffered data to the client.
func (r *ResponseWriterWrapper) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *ResponseWriterWrapper) add(v int) {
	r.counter.Add(float64(v))
}

func requestHeaderSize(req *http.Request) int {
	size := 1

	size += len("Host: ") + len(req.Host) + 1
	size += len(req.Proto) + 1
	size += len(req.URL.Path) + 1
	size += len(req.Method) + 1
	size += headerSize(req.Header) + 1

	return size
}

func responseHeaderSize(h http.Header, proto string, status int) int {
	size := 1
	size += len(proto) + 1
	size += len(fmt.Sprintf("%d", status)) + 1
	size += len(http.StatusText(status)) + 1
	size += headerSize(h) + 1

	return size
}

func headerSize(h http.Header) int {
	// some headers are not sent from the client
	// like X-Forwarded-Server, should they be counted or not?
	size := 0
	for k, v := range h {
		for _, e := range v {
			size += len(k) + len(e) + 3
		}
	}
	return size
}
