package plugin

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// http-mimic disguises the tunnel as an HTTP/1.1 chunked transfer. As a Framer it
// sends an HTTP header preamble first (so a DPI box allowlists the flow as HTTP),
// then carries each payload as one HTTP chunk, whose hex length line provides the
// framing. Encode/Decode are identity; pair it with an inner aead/flate chain.
//
// Place it last in the chain (outermost on the wire):
//
//	flate,aead?key=...,http-mimic
func init() { Register("http-mimic", newHTTPMimic) }

const httpPreamble = "POST /api/v2/upload HTTP/1.1\r\n" +
	"Host: cdn.example.com\r\n" +
	"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64)\r\n" +
	"Accept: */*\r\n" +
	"Content-Type: application/octet-stream\r\n" +
	"Transfer-Encoding: chunked\r\n\r\n"

const httpMaxChunk = 1 << 20

type httpMimic struct {
	wroteHeader bool
	readHeader  bool
	rd          *bufio.Reader
}

func newHTTPMimic(_ Params) (Plugin, error) { return &httpMimic{}, nil }

func (*httpMimic) Encode(src []byte) ([]byte, error) { return src, nil }
func (*httpMimic) Decode(src []byte) ([]byte, error) { return src, nil }
func (*httpMimic) Close() error                      { return nil }

func (h *httpMimic) Frame(payload []byte) ([]byte, error) {
	chunk := []byte(fmt.Sprintf("%x\r\n", len(payload)))
	chunk = append(chunk, payload...)
	chunk = append(chunk, '\r', '\n')
	if h.wroteHeader {
		return chunk, nil
	}
	h.wroteHeader = true
	return append([]byte(httpPreamble), chunk...), nil
}

func (h *httpMimic) Deframe(r io.Reader) ([]byte, error) {
	if h.rd == nil {
		h.rd = bufio.NewReader(r)
	}
	if !h.readHeader {
		// Consume header lines up to the blank line terminating the headers.
		for {
			line, err := h.rd.ReadString('\n')
			if err != nil {
				return nil, err
			}
			if line == "\r\n" || line == "\n" {
				break
			}
		}
		h.readHeader = true
	}

	sizeLine, err := h.rd.ReadString('\n')
	if err != nil {
		return nil, err
	}
	sizeLine = strings.TrimRight(sizeLine, "\r\n")
	// A chunk extension (";...") may follow the size; ignore it.
	if i := strings.IndexByte(sizeLine, ';'); i >= 0 {
		sizeLine = sizeLine[:i]
	}
	n, err := strconv.ParseInt(strings.TrimSpace(sizeLine), 16, 32)
	if err != nil {
		return nil, fmt.Errorf("http-mimic: bad chunk size %q: %w", sizeLine, err)
	}
	if n < 0 || n > httpMaxChunk {
		return nil, fmt.Errorf("http-mimic: chunk size %d out of range", n)
	}
	body := make([]byte, n)
	if _, err := io.ReadFull(h.rd, body); err != nil {
		return nil, err
	}
	// Consume the trailing CRLF after the chunk data.
	var crlf [2]byte
	if _, err := io.ReadFull(h.rd, crlf[:]); err != nil {
		return nil, err
	}
	return body, nil
}
