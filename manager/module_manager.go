package manager

import "parser/util"

type ModuleManager struct {
	*util.DefaultModuleManager
	UserManager  *UserManager
	OtherManager *OtherManager
}

func (moduleManager *ModuleManager) Init() error {
	moduleManager.UserManager = moduleManager.AppendModule(NewUserManager()).(*UserManager)
	moduleManager.UserManager.Parent = moduleManager

	moduleManager.OtherManager = moduleManager.AppendModule(NewOtherManager()).(*OtherManager)
	moduleManager.OtherManager.Parent = moduleManager

	if err := moduleManager.DefaultModuleManager.Init(); err != nil {
		return err
	}

	return nil
}
