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

// httpMaxChunk bounds a single chunk so a peer cannot force a large allocation
// in Deframe before the FramedConn frame-size cap is applied. It comfortably
// exceeds an encoded frame (write chunks are 8 KiB) while staying small.
const httpMaxChunk = 64 * 1024

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
		// ReadSlice (not ReadString) bounds the line to the bufio buffer, so a
		// peer that never sends a newline gets bufio.ErrBufferFull instead of an
		// unbounded allocation (OOM DoS).
		for {
			line, err := h.rd.ReadSlice('\n')
			if err != nil {
				return nil, err
			}
			// Byte check avoids a string allocation per header line.
			if len(line) == 2 && line[0] == '\r' && line[1] == '\n' || len(line) == 1 && line[0] == '\n' {
				break
			}
		}
		h.readHeader = true
	}

	sizeLineBuf, err := h.rd.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	// Trim CRLF and any chunk extension on the byte slice, converting to string
	// only once for parsing, to avoid per-frame allocations on the hot path.
	for len(sizeLineBuf) > 0 && (sizeLineBuf[len(sizeLineBuf)-1] == '\r' || sizeLineBuf[len(sizeLineBuf)-1] == '\n') {
		sizeLineBuf = sizeLineBuf[:len(sizeLineBuf)-1]
	}
	for i, b := range sizeLineBuf {
		if b == ';' {
			sizeLineBuf = sizeLineBuf[:i]
			break
		}
	}
	sizeLine := strings.TrimSpace(string(sizeLineBuf))
	n, err := strconv.ParseInt(sizeLine, 16, 32)
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
