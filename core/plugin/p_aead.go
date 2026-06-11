package plugin

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20poly1305"
)

// aead authenticates and encrypts each frame. It defaults to XChaCha20-Poly1305;
// algo=aesgcm selects AES-256-GCM (hardware-accelerated).
//
// Nonces use a per-connection random base XORed with a monotonic counter, so
// every frame on a connection gets a unique nonce (no birthday-bound collision
// risk even for AES-GCM's 96-bit nonce after many frames), and the two tunnel
// directions draw independent random bases so they never collide. The full nonce
// is prepended to each frame, so a single instance is its own inverse
// (Decode(Encode(x)) == x).
//
// Params:
//
//	key  hex-encoded secret. If shorter/longer than required it is hashed to the
//	     cipher's key size via BLAKE2b, so any passphrase-derived hex works.
//	algo "chacha" (default) or "aesgcm".
func init() { Register("aead", newAEAD) }

type aeadPlugin struct {
	aead      cipher.AEAD
	nonceSize int
	base      []byte // per-connection random nonce base
	counter   uint64 // monotonic, XORed into the base per frame
}

func newAEAD(p Params) (Plugin, error) {
	raw, err := p.Bytes("key", nil)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("aead: missing required param key")
	}
	algo := p.String("algo", "chacha")

	var aead cipher.AEAD
	switch algo {
	case "chacha", "xchacha", "chacha20poly1305":
		key := blake2b.Sum256(raw) // normalize to 32 bytes
		aead, err = chacha20poly1305.NewX(key[:])
	case "aesgcm", "aes":
		key := blake2b.Sum256(raw)
		var block cipher.Block
		block, err = aes.NewCipher(key[:])
		if err == nil {
			aead, err = cipher.NewGCM(block)
		}
	default:
		return nil, fmt.Errorf("aead: unknown algo %q", algo)
	}
	if err != nil {
		return nil, fmt.Errorf("aead: %w", err)
	}
	base := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, base); err != nil {
		return nil, fmt.Errorf("aead: nonce base: %w", err)
	}
	return &aeadPlugin{aead: aead, nonceSize: aead.NonceSize(), base: base}, nil
}

func (a *aeadPlugin) Encode(src []byte) ([]byte, error) {
	nonce := make([]byte, a.nonceSize, a.nonceSize+len(src)+a.aead.Overhead())
	copy(nonce, a.base)
	c := a.counter
	a.counter++
	// XOR the counter into the trailing 8 bytes of the base: unique per frame.
	for i := 0; i < 8 && i < a.nonceSize; i++ {
		nonce[a.nonceSize-1-i] ^= byte(c >> (8 * i))
	}
	return a.aead.Seal(nonce, nonce, src, nil), nil
}

func (a *aeadPlugin) Decode(src []byte) ([]byte, error) {
	if len(src) < a.nonceSize {
		return nil, fmt.Errorf("aead: frame shorter than nonce")
	}
	nonce, ct := src[:a.nonceSize], src[a.nonceSize:]
	out, err := a.aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("aead: open: %w", err)
	}
	return out, nil
}

func (*aeadPlugin) Close() error { return nil }
