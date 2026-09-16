package runtime

import "mk/internal/gc"

// Heap 语言运行期的堆。真正的分配与回收策略由 gc 包提供，
// 目前 gc.New() 返回的是引用计数的空实现，等 GC 落地后这里无需改动。
type Heap struct {
	collector gc.GarbageCollector
	barrier   gc.WriteBarrier
}

func NewHeap() *Heap {
	collector, barrier := gc.New()
	return &Heap{
		collector: collector,
		barrier:   barrier,
	}
}

// Collector 返回堆使用的垃圾回收器
func (h *Heap) Collector() gc.GarbageCollector { return h.collector }

// Barrier 返回写屏障，供赋值路径在引用变化时通知回收器
func (h *Heap) Barrier() gc.WriteBarrier { return h.barrier }
