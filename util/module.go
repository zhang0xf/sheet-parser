package util

import (
	"fmt"
	"strings"
	"sync"
)

type Module interface {
	Init() error
	Start() error
	Run()
	Stop()
	GetParent() interface{}
}

type DefaultModule struct {
	Parent interface{}
}

func (defaultModule *DefaultModule) Init() error {
	return nil
}

func (defaultModule *DefaultModule) Start() error {
	return nil
}

func (defaultModule *DefaultModule) Run() {}

func (defaultModule *DefaultModule) Stop() {}

func (defaultModule *DefaultModule) GetParent() interface{} {
	return defaultModule.Parent
}

type DefaultModuleManager struct {
	modules []Module
}

func NewDefaultModuleManager() *DefaultModuleManager {
	return &DefaultModuleManager{
		modules: make([]Module, 5),
	}
}

func (defaultModuleManager *DefaultModuleManager) Init() error {
	for _, module := range defaultModuleManager.modules {
		if nil == module {
			continue
		}
		// moduleTypeName := reflect.TypeOf(module).Name()
		moduleTypeName := fmt.Sprintf("%T", module)
		dot := strings.Index(moduleTypeName, ".")
		moduleName := moduleTypeName[dot+1:]
		if err := module.Init(); err != nil {
			fmt.Println(fmt.Sprintf(moduleName, " Init: %s", err.Error()))
			return err
		}
		fmt.Println(moduleName + " Init")
	}
	return nil
}

func (defaultModuleManager *DefaultModuleManager) Start() error {
	for _, module := range defaultModuleManager.modules {
		if nil == module {
			continue
		}
		if err := module.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (defaultModuleManager *DefaultModuleManager) Run() {
	for _, module := range defaultModuleManager.modules {
		if nil == module {
			continue
		}
		module.Run()
	}
}

func (defaultModuleManager *DefaultModuleManager) Stop() {
	var wait sync.WaitGroup
	for _, module := range defaultModuleManager.modules {
		if nil == module {
			continue
		}
		wait.Add(1)
		go func(module Module) {
			// moduleTypeName := reflect.TypeOf(module).Name()
			moduleTypeName := fmt.Sprintf("%T", module)
			dot := strings.Index(moduleTypeName, ".")
			moduleName := moduleTypeName[dot:]
			fmt.Println(moduleName + " Stopping ...")
			module.Stop()
			wait.Done()
			fmt.Println(moduleName + " Stopped.")
		}(module) // 参数传入(避免匿名函数的引用特性)
	}
	wait.Wait()
}

func (defaultModuleManager *DefaultModuleManager) AppendModule(module Module) Module {
	defaultModuleManager.modules = append(defaultModuleManager.modules, module)
	return module
}
