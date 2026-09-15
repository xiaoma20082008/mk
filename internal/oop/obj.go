package oop

import "strconv"

type ObjType string

const (
	OBJ_INT    = "INT"
	OBJ_BOOL   = "BOOL"
	OBJ_NULL   = "NULL"
	OBJ_STRING = "STRING"
	OBJ_MAP    = "MAP"
	OBJ_LIST   = "LIST"
	OBJ_TUPLE  = "TUPLE"
	OBJ_FUNC   = "FUNC"
	OBJ_CLASS  = "CLASS"
	OBJ_TSRUCT = "STRUCT"
)

var (
	O_NULL  = &NullObj{}
	O_TRUE  = &BoolObj{Value: true}
	O_FALSE = &BoolObj{Value: false}
)

type Obj interface {
	Type() ObjType
	Inspect() string
}

type IntObj struct {
	Value int64
}

type BoolObj struct {
	Value bool
}

type NullObj struct{}

type StringObj struct {
	Value string
}

type MapObj struct {
	value map[Obj]Obj
}

type ListObj struct {
	value []Obj
}

type TupleObj struct {
	value []Obj
}

type FuncObj struct {
	Name string
	Args []Obj
	Body []Obj
	Call func(args []Obj) (Obj, bool)
}

type ClassObj struct{}

type StructObj struct{}

// ------------------------------------------------------------------------------------------

func (o *IntObj) Type() ObjType {
	return OBJ_INT
}

func (o *IntObj) Inspect() string {
	return strconv.FormatInt(o.Value, 10)
}

func (o *BoolObj) Type() ObjType {
	return OBJ_BOOL
}

func (o *BoolObj) Inspect() string {
	if o.Value {
		return "true"
	}
	return "false"
}

func (o *NullObj) Type() ObjType {
	return OBJ_NULL
}

func (o *NullObj) Inspect() string {
	return "null"
}

func (o *StringObj) Type() ObjType {
	return OBJ_STRING
}

func (o *StringObj) Inspect() string {
	return o.Value
}

func (o *FuncObj) Type() ObjType {
	return OBJ_FUNC
}

func (o *FuncObj) Inspect() string {
	return o.Name
}

func (o *ListObj) Type() ObjType {
	return OBJ_LIST
}

func (o *ListObj) Inspect() string {
	return "[]"
}

func (o *ListObj) Set(k int, v Obj) Obj {
	if k >= len(o.value) {
		return O_NULL
	}
	ov := o.value[k]
	o.value[k] = v
	return ov
}

func (o *ListObj) Get(k int) Obj {
	if k >= len(o.value) {
		return O_NULL
	}
	return o.value[k]
}

func (o *ListObj) Add(val Obj) {
	o.value = append(o.value, val)
}

func (o *MapObj) Type() ObjType {
	return OBJ_MAP
}

func (o *MapObj) Inspect() string {
	return "{}"
}

func (o *MapObj) Put(k, v Obj) {
	o.value[k] = v
}

func (o *MapObj) Get(k Obj) Obj {
	if v, ok := o.value[k]; ok {
		return v
	}
	return O_NULL
}

func (o *TupleObj) Set(k int, v Obj) Obj {
	if k >= len(o.value) {
		return O_NULL
	}
	ov := o.value[k]
	o.value[k] = v
	return ov
}

func (o *TupleObj) Get(k int) Obj {
	if k >= len(o.value) {
		return O_NULL
	}
	return o.value[k]
}

func (o *TupleObj) Type() ObjType {
	return OBJ_TUPLE
}

func (o *TupleObj) Inspect() string {
	return "()"
}

func NewList() *ListObj {
	return &ListObj{
		value: []Obj{},
	}
}

func NewTuple(values ...Obj) *TupleObj {
	o := &TupleObj{
		value: []Obj{},
	}
	for _, val := range values {
		o.value = append(o.value, val)
	}
	return o
}

func NewMap() *MapObj {
	return &MapObj{
		value: map[Obj]Obj{},
	}
}

func NewInt(v int64) *IntObj {
	return &IntObj{Value: v}
}

func NewBool(v bool) *BoolObj {
	if v {
		return O_TRUE
	}
	return O_FALSE
}

func NewString(v string) *StringObj {
	return &StringObj{Value: v}
}
