package vm

import (
	"mk/internal/oop"
	"mk/internal/runtime"
)

// slotImpl 是 runtime.RawSlot 的占位实现。
// 在本学习版中，真正的对象由 Go 运行时分配（oop.Obj），堆只做容量记账，
// 因此 slot 仅需一个地址与一段（象征性的）字节视图。
type slotImpl struct {
	addr uintptr
	size uint32
}

func (s *slotImpl) Addr() uintptr { return s.addr }
func (s *slotImpl) Bytes() []byte {
	if s.size == 0 {
		return []byte{}
	}
	return make([]byte, s.size)
}

// heapImpl 实现了 runtime.Heap 接口：负责容量统计与 slot 登记。
type heapImpl struct {
	total     uint64
	allocated uint64
	objects   []oop.RawSlot
}

func newHeap(size uint64) *heapImpl {
	if size == 0 {
		size = 1 << 20 // 默认 1MB
	}
	return &heapImpl{total: size}
}

func (h *heapImpl) Allocate(size uint32) (oop.RawSlot, error) {
	if h.allocated+uint64(size) > h.total {
		return nil, runtime.ErrOOM
	}
	slot := &slotImpl{addr: uintptr(len(h.objects) + 1), size: size}
	h.objects = append(h.objects, slot)
	h.allocated += uint64(size)
	return slot, nil
}

func (h *heapImpl) Free(slot oop.RawSlot) {
	if s, ok := slot.(*slotImpl); ok {
		h.allocated -= uint64(s.size)
	}
}

func (h *heapImpl) TotalBytes() uint64     { return h.total }
func (h *heapImpl) AllocatedBytes() uint64 { return h.allocated }
func (h *heapImpl) AvailableBytes() uint64 { return h.total - h.allocated }

func (h *heapImpl) ResetCounters() { h.allocated = 0 }

func (h *heapImpl) Contains(slot oop.RawSlot) bool {
	for _, s := range h.objects {
		if s == slot {
			return true
		}
	}
	return false
}

func (h *heapImpl) VisitObjects(visitor func(slot oop.RawSlot) bool) {
	for _, s := range h.objects {
		if !visitor(s) {
			return
		}
	}
}

var _ runtime.Heap = (*heapImpl)(nil)
