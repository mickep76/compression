package compress

import (
	"bytes"
	"fmt"
	"io"
)

var algorithms = make(map[string]Algorithm)

// Algorithm interface.
type Algorithm interface {
	NewAlgorithm() Algorithm
	Ext() string
	NewEncoder(w io.Writer) (Encoder, error)
	NewDecoder(r io.Reader) (Decoder, error)
	Encode(v []byte) ([]byte, error)
	Decode(v []byte) ([]byte, error)
	SetLevel(level Level) error
	SetLitWidth(width int) error
	SetEndian(endian Endian) error
}

// Encoder interface.
type Encoder interface {
	Write(v []byte) (int, error)
	Close() error
}

// Decoder interface.
type Decoder interface {
	Read(v []byte) (int, error)
	Close() error
}

// Option variadic function.
type Option func(Algorithm) error

// Level compression level.
type Level int

const (
	// NoCompression no compression.
	NoCompression Level = 0

	// BestSpeed best speed.
	BestSpeed Level = 1

	// BestCompression best compression.
	BestCompression Level = 9

	// DefaultCompression default compression.
	DefaultCompression Level = -1

	// HuffmanOnly huffman only.
	HuffmanOnly Level = -2
)

// Endian the order in which bytes are arranged into larger values.
type Endian int

const (
	// Little endian LSB (Least Significant Bit).
	Little Endian = 0

	// Big endian MSB (Most Significant Bit).
	Big Endian = 1
)

// Register algorithm.
func Register(name string, algorithm Algorithm) {
	algorithms[name] = algorithm
}

// Algorithms registered.
func Algorithms() []string {
	l := []string{}
	for a := range algorithms {
		l = append(l, a)
	}
	return l
}

// Registered is the algorithm registered.
func Registered(name string) error {
	_, ok := algorithms[name]
	if !ok {
		return fmt.Errorf("algorithm not registered: %s", name)
	}
	return nil
}

// NewAlgorithm variadic constructor.
func NewAlgorithm(name string, opts ...Option) (Algorithm, error) {
	a, ok := algorithms[name]
	if !ok {
		return nil, fmt.Errorf("algorithm not registered: %s", name)
	}
	a = a.NewAlgorithm()
	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}
	return a, nil
}

// WithLevel compression level.
// Supported by gzip, zlib.
func WithLevel(level Level) Option {
	return func(a Algorithm) error {
		return a.SetLevel(level)
	}
}

// WithLitWidth the number of bit's to use for literal codes.
// Supported by lzw.
func WithLitWidth(width int) Option {
	return func(a Algorithm) error {
		return a.SetLitWidth(width)
	}
}

// WithEndian either MSB (most significant byte) or LSB (least significant byte).
// Supported by lzw.
func WithEndian(endian Endian) Option {
	return func(a Algorithm) error {
		return a.SetEndian(endian)
	}
}

// Encode algorithm.
func Encode(a Algorithm, v []byte) ([]byte, error) {
	var buf bytes.Buffer
	e, err := a.NewEncoder(&buf)
	if err != nil {
		return nil, err
	}

	if _, err := e.Write(v); err != nil {
		return nil, err
	}

	if err := e.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decode algorithm.
func Decode(a Algorithm, v []byte) ([]byte, error) {
	d, err := a.NewDecoder(bytes.NewBuffer(v))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, d); err != nil {
		return nil, err
	}

	if err := d.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
