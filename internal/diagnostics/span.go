package diagnostics

type Span struct {
	File string

	StartLine, StartCol    int
	EndLine, EndCol        int
	StartOffset, EndOffset int
}
