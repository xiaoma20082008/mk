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

	BIT_AND = "&"
	BIT_OR  = "|"
	BIT_XOR = "^"

	LOG_AND = "&&"
	LOG_OR  = "||"

	COMMENT

	// 请求上下文
	DOLLAR = "$"

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
	"self":   SELF,
	"new":    NEW,
	"struct": STRUCT,
}

var PREC_edences = map[TokenType]int{
	// 赋值运算符 (最低)
	ASSIGN: PREC_ASSIGN, // =

	// 三元条件运算符
	QUESTION: PREC_QUESTION, // ?

	// 逻辑
	LOG_OR:  PREC_LOG_OR,
	LOG_AND: PREC_LOG_AND,
	// OR:      PREC_LOG_OR,
	// AND:     PREC_LOG_AND,

	// 关系比较
	EQ: PREC_EQUALS,       // ==
	NE: PREC_EQUALS,       // !=
	LT: PREC_LESS_GREATER, // <
	GT: PREC_LESS_GREATER, // >
	LE: PREC_LESS_GREATER, // <=
	GE: PREC_LESS_GREATER, // >=

	// 算术运算
	BIT_OR:  PREC_BIT_OR,
	BIT_XOR: PREC_BIT_XOR,
	BIT_AND: PREC_BIT_AND,
	LTLT:    PREC_SHIFT,   // <<
	GTGT:    PREC_SHIFT,   // >>
	PLUS:    PREC_SUM,     // +
	MINUS:   PREC_SUM,     // -
	SLASH:   PREC_PRODUCT, // /
	STAR:    PREC_PRODUCT, // *
	PERCENT: PREC_PRODUCT, // % (取模)

	// 函数调用
	LPAREN: PREC_CALL, // (
	// 数组/索引访问
	LBRACKET: PREC_CALL, // [
	DOT:      PREC_CALL, // .
}

const (
	_ int = iota
	PREC_LOWEST
	PREC_ASSIGN       // =
	PREC_QUESTION     // ?
	PREC_LOG_OR       // ||
	PREC_LOG_AND      // &&
	PREC_EQUALS       // ==, !=
	PREC_LESS_GREATER // >, <, >=, <=
	PREC_BIT_OR       // |
	PREC_BIT_XOR      // ^
	PREC_BIT_AND      // &
	PREC_SHIFT        // <<, >>
	PREC_SUM          // +, -
	PREC_PRODUCT      // *, /, %
	PREC_PREFIX       // 一元前缀: +, -, !, ~
	PREC_CALL         // 方法/函数调用: .(, (
	PREC_HIGHEST
)

func GetPREC_edence(k TokenType) int {
	if p, ok := PREC_edences[k]; ok {
		return p
	}
	return 0
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
