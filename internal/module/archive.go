package module

import (
	"io"
	"mk/internal/types"
)

type Manifest struct {
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	EntryPointIndex int      `json:"entry_point_index"` // 入口方法在元数据中的索引
	Dependencies    []string `json:"dependencies"`      // 依赖的其他模块名
}

type archiver struct {
}

func (a *archiver) deserialize(r io.Reader) (*types.ModuleDef, error) {
	return nil, nil
}

func (a *archiver) serialize(mod *types.ModuleDef, w io.Writer) error {
	return nil
}
