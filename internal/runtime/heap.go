package runtime

import (
	"log"
	"syscall"
	"unsafe"
)

type Space struct {
	base uint64
	top  uint64
	end  uint64

	size     uint64
	capacity uint64
}

type Heap struct {
	start uintptr
	total uint64

	metaspace *Space
	eden      *Space
	survivor0 *Space
	survivor1 *Space
	old       *Space
}

const (
	HeapMinSize uint64 = 64 * 1024 * 1024
	HeapMaxSize uint64 = 1024 * 1024 * 1024
	PageSize    uint64 = 4096
)

func (s *Space) Allocate(size uint32) uintptr {
	size = uint32((size + 7) &^ 7)
	if s.top+uint64(size) > s.end {
		return 0
	}
	addr := s.top
	s.top += uint64(size)
	return uintptr(addr)
}

func (s *Space) Contains(obj uintptr) bool {
	return uintptr(s.base) <= obj && obj < uintptr(s.end)
}

func (s *Space) Clear() {
	s.top = s.base
}

func (h *Heap) AllocateObj(size uint32) uintptr {
	size = size + uint32(HEADER_SIZE)
	addr := h.eden.Allocate(size)
	if addr == 0 {
		return 0
	}
	header := (*ObjectHeader)(unsafe.Pointer(addr))
	header.MarkWord = 0
	header.TypePtr = 0
	return addr
}

func (h *Heap) AllocateMeta(size uint32) uintptr {
	return h.metaspace.Allocate(size)
}

func NewSpace(base, end uint64) *Space {
	return &Space{
		base:     base,
		top:      base,
		end:      end,
		size:     end - base,
		capacity: end - base,
	}
}

func NewHeap(size uint64) *Heap {
	// 1.
	if size < HeapMinSize {
		size = HeapMinSize
	}
	// 2.
	size = size &^ (PageSize - 1)

	// 3.
	s0Size := size / 32
	s1Size := s0Size
	msSize := s0Size * 2
	oldSize := s0Size * 20
	edenSize := s0Size * 8

	// 4.
	mem, err := syscall.Mmap(-1, 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_ANON|syscall.MAP_PRIVATE)
	if err != nil {
		log.Fatalf("Failed to allocate virtual memory via mmap: %v", err)
	}
	startPtr := uintptr(unsafe.Pointer(&mem[0]))

	// 5.
	offset := uint64(startPtr)
	metaspace := NewSpace(offset, offset+msSize)
	offset += msSize

	eden := NewSpace(offset, offset+edenSize)
	offset += edenSize

	survivor0 := NewSpace(offset, offset+s0Size)
	offset += s0Size

	survivor1 := NewSpace(offset, offset+s1Size)
	offset += s1Size

	old := NewSpace(offset, offset+oldSize)
	offset += oldSize

	return &Heap{
		start: startPtr,
		total: size,

		metaspace: metaspace,
		eden:      eden,
		survivor0: survivor0,
		survivor1: survivor1,
		old:       old,
	}
}
