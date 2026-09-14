package mk

type Analyzer interface {
	Resolve(p *Program)
}

type analyzer struct {
	r DiagnosticReporter
}
