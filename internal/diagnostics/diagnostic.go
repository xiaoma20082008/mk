package diagnostics

import (
	"fmt"
	"strings"
)

type DiagnosticDescriptor struct {
	Code     string         // 错误码，如 "E0001", "E0002"
	Kind     DiagnosticKind // 词法、语法、运行
	Template string         // 带有占位符的文案模板
}

type Diagnostic struct {
	Descriptor DiagnosticDescriptor
	Kind       DiagnosticKind // 错误类别：词法、语法或运行时
	Filename   string
	Line       int
	Column     int
	Message    string
	Source     string
}

type DiagnosticKind int

const (
	KindLexical DiagnosticKind = iota // 词法错误（Lexer 阶段）
	KindSyntax                        // 语法错误（Parser 阶段）
	KindRuntime                       // 运行时错误（Evaluator 阶段）
)

var (
	// ==========================================
	// 1. 词法错误 (Lexical Errors: E1000 - E1999)
	// ==========================================

	// E1001: 遇到了不属于语言规范的未知字符
	ErrInvalidChar = DiagnosticDescriptor{
		Code:     "E1001",
		Kind:     KindLexical,
		Template: "invalid character '%s'",
	}
	// E1002: 字符串没有闭合，比如 "hello 换行了或者到文件末尾了
	ErrUnterminatedString = DiagnosticDescriptor{
		Code:     "E1002",
		Kind:     KindLexical,
		Template: "unterminated string literal",
	}

	// ==========================================
	// 2. 语法错误 (Syntax Errors: E2000 - E2999)
	// ==========================================

	// E2001: 期待某种符号但落空了，比如 if 后面没有带 {，或者没有带 )
	ErrExpectedToken = DiagnosticDescriptor{
		Code:     "E2001",
		Kind:     KindSyntax,
		Template: "expected '%s', got '%s'",
	}
	// E2002: 强行换行却没有加分号 ;
	ErrMissingSemi = DiagnosticDescriptor{
		Code:     "E2002",
		Kind:     KindSyntax,
		Template: "expected ';', got '%s'",
	}
	// E2003: 表达式单独成行，但它既不是赋值也不是调用（如孤零零的 1 + 2;）
	ErrNotAStmt = DiagnosticDescriptor{
		Code:     "E2003",
		Kind:     KindSyntax,
		Template: "expression '%s' cannot be used as a standalone statement",
	}

	// E2004: 赋值号 = 左边不是一个合法的左值（如 5 = a;）
	ErrInvalidLValue = DiagnosticDescriptor{
		Code:     "E2004",
		Kind:     KindSyntax,
		Template: "cannot assign to '%s'; invalid left-hand side in assignment",
	}
	// E2005: 遇到了无法解析的前缀表达式
	ErrNoPrefixParseFunc = DiagnosticDescriptor{
		Code:     "E2005",
		Kind:     KindSyntax,
		Template: "no prefix parse function for '%s' found",
	}
	// E2006: 遇到了未知的前缀表达式
	ErrUnknownPrefixKind = DiagnosticDescriptor{
		Code:     "E2006",
		Kind:     KindSyntax,
		Template: "unknown prefix type '%s'",
	}
	// E2007: 遇到了无法解析的中缀表达式
	ErrNoInfixParseFunc = DiagnosticDescriptor{
		Code:     "E2007",
		Kind:     KindSyntax,
		Template: "no infix parse function for '%s' found",
	}
	// E2008: 遇到了未知的中缀表达式
	ErrUnknownInfixKind = DiagnosticDescriptor{
		Code:     "E2008",
		Kind:     KindSyntax,
		Template: "unknown infix type '%s'",
	}

	// ==========================================
	// 3. 运行时错误 (Runtime Errors: E3000 - E3999)
	// ==========================================

	// E3001: 数学除以零错误
	ErrDivByZero = DiagnosticDescriptor{
		Code:     "E3001",
		Kind:     KindRuntime,
		Template: "runtime error: division by zero",
	}
	// E3002: 使用了没有定义的变量
	ErrUndefinedIdentifier = DiagnosticDescriptor{
		Code:     "E3002",
		Kind:     KindRuntime,
		Template: "runtime error: undefined variable '%s'",
	}
	// E3003: 对非函数对象尝试进行调用，比如 a = 5; a();
	ErrNotAFunction = DiagnosticDescriptor{
		Code:     "E3003",
		Kind:     KindRuntime,
		Template: "runtime error: '%s' is not a function",
	}
	// E3004: 函数调用时，传入的实参个数和定义的形参个数不匹配
	ErrWrongArgCount = DiagnosticDescriptor{
		Code:     "E3004",
		Kind:     KindRuntime,
		Template: "runtime error: wrong number of arguments; expected %d, got %d",
	}
	// E3005: 数组越界，比如长度为 3 的数组访问了 index = 5
	ErrIndexOutOfBounds = DiagnosticDescriptor{
		Code:     "E3005",
		Kind:     KindRuntime,
		Template: "runtime error: index out of bounds: %d",
	}
)

type DiagnosticReporter interface {
	Report(desc DiagnosticDescriptor, line int, col int, args ...any)
	Diagnostics() []Diagnostic
}

type DiagnosticBag struct {
	filename    string
	sourceLines []string
	diagnostics []Diagnostic
}

func (k DiagnosticKind) String() string {
	switch k {
	case KindLexical:
		return "Lexical Error"
	case KindSyntax:
		return "Syntax Error"
	case KindRuntime:
		return "Runtime Error"
	default:
		return "Error"
	}
}

func (d Diagnostic) String() string {
	var result strings.Builder
	fmt.Fprintf(&result, "[%s] %s: %s\n  --> %s:%d:%d\n",
		d.Descriptor.Kind.String(),
		d.Descriptor.Code,
		d.Message,
		d.Filename,
		d.Line,
		d.Column)
	if d.Source != "" {
		lineStr := fmt.Sprintf("%d", d.Line)
		padding := strings.Repeat(" ", len(lineStr))

		// 第一行：    |
		fmt.Fprintf(&result, " %s |\n", padding)

		// 第二行： 5  |  a = 5
		fmt.Fprintf(&result, " %s | %s\n", lineStr, d.Source)

		// 第三行：    |      ^
		fmt.Fprintf(&result, " %s | ", padding)
		for i := 1; i < d.Column; i++ {
			result.WriteString(" ")
		}
		result.WriteString("^\n")
	}
	return result.String()
}

func (b *DiagnosticBag) Report(desc DiagnosticDescriptor, line int, col int, args ...any) {
	var srcLine string
	if line > 0 && line <= len(b.sourceLines) {
		srcLine = b.sourceLines[line-1]
	}
	msg := fmt.Sprintf(desc.Template, args...)
	b.diagnostics = append(b.diagnostics, Diagnostic{
		Descriptor: desc,
		Filename:   b.filename,
		Line:       line,
		Column:     col,
		Message:    msg,
		Source:     srcLine,
	})
}

func (b *DiagnosticBag) Diagnostics() []Diagnostic {
	return b.diagnostics
}

func NewReporter(filename, source string) DiagnosticReporter {
	return &DiagnosticBag{
		filename:    filename,
		sourceLines: strings.Split(source, "\n"),
		diagnostics: []Diagnostic{},
	}
}
