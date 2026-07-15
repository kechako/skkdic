package skkdic

import "errors"

type dicOptions struct {
	delimiter string
}

// A Option represents an option for configuring the SKK dictionary.
type Option interface {
	apply(*dicOptions)
}

type withAnnotationDelimiter string

func (w withAnnotationDelimiter) apply(opts *dicOptions) {
	opts.delimiter = string(w)
}

func WithAnnotationDelimiter(delimiter string) Option {
	return withAnnotationDelimiter(delimiter)
}

type writeOptions struct {
	encoding Encoding
}

func (opts *writeOptions) Validate() error {
	if !opts.encoding.IsValid() {
		return errors.New("invalid encoding: " + string(opts.encoding))
	}
	return nil
}

type WriteOption interface {
	apply(*writeOptions)
}

type withOutputEncoding Encoding

func (w withOutputEncoding) apply(opts *writeOptions) {
	if !Encoding(w).IsValid() {
		opts.encoding = Encoding(w)
	}
}

func WithOutputEncoding(e Encoding) WriteOption {
	return withOutputEncoding(e)
}
