package plugin

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
)

// flate is a size-optimization plugin using stdlib DEFLATE. It is pure Go and
// cross-compiles everywhere. Tiny or already-random payloads may expand
// slightly; place compression first in a chain (before aead/pad) so it operates
// on still-compressible plaintext.
func init() { Register("flate", newFlate) }

type flatePlugin struct {
	level int
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
	w, err := flate.NewWriter(&buf, f.level)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(src); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (f *flatePlugin) Decode(src []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(src))
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("flate: decode: %w", err)
	}
	if err := r.Close(); err != nil {
		return nil, err
	}
	return out, nil
}

func (f *flatePlugin) Close() error { return nil }
