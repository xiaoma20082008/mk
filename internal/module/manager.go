package module

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mk/internal/types"
	"os"
	"sync"
)

type Manager interface {

	// 1. 加载

	LoadFromFile(file string) (*types.ModuleDef, error)
	LoadFromBytes(bytes []byte) (*types.ModuleDef, error)
	LoadFromStream(r io.Reader) (*types.ModuleDef, error)

	// 2. 链接

	Link(mod *types.ModuleDef) error

	// 3. 注册

	Register(mod *types.ModuleDef) error

	// 4.寻址

	LookupModule(name string) (*types.ModuleDef, error)
	LookupType(moduleName, typeName string) (*types.TypeDef, error)

	//
	Bootstrap(path string) (*types.ModuleDef, error)
}

type managerImpl struct {
	modules map[string]*types.ModuleDef

	linker   linker
	archiver archiver

	locker sync.Mutex
}

func (m *managerImpl) LoadFromFile(file string) (*types.ModuleDef, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open module file: %w", err)
	}
	defer f.Close()
	return m.LoadFromStream(bufio.NewReaderSize(f, 1024))
}

func (m *managerImpl) LoadFromBytes(b []byte) (*types.ModuleDef, error) {
	return m.LoadFromStream(bufio.NewReaderSize(bytes.NewReader(b), 1024))
}

func (m *managerImpl) LoadFromStream(r io.Reader) (*types.ModuleDef, error) {
	mod, err := m.archiver.deserialize(r)
	if err != nil {
		return nil, fmt.Errorf("load mod failed: %w", err)
	}
	return mod, nil
}

func (m *managerImpl) Link(mod *types.ModuleDef) error {
	m.locker.Lock()
	defer m.locker.Unlock()
	if err := m.linker.link(mod); err != nil {
		return fmt.Errorf("link module %s failed: %w", mod.Name, err)
	}
	return nil
}

func (m *managerImpl) Register(mod *types.ModuleDef) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	if _, exists := m.modules[mod.Name]; exists {
		return fmt.Errorf("module %s already registered", mod.Name)
	}
	m.modules[mod.Name] = mod
	return nil
}

func (m *managerImpl) LookupModule(name string) (*types.ModuleDef, error) {
	m.locker.Lock()
	defer m.locker.Unlock()

	mod, exists := m.modules[name]
	if !exists {
		return nil, fmt.Errorf("module %s not found", name)
	}
	return mod, nil
}

func (m *managerImpl) LookupType(moduleName, typeName string) (*types.TypeDef, error) {
	if moduleName == "" {
		if t, exists := types.InternalTypes[typeName]; exists {
			return t, nil
		}
		moduleName = "System"
	}
	mod, err := m.LookupModule(moduleName)
	if err != nil {
		return nil, err
	}
	for _, t := range mod.Types {
		if t.Name == typeName {
			return t, nil
		}
	}

	return nil, fmt.Errorf("type %s not found in module %s", typeName, moduleName)
}

func (m *managerImpl) Bootstrap(path string) (*types.ModuleDef, error) {
	mod, err := m.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load core library: %w", err)
	}
	if err := m.Link(mod); err != nil {
		return nil, fmt.Errorf("failed to link core library: %w", err)
	}
	if err := m.Register(mod); err != nil {
		return nil, fmt.Errorf("failed to register core library: %w", err)
	}
	return mod, nil
}

func NewManager() Manager {
	m := &managerImpl{}
	return m
}
