package skkdic

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/btree"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type Encoding string

const (
	Auto       Encoding = ""
	EUCJP      Encoding = "euc-jp"
	eucJIS2004 Encoding = "euc-jis-2004"
	ShiftJIS   Encoding = "shift_jis"
	ISO2022JP  Encoding = "iso-2022-jp"
	UTF8       Encoding = "utf-8"
)

func (e Encoding) isValid() bool {
	switch e {
	case EUCJP, eucJIS2004, ShiftJIS, ISO2022JP, UTF8:
		return true
	}

	return false
}

func (e Encoding) toEncoding() encoding.Encoding {
	switch e {
	case EUCJP, eucJIS2004:
		return japanese.EUCJP
	case ShiftJIS:
		return japanese.ShiftJIS
	case ISO2022JP:
		return japanese.ISO2022JP
	}

	return nil
}

type Dictionary struct {
	delimiter         string
	okuriAriEntries   *btree.BTreeG[*entry]
	okuriNashiEntries *btree.BTreeG[*entry]
}

type dicOptions struct {
	delimiter string
}

type Option interface {
	apply(*dicOptions)
}

type optionFunc func(*dicOptions)

func (f optionFunc) apply(opts *dicOptions) {
	f(opts)
}

func WithAnnotationDelimiter(delimiter string) Option {
	return optionFunc(func(opts *dicOptions) {
		opts.delimiter = delimiter
	})
}

func New(opts ...Option) *Dictionary {
	options := dicOptions{
		delimiter: ",",
	}
	for _, opt := range opts {
		opt.apply(&options)
	}

	return &Dictionary{
		delimiter:         options.delimiter,
		okuriAriEntries:   btree.NewG(2, lessEntryReverse),
		okuriNashiEntries: btree.NewG(2, lessEntry),
	}
}

type writeOptions struct {
	encoding Encoding
}

type WriteOption interface {
	apply(*writeOptions)
}

type writeOptionFunc func(*writeOptions)

func (f writeOptionFunc) apply(opts *writeOptions) {
	f(opts)
}

func WithOutputEncoding(e Encoding) WriteOption {
	return writeOptionFunc(func(opts *writeOptions) {
		if e.isValid() {
			opts.encoding = e
		}
	})
}

func (dic *Dictionary) Write(w io.Writer, opts ...WriteOption) error {
	options := writeOptions{
		encoding: UTF8,
	}
	for _, opt := range opts {
		opt.apply(&options)
	}

	enc := options.encoding.toEncoding()

	var bw *bufio.Writer
	if enc == nil {
		bw = bufio.NewWriter(w)
	} else {
		bw = bufio.NewWriter(transform.NewWriter(w, enc.NewEncoder()))
	}

	var err error
	_, err = bw.WriteString(";; -*- coding: " + string(options.encoding) + " -*-\n")
	if err != nil {
		return err
	}

	_, err = bw.WriteString(";; okuri-ari entries.\n")
	if err != nil {
		return err
	}
	dic.okuriAriEntries.Ascend(func(e *entry) bool {
		err = writeEntry(bw, e)
		return err == nil
	})
	if err != nil {
		return err
	}

	_, err = bw.WriteString(";; okuri-nasi entries.\n")
	if err != nil {
		return err
	}
	dic.okuriNashiEntries.Ascend(func(e *entry) bool {
		err = writeEntry(bw, e)
		return err == nil
	})
	if err != nil {
		return err
	}

	err = bw.Flush()
	if err != nil {
		return err
	}

	return nil
}

func writeEntry(w *bufio.Writer, e *entry) error {
	var err error

	_, err = w.WriteString(e.Midashi)
	if err != nil {
		return err
	}
	err = w.WriteByte(' ')
	if err != nil {
		return err
	}
	_, err = w.WriteString(joinCandidates(e.Candidates))
	if err != nil {
		return err
	}
	err = w.WriteByte('\n')
	if err != nil {
		return err
	}

	return nil
}

func joinCandidates(candidates []*Candidate) string {
	if len(candidates) == 0 {
		return "//"
	}

	var s strings.Builder

	for _, candidate := range candidates {
		s.WriteByte('/')
		s.WriteString(candidate.String())
	}

	s.WriteByte('/')

	return s.String()
}

type MergeMode int

const (
	Add MergeMode = iota
	Sub
	And
)

type readOptions struct {
	encoding Encoding
}

type ReadOption interface {
	apply(*readOptions)
}

type readOptionFunc func(*readOptions)

func (f readOptionFunc) apply(opts *readOptions) {
	f(opts)
}

func WithInputEncoding(e Encoding) ReadOption {
	return readOptionFunc(func(opts *readOptions) {
		if e.isValid() {
			opts.encoding = e
		}
	})
}

func (dic *Dictionary) ReadFile(name string, mode MergeMode, opts ...ReadOption) error {
	if dic == nil {
		return errors.New("Dictionary is nil")
	}

	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	return dic.Read(file, mode, opts...)
}

