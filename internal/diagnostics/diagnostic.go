package diagnostics

import (
	"fmt"
	"strings"
	"sync"
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
	KindLexical  DiagnosticKind = iota // 词法错误（Lexer 阶段）
	KindSyntax                         // 语法错误（Parser 阶段）
	KindSemantic                       // 语义错误（Analyzer 阶段）
	KindRuntime                        // 运行时错误（Evaluator 阶段）
)

type DiagnosticReporter interface {
	Append(filename, source string)
	Report(desc DiagnosticDescriptor, filename string, line int, col int, args ...any)
	Diagnostics() []Diagnostic
}

type DiagnosticBag struct {
	filename    string
	diagnostics []Diagnostic
	mu          sync.Mutex
	sourceCache map[string][]string
}

func (k DiagnosticKind) String() string {
	switch k {
	case KindLexical:
		return "Lexical Error"
	case KindSyntax:
		return "Syntax Error"
	case KindSemantic:
		return "Semantic Error"
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

func (b *DiagnosticBag) Report(desc DiagnosticDescriptor, filename string, line int, col int, args ...any) {
	var srcLine string
	if lines, ok := b.sourceCache[filename]; ok {
		if line > 0 && line <= len(lines) {
			srcLine = lines[line-1]
		}
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

func (b *DiagnosticBag) Append(filename, source string) {
	b.sourceCache[filename] = strings.Split(source, "\n")
}

func NewReporter(filename, source string) DiagnosticReporter {
	return &DiagnosticBag{
		filename:    filename,
		diagnostics: []Diagnostic{},
		sourceCache: map[string][]string{},
	}
}
