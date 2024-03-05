package manager

import "parser/util"

type OtherManager struct {
	util.DefaultModule
}

func NewOtherManager() *OtherManager {
	return &OtherManager{}
}

func (otherManager *OtherManager) Init() error {
	return nil
}
