package got

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Scanner is a tokenizer for got templates.
type Scanner struct {
	// Reader is held until first read.
	r io.Reader

	// Entire reader is read into a buffer.
	b []byte
	i int

	pos Pos
}

// NewScanner initializes a new scanner with a given reader.
func NewScanner(r io.Reader, path string) *Scanner {
	return &Scanner{r: r, pos: Pos{Path: path, LineNo: 1}}
}

// Scan returns the next block from the reader.
func (s *Scanner) Scan() (Block, error) {
	if err := s.init(); err != nil {
		return nil, err
	}

	switch s.peek() {
	case '<':
		// Special handling for got blocks.
		if s.peekN(3) == "<%%" {
			return &TextBlock{Pos: s.pos, Content: "<%"}, nil
		} else if s.peekN(2) == "<%" {
			return s.scanCodeBlock()
		}

	case eof:
		return nil, io.EOF
	}

	return s.scanTextBlock()
}

func (s *Scanner) scanTextBlock() (*TextBlock, error) {
	buf := bytes.NewBufferString(s.readN(1))
	blk := &TextBlock{Pos: s.pos}

loop:
	for {
		switch s.peek() {
		case '<':
			if s.peekN(2) == "<%" {
				break loop
			}

		case eof:
			break loop
		}

		buf.WriteRune(s.read())
	}

	blk.Content = string(buf.Bytes())
	return blk, nil
}

func (s *Scanner) scanCodeBlock() (*CodeBlock, error) {
	blk := &CodeBlock{Pos: s.pos}
	assert(s.readN(2) == "<%")

	content, err := s.scanContent()
	if err != nil {
		return nil, err
	}
	blk.Content = strings.TrimSpace(content)
	return blk, nil
}

// scans the reader until }} is reached.
func (s *Scanner) scanContent() (string, error) {
	var buf bytes.Buffer
	for {
		ch := s.read()
		if ch == eof {
			return "", &SyntaxError{Message: "Expected close tag, found EOF", Pos: s.pos}
		} else if ch == '%' {
			ch := s.read()
			if ch == eof {
				return "", &SyntaxError{Message: "Expected close tag, found EOF", Pos: s.pos}
			} else if ch == '>' {
				break
			} else if ch == '%' && s.peek() == '>' {
				buf.WriteRune(ch)
				buf.WriteRune(s.read())
			} else {
				buf.WriteRune('%')
				buf.WriteRune(ch)
			}
		} else {
			buf.WriteRune(ch)
		}
	}
	return string(buf.Bytes()), nil
}

// init slurps the reader on first scan.
func (s *Scanner) init() (err error) {
	if s.b != nil {
		return nil
	}
	s.b, err = ioutil.ReadAll(s.r)
	return err
}

// read reads the next rune and moves the position forward.
func (s *Scanner) read() rune {
	if s.i >= len(s.b) {
		return eof
	}

	ch, n := utf8.DecodeRune(s.b[s.i:])
	s.i += n

	if ch == '\n' {
		s.pos.LineNo++
	}
	return ch
}

// readN reads the next n characters and moves the position forward.
func (s *Scanner) readN(n int) string {
	var buf bytes.Buffer
	for i := 0; i < n; i++ {
		ch := s.read()
		if ch == eof {
			break
		}
		buf.WriteRune(ch)
	}
	return buf.String()
}

// peek reads the next rune but does not move the position forward.
func (s *Scanner) peek() rune {
	if s.i >= len(s.b) {
		return eof
	}
	ch, _ := utf8.DecodeRune(s.b[s.i:])
	return ch
}

// peekN reads the next n runes but does not move the position forward.
func (s *Scanner) peekN(n int) string {
	if s.i >= len(s.b) {
		return ""
	}
	b := s.b[s.i:]

	var buf bytes.Buffer
	for i := 0; i < n && len(b) > 0; i++ {
		ch, sz := utf8.DecodeRune(b)
		b = b[sz:]
		buf.WriteRune(ch)
	}
	return buf.String()
}

// peekIgnoreWhitespace reads the non-whitespace rune.
func (s *Scanner) peekIgnoreWhitespace() rune {
	var b []byte
	if s.i < len(s.b) {
		b = s.b[s.i:]
	}

	for i := 0; ; i++ {
		if len(b) == 0 {
			return eof
		}

		ch, sz := utf8.DecodeRune(b)
		if !isWhitespace(ch) {
			return ch
		}

		b = b[sz:]
	}
}

func (s *Scanner) skipWhitespace() {
	for ch := s.peek(); isWhitespace(ch); ch = s.peek() {
		s.read()
	}
	return
}

const eof = rune(0)

type SyntaxError struct {
	Message string
	Pos     Pos
}

func NewSyntaxError(pos Pos, format string, args ...interface{}) *SyntaxError {
	return &SyntaxError{
		Message: fmt.Sprintf(format, args...),
		Pos:     pos,
	}
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("%s at %s:%d", e.Message, e.Pos.Path, e.Pos.LineNo)
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func runeString(ch rune) string {
	switch ch {
	case eof:
		return "EOF"
	case ' ':
		return "<space>"
	case '\t':
		return `\t`
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	default:
		return string(ch)
	}
}

func assert(condition bool) {
	if !condition {
		panic("assertion failed")
	}
}
