package gc

import "mk/internal/oop"

type refCount struct {
}

func (gc *refCount) New(size uint32) oop.Obj { return nil }

func (gc *refCount) Collect() {}

func (gc *refCount) OnReferenceChange(oldObj, newObj oop.Obj) {}