func (dic *Dictionary) Read(r io.Reader, mode MergeMode, opts ...ReadOption) error {
	if dic == nil {
		return errors.New("Dictionary is nil")
	}

	options := readOptions{
		encoding: Auto,
	}
	for _, opt := range opts {
		opt.apply(&options)
	}

	var reader io.Reader

	var enc encoding.Encoding
	if options.encoding == Auto {
		b := bufio.NewReader(r)
		first, err := b.ReadBytes('\n')
		if len(first) > 0 {
			enc = extractEncoding(string(first))
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if len(first) == 0 {
			reader = b
		} else {
			reader = io.MultiReader(bytes.NewReader(first), b)
		}
	} else {
		enc = options.encoding.toEncoding()
		reader = r
	}

	if enc != nil {
		reader = transform.NewReader(reader, enc.NewDecoder())
	}

	s := bufio.NewScanner(reader)
	for s.Scan() {
		line := s.Text()
		if len(line) > 0 && line[0] == ';' {
			continue
		}

		midashi, candidates := parseLine(line)

		dic.processCandidates(midashi, candidates, mode)
	}

	if mode == And {
		cleanEntries(dic.okuriAriEntries)
		cleanEntries(dic.okuriNashiEntries)
	}

	return nil
}

var magicRegexp = regexp.MustCompile(`-\*-.*[ \t]coding:[ \t]*([^ \t;]+?)[ \t;].*-\*-`)

func extractEncoding(line string) encoding.Encoding {
	matches := magicRegexp.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil
	}

	return Encoding(matches[1]).toEncoding()
}

func parseLine(s string) (midashi string, candidates []*Candidate) {
	midashi, strCandidates, found := strings.Cut(s, " ")
	if !found || midashi == "" {
		return "", nil
	}

	candidates = parseCandidates(midashi, strCandidates)
	if len(candidates) == 0 {
		return "", nil
	}

	return midashi, candidates
}

func (dic *Dictionary) processCandidates(midashi string, candidates []*Candidate, mode MergeMode) {
	if midashi == "" || len(candidates) == 0 {
		return
	}

	e := &entry{
		Midashi:    midashi,
		Candidates: candidates,
	}

	var entries *btree.BTreeG[*entry]
	if e.IsOkuriAri() {
		entries = dic.okuriAriEntries
	} else {
		entries = dic.okuriNashiEntries
	}

	switch mode {
	case Sub:
		subEntry(entries, e)
	case And:
		andEntry(entries, e)
	default:
		addEntry(entries, e, dic.delimiter)
	}
}

func addEntry(tree *btree.BTreeG[*entry], e *entry, delimiter string) {
	ce, found := tree.Get(e)
	if !found {
		ce = &entry{Midashi: e.Midashi}
		tree.ReplaceOrInsert(ce)
	}

	for _, c := range e.Candidates {
		ce.addCandidate(c, delimiter)
	}
}

func subEntry(tree *btree.BTreeG[*entry], e *entry) {
	ce, found := tree.Get(e)
	if !found {
		return
	}

	for _, c := range e.Candidates {
		ce.removeCandidate(c)
		if len(ce.Candidates) == 0 {
			tree.Delete(ce)
			break
		}
	}
}

func andEntry(tree *btree.BTreeG[*entry], e *entry) {
	ce, found := tree.Get(e)
	if !found {
		return
	}

	for _, c := range e.Candidates {
		ce.andCandidate(c)
	}
}

func cleanEntries(tree *btree.BTreeG[*entry]) {
	var removeEntries []*entry
	tree.Ascend(func(e *entry) bool {
		e.clean()
		if len(e.Candidates) == 0 {
			removeEntries = append(removeEntries, e)
		}
		return true
	})
	for _, e := range removeEntries {
		tree.Delete(e)
	}
}

func (dic *Dictionary) AddCandidates(midashi string, candidates []*Candidate) {
	dic.processCandidates(midashi, candidates, Add)
}

func (dic *Dictionary) SubCandidates(midashi string, candidates []*Candidate) {
	dic.processCandidates(midashi, candidates, Sub)
}

func (dic *Dictionary) AndCandidates(midashi string, candidates []*Candidate) {
	dic.processCandidates(midashi, candidates, And)
}

func (dic *Dictionary) RemoveCandidates(midashi string) {
	var entries *btree.BTreeG[*entry]
	if isOkuriAri(midashi) {
		entries = dic.okuriAriEntries
	} else {
		entries = dic.okuriNashiEntries
	}

	entries.Delete(&entry{Midashi: midashi})
}

func (dic *Dictionary) Lookup(midashi string) []*Candidate {
	var entries *btree.BTreeG[*entry]
	if isOkuriAri(midashi) {
		entries = dic.okuriAriEntries
	} else {
		entries = dic.okuriNashiEntries
	}

	e, found := entries.Get(&entry{Midashi: midashi})
	if !found {
		return nil
	}

	return e.Candidates
}

func (dic *Dictionary) Complete(midashi string) []string {
	var completion []string

	dic.okuriNashiEntries.AscendRange(&entry{Midashi: midashi}, &entry{Midashi: midashi + string(unicode.MaxRune)}, func(e *entry) bool {
		if strings.HasPrefix(e.Midashi, midashi) {
			completion = append(completion, e.Midashi)
			return true
		}

		return false
	})

	return completion
}
