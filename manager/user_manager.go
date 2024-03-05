package manager

import "parser/util"

type UserManager struct {
	util.DefaultModule
}

func NewUserManager() *UserManager {
	return &UserManager{}
}

func (userManager *UserManager) Init() error {
	return nil
}
