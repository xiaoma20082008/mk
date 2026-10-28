package vm

import (
	"context"
	"fmt"
	"io"
	"mk/internal/diagnostics"
	"mk/internal/interp"
	"mk/internal/memory"
	"mk/internal/module"
	"mk/internal/parser"
	"mk/internal/sym"
	"mk/pkg/util"
	"sync"
	"sync/atomic"
)

type VMArgs struct {
	CoreLibDir string
	EntryPath  string

	// OS
	PageSize int

	// Memory
	HeapSize  int
	MetaSize  int
	StackSize int

	// GC
	GCType    string
	GCRatio   float32
	GCThreads int

	// JIT
	JITThreshold int
	JITDepth     int

	// Log
	LogLevel int // 0: debug, 1: info, 2: warn, 3: error
	Trace    bool
}

type VM struct {
	args    *VMArgs
	heap    *memory.Heap
	manager module.Manager

	// --- 多线程 ---
	threads map[uint32]*Thread
	wg      sync.WaitGroup

	// --- GC ---
	gcStatus  atomic.Uint32
	safepoint sync.WaitGroup

	// --- 退出 ---
	ctx    context.Context
	cancel context.CancelFunc
}

func (vm *VM) initialize() error {
	if vm.args == nil {
		return fmt.Errorf("vm args is nil")
	}
	if vm.args.CoreLibDir == "" {
		return fmt.Errorf("core lib dir is nil")
	}
	vm.heap = memory.NewHeap(uint64(vm.args.HeapSize))
	vm.threads = make(map[uint32]*Thread)
	vm.manager = module.NewManager()
	vm.ctx, vm.cancel = context.WithCancel(context.Background())
	vm.gcStatus.Store(0)
	return nil
}

func (vm *VM) bootstrap() error {
	if _, err := vm.manager.Bootstrap(vm.args.CoreLibDir); err != nil {
		return err
	}
	entryMod, err := vm.manager.Bootstrap(vm.args.EntryPath)
	if err != nil {
		return err
	}
	if err := vm.manager.Link(entryMod); err != nil {
		return fmt.Errorf("failed to link entry module: %w", err)
	}
	if err := vm.manager.Register(entryMod); err != nil {
		return fmt.Errorf("failed to register entry module: %w", err)
	}
	if entryMod.EntryPoint == nil {
		return fmt.Errorf("entry module '%s' does not have an EntryPoint (Main function)", entryMod.Name)
	}
	mainThread := NewThread(vm)
	mainThread.SetName("main")
	mainThread.current = newFrame(entryMod)
	return mainThread.Execute()
}

func (vm *VM) InterpRun(code string, w io.Writer) error {
	ctx := util.NewContext()
	global := sym.NewRootTable(ctx)
	r := diagnostics.NewReporter("", code)
	l := parser.NewLexer(r)
	l.Reset("", code)
	p := parser.NewParser(l, r)
	ast := p.ParseCode()
	e := interp.NewEvaluator(global, r)
	e.EvalWithWriter(ast, w)
	return nil
}

func (vm *VM) CompileRun(r io.Reader, w io.Writer) error {
	return nil
}

func NewVMArgs() *VMArgs {
	return &VMArgs{
		PageSize: 4 * 1024, // 4KB

		CoreLibDir: "",

		HeapSize:  64 * 1024 * 1024, // 64MB
		MetaSize:  16 * 1024 * 1024, // 16MB
		StackSize: 64 * 1024,        // 64KB

		JITThreshold: 1000,
		JITDepth:     5,

		GCType:    "cms",
		GCRatio:   0.75,
		GCThreads: 4,

		LogLevel: 0,
		Trace:    true,
	}
}

func New() *VM {
	return &VM{
		args: NewVMArgs(),
	}
}
