package runtime

const (
	New int32 = iota
	Running
	Suspended
	Terminated
)

type Thread interface {
	ID() int32
	Status() int32 // 获取线程状态(Running, Suspended, Ready, Terminated)

	// 执行与调度
	Execute()
	Yield()

	// 安全点机制
	CheckSafepoint()
	Suspend()
	Resume()

	// GC 根集合扫描
	ScanRoots(gc GarbageCollector)
}
