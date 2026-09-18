package runtime

import "mk/internal/oop"

type SafepointCoordinator interface {
	// RequestSafepoint 触发全局安全点请求，挂起所有正在运行的虚拟线程
	RequestSafepoint()

	// ReleaseSafepoint 解除安全点，恢复所有虚拟线程的执行
	ReleaseSafepoint()
}

type GarbageCollector interface {
	// Alloc 创建对象
	Alloc(size uint32, creator func() oop.Obj) oop.Obj
	// Trigger 触发回收
	Trigger()
	// Collect 执行垃圾收集
	Collect()
	// Update 更新引用
	Update(slot *oop.Obj, newObj oop.Obj)
}

type RootScanner interface {
	ScanThreadRoots(gc GarbageCollector)

	GetGlobalRoots() []oop.Obj
}
