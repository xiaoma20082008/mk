package vm

import (
	"mk/internal/oop"
	"mk/internal/runtime"
	"unsafe"
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

	// TODO 使用 FreeList管理
	raw  []byte
	head *block
}

type block struct {
	Size uint64
	Free bool
	Next *block
}

func newHeap(size uint64) *heapImpl {
	if size == 0 {
		size = 1 << 20 // 默认 1MB
	}
	raw := make([]byte, size)

	head := (*block)(unsafe.Pointer(&raw))
	head.Size = size - uint64(unsafe.Sizeof(block{}))
	head.Free = true
	head.Next = nil

	return &heapImpl{
		raw:   raw,
		head:  head,
		total: size,
		//
		allocated: 0,
		objects:   make([]oop.RawSlot, 0),
	}
}

func (h *heapImpl) Allocate(size uint64) (oop.RawSlot, error) {
	if size == 0 {
		return &slotImpl{}, nil
	}
	// 内存对齐
	size = (size + 7) &^ 7
	if h.allocated+uint64(size) > h.total {
		return nil, runtime.ErrOOM
	}
	for curr := h.head; curr != nil; {
		if curr.Free && curr.Size >= size {
			minSplitSize := unsafe.Sizeof(block{}) + 8
			if curr.Size-size >= uint64(minSplitSize) {
				// 计算新的block地址
				currAddr := uintptr(unsafe.Pointer(curr))
				nextAddr := currAddr + unsafe.Sizeof(block{}) + uintptr(size)

				// 创建新的block
				nextBlock := (*block)(unsafe.Pointer(nextAddr))
				nextBlock.Size = curr.Size - size - uint64(unsafe.Sizeof(block{}))
				nextBlock.Free = true
				nextBlock.Next = curr

				// 更新当前block
				curr.Size = size
				curr.Next = nextBlock
			}
			curr.Free = false

			dataAddr := uintptr(unsafe.Pointer(curr)) + unsafe.Sizeof(block{})

			slot := &slotImpl{
				addr: dataAddr,
				size: uint32(size),
			}

			h.allocated += uint64(size)
			h.objects = append(h.objects, slot)

			return slot, nil
		}
	}
	return &slotImpl{}, runtime.ErrOOM
}

func (h *heapImpl) Free(slot oop.RawSlot) {
	if slot.Addr() == 0 {
		return
	}
	// 1.
	blockAddr := slot.Addr() - unsafe.Sizeof(block{})
	block := (*block)(unsafe.Pointer(blockAddr))
	if block.Free {
		return
	}
	// 2.
	h.allocated -= uint64(block.Size)

	// 3.
	for i, obj := range h.objects {
		if obj.Addr() == slot.Addr() {
			h.objects = append(h.objects[:i], h.objects[i+1:]...)
			break
		}
	}

	// 4.
	block.Free = true
	h.coalesce()
}

// coalesce 合并相连的内存块
func (h *heapImpl) coalesce() {
	for curr := h.head; curr != nil && curr.Next != nil; {
		if curr.Free && curr.Next.Free {
			curr.Size += uint64(unsafe.Sizeof(block{})) + curr.Next.Size
			curr.Next = curr.Next.Next
		} else {
			curr = curr.Next
		}
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
