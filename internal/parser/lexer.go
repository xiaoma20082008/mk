package parser

import (
	"mk/internal/diagnostics"
	"mk/internal/token"
	"sort"
	"strings"
)

const EOI = 0

type Lexer struct {
	filename string
	input    string

	pos     int
	readPos int
	ch      byte

	token token.Token
	saved []token.Token

	lineStarts []int
	errors     diagnostics.DiagnosticReporter
}

func (l *Lexer) NextToken() token.Token {
	if len(l.saved) > 0 {
		l.token = l.saved[0]
		l.saved = l.saved[1:]
	} else {
		l.token = l.readToken()
	}
	return l.token
}

func (l *Lexer) Token() token.Token {
	return l.Lookahead(0)
}

func (l *Lexer) Lookahead(lookahead int) token.Token {
	if lookahead == 0 {
		return l.token
	} else {
		l.ensure(lookahead)
		return l.saved[lookahead-1]
	}
}

func (l *Lexer) LineMap(pos int) (int, int) {
	if pos <= 0 {
		return 1, 1
	}
	if pos > len(l.input) {
		pos = len(l.input)
	}
	line := sort.Search(len(l.lineStarts), func(i int) bool {
		return l.lineStarts[i] > pos
	})
	startOfLine := l.lineStarts[line-1]
	column := pos - startOfLine + 1
	return line, column
}

func (l *Lexer) File() string {
	return l.filename
}

func (l *Lexer) Reset(filename, source string) {
	l.filename = filename
	l.input = source

	l.saved = []token.Token{}
	l.token = token.DUMMY

	l.pos = 0
	l.readPos = 0
	l.ch = 0

	l.lineStarts = []int{0}

	l.readChar()
}

func (l *Lexer) readToken() token.Token {
	l.skipWhitespace()
	var kind token.TokenType
	startPos := l.pos
	lit := ""
	line, column := 0, 0
	shouldAdvance := true
	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			kind = token.EQ
			lit = "=="
		} else {
			kind = token.ASSIGN
			lit = "="
		}
	case '+':
		kind = token.PLUS
		lit = "+"
	case '-':
		kind = token.MINUS
		lit = "-"
	case '*':
		kind = token.STAR
		lit = "*"
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			kind = token.LOG_AND
			lit = "&&"
		} else {
			kind = token.BIT_AND
			lit = "&"
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			kind = token.LOG_OR
			lit = "||"
		} else {
			kind = token.BIT_OR
			lit = "|"
		}
	case '^':
		kind = token.BIT_XOR
		lit = "^"
	case '/':
		pc := l.peekChar()
		switch pc {
		case '/':
			comment, ok := l.readComment(false)
			if !ok {
				lit = ""
				kind = token.ERR
			} else {
				lit = comment
				kind = token.COMMENT
			}
		case '*':
			comment, ok := l.readComment(true)
			if !ok {
				lit = ""
				kind = token.ERR
			} else {
				lit = comment
				kind = token.COMMENT
			}
		default:
			kind = token.SLASH
			lit = "/"
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			kind = token.NE
			lit = "!="
		} else {
			kind = token.BANG
			lit = "!"
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			kind = token.LE
			lit = "<="
		} else if l.peekChar() == '<' {
			l.readChar()
			kind = token.LTLT
			lit = "<<"
		} else {
			kind = token.LT
			lit = "<"
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			kind = token.GE
			lit = ">="
		} else if l.peekChar() == '>' {
			l.readChar()
			kind = token.GTGT
			lit = ">>"
		} else {
			kind = token.GT
			lit = ">"
		}
	case '%':
		kind = token.PERCENT
		lit = "%"
	case '?':
		kind = token.QUESTION
		lit = "?"
	case '~':
		kind = token.TILDE
		lit = "~"
	case ',':
		kind = token.COMMA
		lit = ","
	case ';':
		kind = token.SEMI
		lit = ";"
	case '{':
		kind = token.LBRACE
		lit = "{"
	case '}':
		kind = token.RBRACE
		lit = "}"
	case '(':
		kind = token.LPAREN
		lit = "("
	case ')':
		kind = token.RPAREN
		lit = ")"
	case '[':
		kind = token.LBRACKET
		lit = "["
	case ']':
		kind = token.RBRACKET
		lit = "]"
	case ':':
		kind = token.COLON
		lit = ":"
	case '$':
		kind = token.DOLLAR
		lit = "$"
	case '.':
		kind = token.DOT
		lit = "."
	case 0:
		kind = token.EOF
		lit = ""
	case '"':
		s, ok := l.readString()
		if ok {
			lit = s
			kind = token.STRING
		} else {
			lit = ""
			kind = token.ERR
			// 此时 l.ch 已经是 0 (EOI) 了
		}
		shouldAdvance = false
	default:
		if isLetter(l.ch) {
			lit = l.readIdent()
			kind = token.LookupIdentKind(lit)
			shouldAdvance = false
		} else if isDigit(l.ch) {
			kind = token.INT
			lit = l.readInt()
			shouldAdvance = false
		} else {
			line, column = l.LineMap(startPos)
			l.errors.Report(diagnostics.ErrInvalidChar, l.filename, line, column, string(l.ch))
			kind = token.ERR
			lit = string(l.ch)
		}
	}
	if shouldAdvance {
		l.readChar()
	}
	if line == 0 && column == 0 {
		line, column = l.LineMap(startPos)
	}
	return token.NewToken(kind, lit, line, column, startPos)
}

