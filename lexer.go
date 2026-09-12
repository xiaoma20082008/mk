package mk

type Lexer struct {
	input string

	pos     int
	readPos int
	ch      byte

	token Token
	saved []Token

	ln  int // 行
	col int // 列
}

func (l *Lexer) NextToken() Token {
	if len(l.saved) > 0 {
		l.token = l.saved[0]
		l.saved = l.saved[1:]
	} else {
		l.token = l.readToken()
	}
	return l.token
}

func (l *Lexer) Token() Token {
	return l.Lookhead(0)
}

func (l *Lexer) Lookhead(lookahead int) Token {
	if lookahead == 0 {
		return l.token
	} else {
		l.ensure(lookahead)
		return l.saved[lookahead-1]
	}
}

func (l *Lexer) readToken() Token {
	var tok Token
	l.skipWhitespace()
	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: EQ, Lit: "=="}
		} else {
			tok = newToken(ASSIGN, l.ch)
		}
	case '+':
		tok = newToken(PLUS, l.ch)
	case '-':
		tok = newToken(MINUS, l.ch)
	case '*':
		tok = newToken(STAR, l.ch)
	case '/':
		tok = newToken(SLASH, l.ch)
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: NE, Lit: "!="}
		} else {
			tok = newToken(BANG, l.ch)
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: LE, Lit: "<="}
		} else if l.peekChar() == '<' {
			l.readChar()
			tok = Token{Type: LTLT, Lit: "<<"}
		} else {
			tok = newToken(LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: GE, Lit: "<="}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = Token{Type: GTGT, Lit: "<<"}
		} else {
			tok = newToken(GT, l.ch)
		}
	case '%':
		tok = newToken(PERCENT, l.ch)
	case '?':
		tok = newToken(QUESTION, l.ch)
	case '~':
		tok = newToken(TILDE, l.ch)
	case ',':
		tok = newToken(COMMA, l.ch)
	case ';':
		tok = newToken(SEMI, l.ch)
	case '{':
		tok = newToken(LBRACE, l.ch)
	case '}':
		tok = newToken(RBRACE, l.ch)
	case '(':
		tok = newToken(LPAREN, l.ch)
	case ')':
		tok = newToken(RPAREN, l.ch)
	case '[':
		tok = newToken(LBRACKET, l.ch)
	case ']':
		tok = newToken(RBRACKET, l.ch)
	case ':':
		tok = newToken(COLON, l.ch)
	case 0:
		tok.Lit = ""
		tok.Type = EOF
	case '"':
		tok.Type = STRING
		tok.Lit = l.readString()
	default:
		if isLetter(l.ch) {
			tok.Lit = l.readIdent()
			tok.Type = lookupIdent(tok.Lit)
			return tok
		} else if isDigit(l.ch) {
			tok.Lit = l.readNum()
			tok.Type = INT
			return tok
		} else {
			tok = newToken(ERR, l.ch)
		}
	}
	l.readChar()
	return tok
}

func (l *Lexer) ensure(lookahead int) {
	for i := len(l.saved); i < lookahead; i++ {
		l.saved = append(l.saved, l.readToken())
	}
}

func (l *Lexer) readString() string {
	p := l.pos

	l.readChar()
	for l.ch != '"' {
		l.readChar()
	}

	return l.input[p+1 : l.pos]
}

func (l *Lexer) readNum() string {
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
		return 0
	} else {
		return l.input[l.readPos]
	}
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++

	if l.ch == '\n' {
		l.ln++
		l.col = 1
	} else {
		l.col++
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\n' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) Position() (ln int, col int) {
	ln = l.ln
	col = l.col
	return ln, col
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: input,

		saved: []Token{},
		token: DUMMY,

		ln:  1,
		col: 1,
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

func lookupIdent(ident string) TokenType {
	if typ, ok := keywords[ident]; ok {
		return typ
	}
	return IDENT
}
