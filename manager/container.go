package manager

import (
	"fmt"
	"parser/gamedb"
	"parser/util"
)

var gameDB *gamedb.GameDB

func GameDB() *gamedb.GameDB {
	return gameDB
}

var container *Container

type Container struct {
	models map[int]*ModuleManager // key = serverId
}

func NewContainer() *Container {
	return &Container{
		models: make(map[int]*ModuleManager),
	}
}

func GetContainer() *Container {
	if nil == container {
		container = NewContainer()
	}
	return container
}

func (container *Container) Init() error {
	var err error
	var basePath string = "./Configs"
	if gameDB, err = gamedb.Load(basePath); err != nil {
		return err
	}
	container.initModules()
	return nil
}

func (container *Container) Start() error {
	for serverId, moduleManager := range container.models {
		if err := moduleManager.Start(); err != nil {
			return err
		}
		fmt.Printf("server : %d Start\n", serverId)
	}
	return nil
}

func (container *Container) Run() {
	for serverId, moduleManager := range container.models {
		moduleManager.Run()
		fmt.Printf("server : %d Run ...\n", serverId)
	}
}

func (container *Container) Stop() {
	for serverId, moduleManager := range container.models {
		moduleManager.Stop()
		fmt.Printf("server : %d Stopped.\n", serverId)
	}
}

func (container *Container) initModules() error {
	var serverId int = 1 // 简化版
	container.models[serverId] = &ModuleManager{
		DefaultModuleManager: util.NewDefaultModuleManager(),
	}

	if err := container.models[1].Init(); err != nil {
		return err
	}

	return nil
}

func (container *Container) GetManager(serverId int) *ModuleManager {
	if nil == container {
		return nil
	}
	return container.models[serverId]
}
