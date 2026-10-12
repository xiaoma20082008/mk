package runtime

import (
	"fmt"
	"time"
)

type GarbageCollector struct {
	vm   *VM
	heap *Heap

	ygcCount uint32
	ygcTime  int64

	fgcCount uint32
	fgcTime  int64

	threshold uint32
}

func (gc *GarbageCollector) MarkAndRelocate(ptr uintptr) {

}
func (gc *GarbageCollector) MinorGC() {
	fmt.Println("[GC] Minor GC starting...")
	gc.vm.stw = true

	gc.ygcCount++
	start := time.Now()

	for _, thread := range gc.vm.threads {
		thread.ScanRoots(gc)
	}
	fmt.Println("[GC] Scaning root objects in Eden")
	// TODO
	fmt.Println("[GC] Copying S0 => S1")
	// TODO

	fmt.Println("[GC] Reset Eden and S0")
	gc.heap.eden.Clear()
	gc.heap.survivor0.Clear()

	fmt.Println("[GC] Swap S0 and S1")
	gc.heap.survivor0, gc.heap.survivor0 = gc.heap.survivor1, gc.heap.survivor0

	gc.vm.stw = false
	gc.ygcTime = time.Since(start).Milliseconds()
	fmt.Println("[GC] Minor GC finished.")
}

func (gc *GarbageCollector) MajorGC() {
	fmt.Println("[GC] Major GC starting...")
	gc.fgcCount++
	start := time.Now()
	// TODO
	fmt.Println("[GC] Scaning root objects in Metaspace")
	fmt.Println("[GC] Scaning root objects in Old")

	gc.fgcTime = time.Since(start).Milliseconds()
	fmt.Println("[GC] Major GC finished.")
}

func (gc *GarbageCollector) Allocate(size uint32) (uintptr, error) {
	addr := gc.heap.AllocateObj(size)
	if addr != 0 {
		return addr, nil
	}

	// 1. young gc
	gc.MinorGC()

	addr = gc.heap.AllocateObj(size)
	if addr != 0 {
		return addr, nil
	}

	// 2. full gc
	gc.MajorGC()

	addr = gc.heap.AllocateObj(size)
	if addr != 0 {
		return addr, nil
	}
	return 0, fmt.Errorf("OOM: Total size %d bytes", gc.heap.total)
}
