package oop

import (
	"sort"
	"strconv"
	"strings"
)

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

	// 基于栈的虚拟机（internal/runtime）使用的对象类型
	OBJ_COMPILED_FUNC = "COMPILED_FUNC"
	OBJ_CLOSURE       = "CLOSURE"
	OBJ_BUILTIN       = "BUILTIN"
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

type mapEntry struct {
	key Obj
	val Obj
}

type MapObj struct {
	// 注意：key 不能直接用作 map 的键（Obj 是指针），否则两次求值的 1 与 1 会被当成不同的键。
	// 这里统一用「类型 + 字面量」生成可比较的字符串键，并保留原始 key 对象以便打印。
	value map[string]mapEntry
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

// Closure 闭包：函数模板 + 捕获的自由变量
type Closure struct {
	Fn   any
	Free []Obj
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

func (o *StringObj) Len() int {
	return len(o.Value)
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
	items := make([]string, 0, len(o.value))
	for _, v := range o.value {
		items = append(items, v.Inspect())
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func (o *ListObj) Len() int {
	return len(o.value)
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
	pairs := make([]string, 0, len(o.value))
	for _, e := range o.value {
		pairs = append(pairs, e.key.Inspect()+": "+e.val.Inspect())
	}
	sort.Strings(pairs)
	return "{" + strings.Join(pairs, ", ") + "}"
}

func (o *MapObj) Put(k, v Obj) {
	o.value[keyOf(k)] = mapEntry{key: k, val: v}
}

func (o *MapObj) Get(k Obj) Obj {
	if e, ok := o.value[keyOf(k)]; ok {
		return e.val
	}
	return O_NULL
}

func (o *MapObj) Len() int {
	return len(o.value)
}

// keyOf 生成稳定的字符串键，保证「值相等」的键命中同一个槽位
func keyOf(o Obj) string {
	return string(o.Type()) + ":" + o.Inspect()
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
	items := make([]string, 0, len(o.value))
	for _, v := range o.value {
		items = append(items, v.Inspect())
	}
	return "(" + strings.Join(items, ", ") + ",)"
}

func (o *TupleObj) Len() int {
	return len(o.value)
}

func (c *Closure) Type() ObjType {
	return OBJ_CLOSURE
}

func (c *Closure) Inspect() string {
	return "<fn>"
}

func NewClosure(fn any, free []Obj) *Closure {
	return &Closure{Fn: fn, Free: free}
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
		value: map[string]mapEntry{},
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

// IsTruthy 判断对象在布尔上下文中的真假。
// 规则与多数脚本语言一致：nil / null / false 为假，其余皆为真。
// 树遍历解释器与栈式虚拟机共用这一份判定，避免两边语义漂移。
func IsTruthy(o Obj) bool {
	if o == nil {
		return false
	}
	switch v := o.(type) {
	case *NullObj:
		return false
	case *BoolObj:
		return v.Value
	}
	return true
}
