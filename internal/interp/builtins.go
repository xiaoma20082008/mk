package interp

import (
	"fmt"

	"mk/internal/types"
)

type funtion func(args []any) (*types.TypeDef, any, error)

// builtins 内建函数表
var builtins = map[string]funtion{
	"len":    builtinLen,
	"typeof": builtinTypeOf,
	"sizeof": builtinSizeOf,
}

func builtinLen(args []any) (*types.TypeDef, any, error) {
	if len(args) != 1 {
		return nil, nil, fmt.Errorf("len: wrong number of arguments, want 1, got %d", len(args))
	}
	switch v := args[0].(type) {
	case string:
		return types.SysInt64, int64(len(v)), nil
	case []any:
		return types.SysInt64, int64(len(v)), nil
	case map[any]any:
		return types.SysInt64, int64(len(v)), nil
	case map[string]any:
		return types.SysInt64, int64(len(v)), nil
	default:
		return nil, nil, fmt.Errorf("len: argument of type %T has no length", args[0])
	}
}

func builtinTypeOf(args []any) (*types.TypeDef, any, error) {
	if len(args) != 1 {
		return nil, nil, fmt.Errorf("typeof: wrong number of arguments, want 1, got %d", len(args))
	}
	return types.SysString, typeName(args[0]), nil
}

func builtinSizeOf(args []any) (*types.TypeDef, any, error) {
	if len(args) != 1 {
		return nil, nil, fmt.Errorf("sizeof: wrong number of arguments, want 1, got %d", len(args))
	}
	switch args[0].(type) {
	case nil, bool:
		return types.SysInt64, int64(1), nil
	case int64:
		return types.SysInt64, int64(8), nil
	case string:
		return types.SysInt64, int64(len(args[0].(string))), nil
	default:
		return types.SysInt64, int64(0), nil
	}
}

// typeName 返回运行期值的类型名，尽量与 oop 侧的 ObjType 命名保持一致。
func typeName(v any) string {
	switch v.(type) {
	case nil:
		return "NULL"
	case bool:
		return "BOOL"
	case int64:
		return "INT"
	case string:
		return "STRING"
	case []any:
		return "LIST"
	case map[any]any, map[string]any:
		return "MAP"
	case funtion:
		return "FUNCTION"
	default:
		return fmt.Sprintf("%T", v)
	}
}
