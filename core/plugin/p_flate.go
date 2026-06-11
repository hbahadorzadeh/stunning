package plugin

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"sync"
)

// flate is a size-optimization plugin using stdlib DEFLATE. It is pure Go and
// cross-compiles everywhere. Tiny or already-random payloads may expand
// slightly; place compression first in a chain (before aead/pad) so it operates
// on still-compressible plaintext.
//
// A flate.Writer/Reader carries a large (~600 KiB) internal window, so both are
// pooled per plugin instance and reset per frame; this turns a per-frame
// allocation into reuse, the dominant cost on the compression path.
func init() { Register("flate", newFlate) }

type flatePlugin struct {
	level int
	wpool sync.Pool // *flate.Writer
	rpool sync.Pool // io.ReadCloser implementing flate.Resetter
}

func newFlate(p Params) (Plugin, error) {
	lvl := p.Int("level", flate.DefaultCompression)
	if lvl != flate.DefaultCompression && (lvl < flate.NoCompression || lvl > flate.BestCompression) {
		return nil, fmt.Errorf("flate: level %d out of range", lvl)
	}
	return &flatePlugin{level: lvl}, nil
}

func (f *flatePlugin) Encode(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, _ := f.wpool.Get().(*flate.Writer)
	if w == nil {
		var err error
		if w, err = flate.NewWriter(&buf, f.level); err != nil {
			return nil, err
		}
	} else {
		w.Reset(&buf)
	}
	// Reset to io.Discard before pooling so the writer does not retain a
	// reference to buf (which holds the encoded output).
	defer func() {
		w.Reset(io.Discard)
		f.wpool.Put(w)
	}()
	if _, err := w.Write(src); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (f *flatePlugin) Decode(src []byte) ([]byte, error) {
	r, _ := f.rpool.Get().(io.ReadCloser)
	if r == nil {
		r = flate.NewReader(bytes.NewReader(src))
	} else if err := r.(flate.Resetter).Reset(bytes.NewReader(src), nil); err != nil {
		return nil, err
	}
	// Reset onto an empty reader before pooling so the reader does not retain a
	// reference to src.
	defer func() {
		_ = r.(flate.Resetter).Reset(bytes.NewReader(nil), nil)
		f.rpool.Put(r)
	}()
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("flate: decode: %w", err)
	}
	if err := r.Close(); err != nil {
		return nil, err
	}
	return out, nil
}

func (*flatePlugin) Close() error { return nil }
