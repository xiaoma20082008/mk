package token

type TokenType string
type Token struct {
	Type   TokenType
	Lit    string
	Line   int
	Column int
	Offset int
}

const (
	ERR = "ERR"
	EOF = "EOF"

	IDENT  = "IDENT"
	INT    = "INT"
	STRING = "STRING"

	// 运算符
	ASSIGN = "="
	PLUS   = "+"
	MINUS  = "-"
	STAR   = "*"
	SLASH  = "/"
	BANG   = "!"
	LT     = "<"
	LE     = "<="
	GT     = ">"
	GE     = ">="
	NE     = "!="
	EQ     = "=="
	TILDE  = "~"
	LTLT   = "<<"
	GTGT   = ">>"

	PERCENT  = "%"
	QUESTION = "?"

	// 分隔符
	COMMA = ","
	SEMI  = ";"
	COLON = ":"

	DOT = "."

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// 关键字
	FUNCTION = "FUNCTION"
	LET      = "LET"
	RETURN   = "RETURN"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	IS       = "IS"
	ELSE     = "ELSE"
	WHILE    = "WHILE"
	NULL     = "NULL"
	SELF     = "SELF"
	CLASS    = "CLASS"
	STRUCT   = "STRUCT"
	NEW      = "NEW"
	TYPEOF   = "TYPEOF"
	SIZEOF   = "SIZEOF"
)

var DUMMY = Token{ERR, "", 0, 0, 0}

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"return": RETURN,
	"true":   TRUE,
	"false":  FALSE,
	"null":   NULL,
	"if":     IF,
	"is":     IS,
	"else":   ELSE,
	"while":  WHILE,
	"class":  CLASS,
	"typeof": TYPEOF,
	"sizeof": SIZEOF,
	"self":   SELF,
	"new":    NEW,
	"struct": STRUCT,
}

var precedences = map[TokenType]int{
	// 赋值运算符 (最低)
	ASSIGN: 1, // =

	// 三元条件运算符
	QUESTION: 2, // ?

	// 逻辑/关系比较
	EQ: 21, // ==
	NE: 22, // !=
	LT: 23, // <
	GT: 24, // >
	LE: 25, // <=
	GE: 26, // >=

	// 算术运算
	PLUS:    31, // +
	MINUS:   32, // -
	SLASH:   33, // /
	STAR:    34, // *
	PERCENT: 35, // % (取模)

	// 函数调用
	LPAREN: 41, // (

	// 数组/索引访问
	LBRACKET: 42, // [
	DOT:      43, // .
}

func GetPrecedence(k TokenType) int {
	if p, ok := precedences[k]; ok {
		return p
	}
	return 0
}

func LookupOpKind(ident string) {
}

func LookupIdentKind(ident string) TokenType {
	if typ, ok := keywords[ident]; ok {
		return typ
	}
	return IDENT
}

func NewToken(typ TokenType, lit string, line, col, pos int) Token {
	return Token{Type: typ, Lit: lit, Line: line, Column: col, Offset: pos}
}
