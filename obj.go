package mk

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
	Value map[any]any
}

type ListObj struct {
	Value []any
}

type TupleObj struct {
	Value []any
}

type FuncObj struct {
	Name string
	Args []any
	Body []any
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
