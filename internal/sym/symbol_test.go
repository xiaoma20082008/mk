package sym_test

import (
	"mk/internal/sym"
	"mk/internal/types"
	"mk/pkg/util"
	"testing"
)

func TestSymbolTable_EnterAndLeave(t *testing.T) {
	ctx := util.NewContext()
	root := sym.NewRootTable(ctx)
	if root.Name() != "<global>" {
		t.Errorf("root name = %q, want <global>", root.Name())
	}
	if root.Depth() != 0 {
		t.Errorf("root depth = %d, want 0", root.Depth())
	}

	// Enter block1
	root.Enter("block1")
	if root.Name() != "block1" {
		t.Errorf("after Enter name = %q, want block1", root.Name())
	}
	if root.Depth() != 1 {
		t.Errorf("after Enter depth = %d, want 1", root.Depth())
	}

	// Enter if
	root.Enter("if")
	if root.Depth() != 2 {
		t.Errorf("nested Enter depth = %d, want 2", root.Depth())
	}

	// Leave if
	root.Leave()
	if root.Name() != "block1" || root.Depth() != 1 {
		t.Errorf("after Leave name=%q depth=%d, want block1/1", root.Name(), root.Depth())
	}

	// Leave block1
	root.Leave()
	if root.Name() != "<global>" || root.Depth() != 0 {
		t.Errorf("back to root name=%q depth=%d, want <global>/0", root.Name(), root.Depth())
	}

	// 在根上 Leave 是安全的 no-op
	root.Leave()
	if root.Depth() != 0 {
		t.Errorf("Leave at root should be no-op, got depth=%d", root.Depth())
	}
}

func TestSymbolTable_DefineAndResolve(t *testing.T) {
	ctx := util.NewContext()
	root := sym.NewRootTable(ctx)

	a, err := root.Define("a", sym.KindVariable, types.SysInt32)
	if err != nil {
		t.Fatalf("define a failed: %v", err)
	}
	if a.Index != 0 {
		t.Errorf("expected a.Index=0, got %d", a.Index)
	}

	// 重复定义应当报错
	if _, err := root.Define("a", sym.KindVariable, types.SysInt32); err == nil {
		t.Fatal("redefine 'a' should fail")
	}

	// Enter block1
	root.Enter("block1")
	b, err := root.Define("b", sym.KindVariable, types.SysInt32)
	if err != nil {
		t.Fatalf("define b failed: %v", err)
	}
	if b.Depth != 1 {
		t.Errorf("expected b.Depth=1, got %d", b.Depth)
	}

	// 子作用域能看到外层 a
	if a2, ok := root.Resolve("a"); !ok || a2 != a {
		t.Fatal("child scope should resolve 'a' from parent")
	}

	// 在子作用域遮蔽 a
	aShadow, err := root.Define("a", sym.KindVariable, types.SysBool)
	if err != nil {
		t.Fatalf("shadow 'a' in child should be allowed: %v", err)
	}
	if got, _ := root.Resolve("a"); got != aShadow {
		t.Fatal("child should resolve to shadowed 'a'")
	}

	// Leave block1
	root.Leave()
	if got, _ := root.Resolve("a"); got != a {
		t.Fatal("root should still resolve to original 'a' after Leave")
	}
	if _, ok := root.Resolve("b"); ok {
		t.Fatal("root scope should NOT see 'b' defined in left child")
	}
}
