package main

import (
	"fmt"
	"parser/manager"
	"parser/util"
)

// 主要功能包括: Module管理,读取表格数据.

func main() {

	container := manager.GetContainer()

	if err := container.Init(); err != nil {
		fmt.Printf("container Init err : %s\n", err.Error())
		return
	}

	if err := container.Start(); err != nil {
		fmt.Printf("container Start err : %s\n", err.Error())
		return
	}

	container.Run()
	util.WaitTerminate()
	container.Stop()
	fmt.Printf("server stopped.\n")
}
