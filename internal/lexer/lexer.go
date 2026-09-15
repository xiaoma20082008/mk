package lexer

import (
	"mk/internal/diagnostics"
	"mk/internal/token"
	"sort"
)

type Lexer interface {
	// NextToken 推进词法流,并返回并消费掉下一个Token
	NextToken() token.Token
	// Token 返回当前已经被消费、Parser 正在处理的那个 Token
	Token() token.Token
	// Lookahead 向前查看第k个Token但不消费它
	Lookahead(k int) token.Token
	// ErrPos 词法错误位置
	ErrPos() int
	// LineMap 将任意绝对字节位置转换为行号和列号（从 1 开始计数）
	LineMap(pos int) (line int, column int)
}

const EOI = 0

type lexerImpl struct {
	input string

	pos     int
	readPos int
	ch      byte

	token token.Token
	saved []token.Token

	lineStarts []int
	errPos     int
	r          diagnostics.DiagnosticReporter
}

func (l *lexerImpl) NextToken() token.Token {
	if len(l.saved) > 0 {
		l.token = l.saved[0]
		l.saved = l.saved[1:]
	} else {
		l.token = l.readToken()
	}
	return l.token
}

func (l *lexerImpl) Token() token.Token {
	return l.Lookahead(0)
}

func (l *lexerImpl) Lookahead(lookahead int) token.Token {
	if lookahead == 0 {
		return l.token
	} else {
		l.ensure(lookahead)
		return l.saved[lookahead-1]
	}
}

func (l *lexerImpl) ErrPos() int {
	return l.errPos
}

func (l *lexerImpl) LineMap(pos int) (int, int) {
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

func (l *lexerImpl) readToken() token.Token {
	l.skipWhitespace()
	var kind token.TokenType
	startPos := l.pos
	lit := ""
	line, column := l.LineMap(startPos)
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
	case '/':
		kind = token.SLASH
		lit = "/"
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
	case 0:
		kind = token.EOF
		lit = ""
	case '"':
		s, ok := l.readString()
		if ok {
			lit = s
			kind = token.STRING
			shouldAdvance = false
		} else {
			lit = ""
			kind = token.ERR
			// 此时 l.ch 已经是 0 (EOI) 了
			shouldAdvance = false
		}
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
			l.r.Report(diagnostics.ErrInvalidChar, line, column, string(l.ch))
			l.errPos = startPos
			kind = token.ERR
			lit = string(l.ch)
		}
	}
	if shouldAdvance {
		l.readChar()
	}
	return token.NewToken(kind, lit, line, column, startPos)
}

func (l *lexerImpl) ensure(lookahead int) {
	for i := len(l.saved); i < lookahead; i++ {
		l.saved = append(l.saved, l.readToken())
	}
}

func (l *lexerImpl) readString() (string, bool) {
	// 1.
	p := l.pos

	// 2.跳过开头的"
	l.readChar()

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
		}
		l.readChar()
	}

	// 3.
	if l.ch == 0 {
		line, col := l.LineMap(p)
		l.r.Report(diagnostics.ErrUnterminatedString, line, col)
		l.errPos = p
		return "", false
	}

	// 4.
	ret := l.input[p+1 : l.pos]

	// 5. 跳过末尾的"
	l.readChar()

	return ret, true
}

func (l *lexerImpl) readInt() string {
	p := l.pos
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[p:l.pos]
}

func (l *lexerImpl) readIdent() string {
	p := l.pos
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[p:l.pos]
}

func (l *lexerImpl) peekChar() byte {
	if l.readPos >= len(l.input) {
		return EOI
	} else {
		return l.input[l.readPos]
	}
}

func (l *lexerImpl) readChar() {
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

func (l *lexerImpl) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\n' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func NewLexer(input string, r diagnostics.DiagnosticReporter) Lexer {
	l := &lexerImpl{
		input: input,

		saved: []token.Token{},
		token: token.DUMMY,

		lineStarts: []int{0},
		errPos:     -1,
		r:          r,
	}
	l.readChar()
	return l
}

func isDigit(ch byte) bool {
	return ('0' <= ch && ch <= '9')
}

func isLetter(ch byte) bool {
	return ch == '_' || ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}
