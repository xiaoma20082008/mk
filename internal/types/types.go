package types

var (
	SysInt8  = &TypeDef{Name: "System.Int8", Kind: KindPrimitive, Size: 1}
	SysInt32 = &TypeDef{Name: "System.Int32", Kind: KindPrimitive, Size: 4}
	SysInt64 = &TypeDef{Name: "System.Int64", Kind: KindPrimitive, Size: 8}
	SysBool  = &TypeDef{Name: "System.Bool", Kind: KindPrimitive, Size: 1}
	SysVoid  = &TypeDef{Name: "System.Void", Kind: KindPrimitive, Size: 0}

	SysString   = &TypeDef{Name: "System.String", Kind: KindClass, Size: 0}
	SysObject   = &TypeDef{Name: "System.Object", Kind: KindClass, Size: 0}
	SysFunction = &TypeDef{Name: "System.Function", Kind: KindFunction, Size: 0}
)

var InternalTypes = map[string]*TypeDef{
	"System.Int8":  SysInt8,
	"System.Int32": SysInt32,
	"System.Int64": SysInt32,
	"System.Bool":  SysBool,
	"System.Void":  SysVoid,

	"System.String":   SysString,
	"System.Object":   SysObject,
	"System.Function": SysFunction,

	"byte":   SysInt8,
	"int":    SysInt32,
	"long":   SysInt64,
	"bool":   SysBool,
	"void":   SysVoid,
	"string": SysString,
	"object": SysObject,
	"fn":     SysFunction,
}
