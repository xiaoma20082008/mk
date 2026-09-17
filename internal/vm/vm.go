package vm

import (
	"sync/atomic"

	"mk/internal/oop"
	"mk/internal/runtime"
)

const (
	StackSize   = 2048
	GlobalsSize = 65536
	MaxFrames   = 1024
)

// 编译期生成的全局线程 ID 计数器，保证每个虚拟线程有唯一标识。
var nextThreadID int32

// VM 是 runtime.VM 的具体实现：持有常量池、全局区与入口闭包，
// 并负责创建与驱动线程执行字节码。GC 收集器与堆通过 SetCollector / NewHeap 注入。
type VM struct {
	gc      runtime.GarbageCollector
	threads map[int32]runtime.Thread

	constants []oop.Obj
	globals   []oop.Obj
	entry     *oop.Closure

	result oop.Obj
	err    error

	globalNames map[string]oop.Obj
}

// 编译期断言：*VM 必须满足 runtime.VM 接口。
var _ runtime.VM = (*VM)(nil)

// NewHeap 创建给定容量的托管堆（runtime.Heap）。
func NewHeap(heapSize uint64) runtime.Heap {
	return newHeap(heapSize)
}

// NewVM 创建一台虚拟机实例。
func NewVM() *VM {
	return &VM{
		threads:     make(map[int32]runtime.Thread),
		globals:     make([]oop.Obj, GlobalsSize),
		globalNames: make(map[string]oop.Obj),
	}
}

// Load 注入本轮要执行的入口闭包、常量池与全局区。
func (vm *VM) Load(entry *oop.Closure, constants, globals []oop.Obj) {
	vm.entry = entry
	vm.constants = constants
	if globals == nil {
		vm.globals = make([]oop.Obj, GlobalsSize)
	} else {
		vm.globals = globals
	}
}

// Constants 返回当前常量池（线程执行 OpConstant 时按下标取用）。
func (vm *VM) Constants() []oop.Obj { return vm.constants }

// Globals 返回全局区切片（跨行复用时由调用方取回保存）。
func (vm *VM) Globals() []oop.Obj { return vm.globals }

// Entry 返回入口闭包（顶层脚本被包装成的闭包）。
func (vm *VM) Entry() *oop.Closure { return vm.entry }

// Result 返回最近一次执行的结果与错误。
func (vm *VM) Result() (oop.Obj, error) { return vm.result, vm.err }

// SetResult 由执行线程在执行结束或出错时回填结果。
func (vm *VM) SetResult(obj oop.Obj, err error) {
	vm.result = obj
	vm.err = err
}

// ------------------------------------------------------------------------------------------
// runtime.VM 接口实现
// ------------------------------------------------------------------------------------------

func (vm *VM) GetCollector() runtime.GarbageCollector { return vm.gc }

func (vm *VM) SetCollector(gc runtime.GarbageCollector) {
	vm.gc = gc
	// 若收集器支持反向绑定（拿到 VM 的协调器/根扫描器角色），则把自己注入进去。
	if cfg, ok := gc.(interface{ Bind(runtime.VM) }); ok {
		cfg.Bind(vm)
	}
}

func (vm *VM) RequestSafepoint() {}
func (vm *VM) ReleaseSafepoint() {}
func (vm *VM) Start()            {}
func (vm *VM) StopTheWorld()     {}

func (vm *VM) RegisterThread(t runtime.Thread)   { vm.threads[t.ID()] = t }
func (vm *VM) UnregisterThread(t runtime.Thread) { delete(vm.threads, t.ID()) }
func (vm *VM) GetAllThreads() []runtime.Thread {
	out := make([]runtime.Thread, 0, len(vm.threads))
	for _, t := range vm.threads {
		out = append(out, t)
	}
	return out
}

func (vm *VM) ScanThreadRoots(gc runtime.GarbageCollector) {
	for _, t := range vm.threads {
		if sc, ok := t.(interface {
			ScanRoots(runtime.GarbageCollector)
		}); ok {
			sc.ScanRoots(gc)
		}
	}
}

func (vm *VM) GetGlobalRoots() []oop.Obj {
	roots := make([]oop.Obj, 0, len(vm.globals))
	for _, g := range vm.globals {
		if g != nil {
			roots = append(roots, g)
		}
	}
	return roots
}

func (vm *VM) SetGlobal(name string, val oop.Obj) { vm.globalNames[name] = val }

func (vm *VM) GetGlobal(name string) (oop.Obj, bool) {
	v, ok := vm.globalNames[name]
	return v, ok
}

// NewThread 创建一条绑定到指定虚拟机的解释线程。
func NewThread(v runtime.VM) runtime.Thread {
	concrete, _ := v.(*VM)
	t := &InterpreterThread{
		id:     atomic.AddInt32(&nextThreadID, 1),
		status: runtime.New,
		vm:     concrete,
		stack:  make([]oop.Obj, StackSize),
		frames: make([]*StackFrame, 0, MaxFrames),
		index:  -1,
	}
	return t
}
