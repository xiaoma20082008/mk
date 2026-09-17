package runtime

import "mk/internal/oop"

type VM interface {
	SafepointCoordinator
	RootScanner

	GetCollector() GarbageCollector
	SetCollector(GarbageCollector)

	Start()
	StopTheWorld()

	// 线程/协程生命周期管理
	RegisterThread(t Thread)
	UnregisterThread(t Thread)
	GetAllThreads() []Thread

	SetGlobal(name string, val oop.Obj)
	GetGlobal(name string) (oop.Obj, bool)
}
