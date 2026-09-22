package oop

import (
	"fmt"
	"os"
	"strings"
)

// BuiltinFn 内建函数签名：返回 (结果, 错误)
type BuiltinFn func(args ...Obj) (Obj, error)

// Builtin 由宿主语言（Go）实现的函数
type Builtin struct {
	Name string
	Fn   BuiltinFn
}

func (b *Builtin) Type() ObjType {
	return OBJ_BUILTIN
}

func (b *Builtin) Inspect() string {
	return "builtin<" + b.Name + ">"
}

// Builtins 的顺序即内建函数的索引，编译期写入符号表，运行期由 OpGetBuiltin 按下标取用
var Builtins = []*Builtin{
	{Name: "len", Fn: builtinLen},
	{Name: "puts", Fn: builtinPuts},
	{Name: "print", Fn: builtinPrint},
	{Name: "first", Fn: builtinFirst},
	{Name: "last", Fn: builtinLast},
	{Name: "push", Fn: builtinPush},
	{Name: "typeof", Fn: builtinTypeOf},
	{Name: "sizeof", Fn: builtinSizeOf},
}

// LookupBuiltin 按名字查找内建函数及其索引
func LookupBuiltin(name string) (*Builtin, int, bool) {
	for i, b := range Builtins {
		if b.Name == name {
			return b, i, true
		}
	}
	return nil, 0, false
}

func builtinLen(args ...Obj) (Obj, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("len: wrong number of arguments, want 1, got %d", len(args))
	}
	if l, ok := args[0].(interface{ Len() int }); ok {
		return NewInt(int64(l.Len())), nil
	}
	return nil, fmt.Errorf("len: argument of type %s has no length", args[0].Type())
}

func builtinPuts(args ...Obj) (Obj, error) {
	items := make([]string, 0, len(args))
	for _, arg := range args {
		items = append(items, arg.Inspect())
	}
	fmt.Fprintln(os.Stdout, strings.Join(items, " "))
	return O_NULL, nil
}

func builtinPrint(args ...Obj) (Obj, error) {
	items := make([]string, 0, len(args))
	for _, arg := range args {
		items = append(items, arg.Inspect())
	}
	fmt.Fprint(os.Stdout, strings.Join(items, " "))
	return O_NULL, nil
}

func builtinTypeOf(args ...Obj) (Obj, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("type: wrong number of arguments, want 1, got %d", len(args))
	}
	return NewString(string(args[0].Type())), nil
}

func builtinSizeOf(args ...Obj) (Obj, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("type: wrong number of arguments, want 1, got %d", len(args))
	}
	obj := args[0]
	switch obj.(type) {
	case *BoolObj, *NullObj:
		return NewInt(1), nil
	case *IntObj:
		return NewInt(4), nil
	default:
		return NewInt(0), nil
	}
}

func builtinFirst(args ...Obj) (Obj, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("first: wrong number of arguments, want 1, got %d", len(args))
	}
	switch obj := args[0].(type) {
	case *ListObj:
		if obj.Len() == 0 {
			return O_NULL, nil
		}
		return obj.Get(0), nil
	case *TupleObj:
		if obj.Len() == 0 {
			return O_NULL, nil
		}
		return obj.Get(0), nil
	}
	return nil, fmt.Errorf("first: argument of type %s is not a sequence", args[0].Type())
}

func builtinLast(args ...Obj) (Obj, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("last: wrong number of arguments, want 1, got %d", len(args))
	}
	switch obj := args[0].(type) {
	case *ListObj:
		if obj.Len() == 0 {
			return O_NULL, nil
		}
		return obj.Get(obj.Len() - 1), nil
	case *TupleObj:
		if obj.Len() == 0 {
			return O_NULL, nil
		}
		return obj.Get(obj.Len() - 1), nil
	}
	return nil, fmt.Errorf("last: argument of type %s is not a sequence", args[0].Type())
}

// builtinPush 把元素追加到 list 尾部（原地修改），返回同一个 list
func builtinPush(args ...Obj) (Obj, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("push: wrong number of arguments, want 2, got %d", len(args))
	}
	list, ok := args[0].(*ListObj)
	if !ok {
		return nil, fmt.Errorf("push: first argument must be LIST, got %s", args[0].Type())
	}
	list.Add(args[1])
	return list, nil
}
