package skkdic

import (
	"errors"
	"strconv"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/unicode"
)

type Encoding string

const (
	Auto       Encoding = ""
	EUCJP      Encoding = "euc-jp"
	eucJIS2004 Encoding = "euc-jis-2004"
	ShiftJIS   Encoding = "shift-jis"
	ISO2022JP  Encoding = "iso-2022-jp"
	UTF8       Encoding = "utf-8"
)

func (e Encoding) IsValid() bool {
	switch e {
	case EUCJP, eucJIS2004, ShiftJIS, ISO2022JP, UTF8:
		return true
	}

	return false
}

func (e Encoding) String() string {
	if e.IsValid() {
		return string(e)
	}
	return "unknown"
}

func (e Encoding) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuote(nil, string(e)), nil
}

func (e *Encoding) UnmarshalJSON(data []byte) error {
	s, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	return e.parse(s)
}

func (e Encoding) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

func (e *Encoding) UnmarshalText(data []byte) error {
	return e.parse(string(data))
}

func (e *Encoding) parse(s string) error {
	switch strings.ToLower(s) {
	case "euc-jp", "eucjp", "euc_jp":
		*e = EUCJP
	case "euc-jis-2004", "eucjis2004", "euc_jis_2004":
		*e = eucJIS2004
	case "shift-jis", "shiftjis", "shift_jis", "sjis":
		*e = ShiftJIS
	case "iso-2022-jp", "iso2022jp", "iso_2022_jp":
		*e = ISO2022JP
	case "utf-8", "utf8", "utf_8":
		*e = UTF8
	default:
		return errors.New("invalid encoding: " + s)
	}

	return nil
}

func (e Encoding) NewEncoder() *encoding.Encoder {
	return e.encoding().NewEncoder()
}

func (e Encoding) NewDecoder() *encoding.Decoder {
	return e.encoding().NewDecoder()
}

func (e Encoding) encoding() encoding.Encoding {
	switch e {
	case EUCJP, eucJIS2004:
		return japanese.EUCJP
	case ShiftJIS:
		return japanese.ShiftJIS
	case ISO2022JP:
		return japanese.ISO2022JP
	case UTF8:
		return unicode.UTF8
	default:
		return unicode.UTF8
	}
}
