package gc

import "mk/internal/oop"

type GarbageCollector interface {
	New(size uint32) oop.Obj
	Collect()
}

// WriteBarrier 负责监听引用变化
type WriteBarrier interface {
	// 当把 slot 中的 oldObj 替换为 newObj 时触发
	OnReferenceChange(oldObj, newObj oop.Obj)
}

func New() (GarbageCollector, WriteBarrier) {
	rc := new(refCount)
	return rc, rc
}
