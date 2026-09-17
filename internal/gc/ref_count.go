package gc

import (
	"sync"
	"sync/atomic"

	"mk/internal/oop"
	"mk/internal/runtime"
)

// refCount 是 runtime.GarbageCollector 的一个最小实现（引用计数 + 阈值触发清扫）。
// 在学习版中，真正的对象由 Go 运行时管理，这里主要负责：
//  1. 通过堆做容量记账与 OOM 兜底；
//  2. 在分配量越过阈值时，经协调器请求安全点并执行一次根扫描式的回收。
type refCount struct {
	heap        runtime.Heap
	coordinator runtime.SafepointCoordinator
	scanner     runtime.RootScanner

	gcThreshold uint64
	gcLock      sync.Mutex

	isMarking int32
	grayQueue []oop.Obj
	queueLock sync.Mutex
}

func (gc *refCount) Alloc(size uint32, creator func() oop.Obj) oop.Obj {
	if gc.heap != nil {
		allocated := gc.heap.AllocatedBytes()
		if allocated+uint64(size) > atomic.LoadUint64(&gc.gcThreshold) {
			gc.Trigger()
		}
		slot, err := gc.heap.Allocate(size)
		if err == runtime.ErrOOM {
			gc.Trigger()
			slot, err = gc.heap.Allocate(size)
			if err != nil {
				panic(err)
			}
		}
		_ = slot
	}
	return creator()
}

func (gc *refCount) Collect() {
	// 1. 扫描线程栈，把它们当作 GC 根
	if gc.scanner != nil {
		gc.scanner.ScanThreadRoots(gc)
	}
	// 2. 扫描全局变量根
	if gc.scanner != nil {
		globals := gc.scanner.GetGlobalRoots()
		for _, root := range globals {
			_ = root
		}
	}
	// 3. 执行清理：本学习版对象由 Go 管理，这里仅占位，便于后续接入真实回收。
}

func (gc *refCount) Trigger() {
	if gc.coordinator == nil {
		// 没有协调器时不做世界暂停，直接走一次轻量回收。
		gc.Collect()
		return
	}
	gc.gcLock.Lock()
	defer gc.gcLock.Unlock()

	if gc.heap != nil && gc.heap.AllocatedBytes() <= atomic.LoadUint64(&gc.gcThreshold) {
		return
	}

	gc.coordinator.RequestSafepoint()
	atomic.StoreInt32(&gc.isMarking, 1)

	gc.Collect()
	atomic.StoreInt32(&gc.isMarking, 0)

	if gc.heap != nil {
		nextThreshold := gc.heap.AllocatedBytes() * 2
		atomic.StoreUint64(&gc.gcThreshold, nextThreshold)
	}

	gc.coordinator.ReleaseSafepoint()
}

func (gc *refCount) Update(slot *oop.Obj, newObj oop.Obj) {
	if slot != nil {
		*slot = newObj
	}
}

// Bind 把虚拟机同时作为协调器（SafepointCoordinator）与根扫描器（RootScanner）接入收集器。
// 由 runtime.VM.SetCollector 通过接口断言调用。
func (gc *refCount) Bind(v runtime.VM) {
	gc.coordinator = v
	gc.scanner = v
}

// New 创建一个引用计数收集器并绑定给定堆。
func New(heap runtime.Heap) runtime.GarbageCollector {
	rc := &refCount{
		heap:        heap,
		gcThreshold: 1 << 30, // 保守阈值：学习阶段不会主动触发
	}
	return rc
}

var _ runtime.GarbageCollector = (*refCount)(nil)