func (l *Lexer) ensure(lookahead int) {
	for i := len(l.saved); i < lookahead; i++ {
		l.saved = append(l.saved, l.readToken())
	}
}

func (l *Lexer) readString() (string, bool) {
	p := l.pos

	// 跳过开头的"
	l.readChar()

	var buf strings.Builder

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\n' {
			line, col := l.LineMap(p)
			l.errors.Report(diagnostics.ErrUnterminatedString, l.filename, line, col)
			return "", false
		}
		if l.ch != '\\' {
			buf.WriteByte(l.ch)
			l.readChar()
			continue
		}
		bslashPos := l.pos
		l.readChar()

		if l.ch == EOI {
			break
		}
		switch l.ch {
		case 'r':
			buf.WriteByte('\r')
		case 'n':
			buf.WriteByte('\n')
		case 't':
			buf.WriteByte('\t')
		case 'f':
			buf.WriteByte('\f')
		case '"':
			buf.WriteByte('"')
		case '0':
			buf.WriteByte(0)
		case '\\':
			buf.WriteByte('\\')
		case '\'':
			buf.WriteByte('\'')
		default:
			line, col := l.LineMap(bslashPos)
			l.errors.Report(diagnostics.ErrInvalidEscape, l.filename, line, col, string(l.ch))
			buf.WriteByte('\\')
			buf.WriteByte(l.ch)
		}
		l.readChar()
	}

	if l.ch == EOI {
		line, col := l.LineMap(p)
		l.errors.Report(diagnostics.ErrUnterminatedString, l.filename, line, col)
		return "", false
	}

	ret := buf.String()

	// 跳过末尾的"
	l.readChar()

	return ret, true
}

func (l *Lexer) readComment(multi bool) (string, bool) {
	l.readChar() // 跳过 / 或 *
	p := l.pos

	for l.ch != EOI {
		if multi && l.ch == '*' && l.peekChar() == '/' {
			l.readChar()
			break
		} else if !multi && l.ch == '\n' {
			break
		}
		l.readChar()
	}

	if multi && l.ch == EOI {
		line, col := l.LineMap(p)
		l.errors.Report(diagnostics.ErrUnterminatedComment, l.filename, line, col)
		return "", false
	}
	ret := l.input[p:l.pos]

	if multi { // 读取下一个
		l.readChar()
	}
	return ret, true
}

func (l *Lexer) readInt() string {
	p := l.pos
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[p:l.pos]
}

func (l *Lexer) readIdent() string {
	p := l.pos
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[p:l.pos]
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return EOI
	} else {
		return l.input[l.readPos]
	}
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = EOI
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++

	if l.ch == '\n' {
		l.lineStarts = append(l.lineStarts, l.pos+1)
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\n' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func NewLexer(r diagnostics.DiagnosticReporter) *Lexer {
	l := &Lexer{
		errors: r,
		saved:  []token.Token{},
	}
	return l
}

func isDigit(ch byte) bool {
	return ('0' <= ch && ch <= '9')
}

func isLetter(ch byte) bool {
	return ch == '_' || ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}
