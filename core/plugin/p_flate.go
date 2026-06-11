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
//
// A flate.Writer/Reader carries a large (~600 KiB) internal window. The plugin
// is per connection and Encode/Decode run on separate (write/read) goroutines,
// so the writer and reader are kept as direct fields and reset per frame --
// guaranteeing reuse without sync.Pool overhead or GC reclaiming the window
// between frames.
func init() { Register("flate", newFlate) }

type flatePlugin struct {
	level int
	w     *flate.Writer
	r     io.ReadCloser
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
	if f.w == nil {
		var err error
		if f.w, err = flate.NewWriter(&buf, f.level); err != nil {
			return nil, err
		}
	} else {
		f.w.Reset(&buf)
	}
	// Reset to io.Discard on cleanup so the writer does not retain a reference to
	// buf (which holds the encoded output).
	defer func() {
		if f.w != nil {
			f.w.Reset(io.Discard)
		}
	}()
	if _, err := f.w.Write(src); err != nil {
		return nil, err
	}
	if err := f.w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (f *flatePlugin) Decode(src []byte) ([]byte, error) {
	if f.r == nil {
		f.r = flate.NewReader(bytes.NewReader(src))
	} else if err := f.r.(flate.Resetter).Reset(bytes.NewReader(src), nil); err != nil {
		return nil, err
	}
	// Reset onto an empty reader on cleanup so the reader does not retain a
	// reference to src.
	defer func() {
		if f.r != nil {
			_ = f.r.(flate.Resetter).Reset(bytes.NewReader(nil), nil)
		}
	}()
	out, err := io.ReadAll(f.r)
	if err != nil {
		return nil, fmt.Errorf("flate: decode: %w", err)
	}
	if err := f.r.Close(); err != nil {
		return nil, err
	}
	return out, nil
}

func (*flatePlugin) Close() error { return nil }
