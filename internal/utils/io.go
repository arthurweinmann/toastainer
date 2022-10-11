package utils

import "io"

type ProxyReader struct {
	totalread        int
	progress         chan int
	underlyingReader io.Reader
}

func NewProxyReader(underlying io.Reader, progress chan int) *ProxyReader {
	return &ProxyReader{underlyingReader: underlying, progress: progress}
}

func (pr *ProxyReader) Read(b []byte) (int, error) {
	n, err := pr.underlyingReader.Read(b)

	pr.totalread += n
	if pr.progress != nil {
		select {
		case pr.progress <- pr.totalread:
		default:
		}
	}

	return n, err
}

type TeeWriter struct {
	wrs []io.Writer
}

// TeeWriter returns the first writer results
func NewTeeWriter(wrs ...io.Writer) *TeeWriter {
	return &TeeWriter{wrs: wrs}
}

func (t *TeeWriter) Write(b []byte) (int, error) {
	if len(t.wrs) == 0 {
		return 0, io.ErrClosedPipe
	}

	n, err := t.wrs[0].Write(b)

	for i := 1; i < len(t.wrs); i++ {
		t.wrs[i].Write(b)
	}

	return n, err
}
