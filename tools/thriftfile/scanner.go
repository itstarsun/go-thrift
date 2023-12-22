package thriftfile

import (
	"fmt"
	gotoken "go/token"
	"unicode"
	"unicode/utf8"
)

const (
	bom = 0xFEFF
	eof = -1
)

type scanner struct {
	file *gotoken.File
	src  []byte
	err  func(pos gotoken.Position, msg string)

	ch         rune
	offset     int
	rdOffset   int
	lineOffset int

	tok token
	pos gotoken.Pos
}

func (s *scanner) init(file *gotoken.File, src []byte, err func(pos gotoken.Position, msg string)) {
	s.file = file
	s.src = src
	s.err = err
	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.tok = _ILLEGAL
	s.pos = gotoken.NoPos
	s.nextch()
	if s.ch == bom {
		s.nextch()
	}
}

func (s *scanner) lit() string {
	return string(s.src[s.file.Offset(s.pos):s.offset])
}

func (s *scanner) error(offs int, msg string) {
	if s.err != nil {
		s.err(s.file.Position(s.file.Pos(offs)), msg)
	}
}

func (s *scanner) errorf(offs int, format string, args ...any) {
	s.error(offs, fmt.Sprintf(format, args...))
}

func (s *scanner) nextch() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal UTF-8 encoding")
			} else if r == bom && s.offset > 0 {
				s.error(s.offset, "illegal byte order mark")
			}
		}
		s.rdOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch = eof
	}
}

func (s *scanner) next() {
redo:
	for s.ch == '\n' || s.ch == '\r' || s.ch == '\t' || s.ch == ' ' {
		s.nextch()
	}
	s.pos = s.file.Pos(s.offset)

	for isLetter(s.ch) || s.ch >= utf8.RuneSelf && s.atIdent(true) {
		s.nextch()
		s.ident()
		return
	}

	switch s.ch {
	case eof:
		s.tok = _EOF

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		s.number(false)

	case '"', '\'':
		s.string()

	case '+':
		s.nextch()
		if isDecimal(s.ch) {
			s.number(false)
		} else {
			s.tok = _ADD
		}

	case '-':
		s.nextch()
		if isDecimal(s.ch) {
			s.number(false)
		} else {
			s.tok = _SUB
		}

	case '*':
		s.nextch()
		s.tok = _MUL

	case '/':
		s.nextch()
		if s.ch == '/' {
			s.nextch()
			s.lineComment()
			return
		}
		switch s.ch {
		case '/':
			s.nextch()
			s.lineComment()
		case '*':
			s.nextch()
			s.fullComment()
		default:
			s.tok = _QUO
		}

	case '%':
		s.nextch()
		s.tok = _REM

	case '&':
		s.nextch()
		s.tok = _AND

	case '|':
		s.nextch()
		s.tok = _OR

	case '^':
		s.nextch()
		s.tok = _XOR

	case '=':
		s.nextch()
		if s.ch == '=' {
			s.nextch()
			s.tok = _EQL
		} else {
			s.tok = _ASSIGN
		}

	case '<':
		s.nextch()
		s.tok = _LSS

	case '>':
		s.nextch()
		s.tok = _GTR

	case '!':
		s.nextch()
		s.tok = _NOT

	case '(':
		s.nextch()
		s.tok = _LPAREN

	case '[':
		s.nextch()
		s.tok = _LBRACK

	case '{':
		s.nextch()
		s.tok = _LBRACE

	case ',':
		s.nextch()
		s.tok = _COMMA

	case '.':
		s.nextch()
		s.tok = _PERIOD

	case ')':
		s.nextch()
		s.tok = _RPAREN

	case ']':
		s.nextch()
		s.tok = _RBRACK

	case '}':
		s.nextch()
		s.tok = _RBRACE

	case ';':
		s.nextch()
		s.tok = _SEMICOLON

	case ':':
		s.nextch()
		s.tok = _COLON

	case '#':
		s.nextch()
		s.lineComment()

	default:
		s.errorf(s.offset, "invalid character %#U", s.ch)
		s.nextch()
		goto redo
	}
}

func (s *scanner) ident() {
	for isLetter(s.ch) || isDecimal(s.ch) || s.ch == '.' {
		s.nextch()
	}
	if s.ch >= utf8.RuneSelf {
		for s.atIdent(false) {
			s.nextch()
		}
	}
	if lit := s.lit(); len(lit) > 2 {
		if tok, ok := keywords[lit]; ok {
			s.tok = tok
			return
		}
	}
	s.tok = _IDENT
}

func (s *scanner) atIdent(first bool) bool {
	switch {
	case unicode.IsLetter(s.ch) || s.ch == '_':
		return true
	case unicode.IsDigit(s.ch) || s.ch == '.':
		if first {
			s.errorf(s.offset, "identifier cannot begin with %#U", s.ch)
		}
	default:
		return false
	}
	return true
}

func (s *scanner) number(seenPoint bool) {
	tok := _INT

	if !seenPoint {
		if s.ch == '0' {
			s.nextch()
			switch lower(s.ch) {
			case 'x':
				s.nextch()
				for isHex(s.ch) {
					s.nextch()
				}
				s.tok = _INT
				return
			}
		}
		for isDecimal(s.ch) {
			s.nextch()
		}
		if s.ch == '.' {
			s.nextch()
			seenPoint = true
		}
	}

	if seenPoint {
		tok = _FLOAT
		for isDecimal(s.ch) {
			s.nextch()
		}
	}

	if lower(s.ch) == 'e' {
		s.nextch()
		tok = _FLOAT
		if s.ch == '+' || s.ch == '-' {
			s.nextch()
		}
		for isDecimal(s.ch) {
			s.nextch()
		}
	}

	s.tok = tok
}

func (s *scanner) string() {
	quote := s.ch
	s.nextch()
	for {
		if s.ch == quote {
			s.nextch()
			break
		}
		if s.ch == '\\' {
			s.nextch()
			if !s.escape(quote) {
				return
			}
			continue
		}
		if s.ch == '\n' {
			s.error(s.offset, "newline in string")
			return
		}
		if s.ch < 0 {
			s.error(s.offset, "string not terminated")
			break
		}
		s.nextch()
	}
	s.tok = _STRING
}

func (s *scanner) escape(quote rune) bool {
	switch s.ch {
	case quote, '\\', 'n', 'r', 't':
		s.nextch()
		return true
	default:
		if s.ch < 0 {
			return true
		}
		s.error(s.offset, "unknown escape")
		return false
	}
}

func (s *scanner) lineComment() {
	s.tok = _COMMENT
	for s.ch >= 0 && s.ch != '\n' {
		s.nextch()
	}
}

func (s *scanner) fullComment() {
	s.tok = _COMMENT
	for s.ch >= 0 {
		for s.ch == '*' {
			s.nextch()
			if s.ch == '/' {
				s.nextch()
				return
			}
		}
		s.nextch()
	}
	s.error(s.offset, "comment not terminated")
}

func lower(ch rune) rune     { return ('a' - 'A') | ch }
func isLetter(ch rune) bool  { return 'a' <= lower(ch) && lower(ch) <= 'z' || ch == '_' }
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }
func isHex(ch rune) bool     { return isDecimal(ch) || 'a' <= lower(ch) && lower(ch) <= 'f' }
