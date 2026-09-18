package runtime

import (
	"errors"
	"mk/internal/oop"
)

var (
	ErrOOM           = errors.New("out of memory: heap exhausted")
	ErrInvalidObject = errors.New("invalid object pointer or address")
)

type Heap interface {
	// Allocate 分配指定大小（字节）的连续物理空间。
	Allocate(size uint64) (oop.RawSlot, error)

	// Free 释放分配的Slot。
	Free(slot oop.RawSlot)

	TotalBytes() uint64     // 堆的总物理限制容量（单位：字节）
	AllocatedBytes() uint64 // 堆当前已分配并被占用的容量（单位：字节）
	AvailableBytes() uint64 // 堆当前还剩余的可用容量（单位：字节）
	ResetCounters()         // 重置统计计数器（通常在一次完全 GC 结束后调用）

	// Contains 判定 Slot 是否属于当前堆的管辖范围。
	Contains(slot oop.RawSlot) bool

	// VisitObjects 遍历堆中的所有对象。
	VisitObjects(visitor func(slot oop.RawSlot) bool)
}
