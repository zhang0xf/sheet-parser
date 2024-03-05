package gamedb

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	SEMICOLON = ";"
	COMMA     = ","
	COLON     = ":"
	PIPE      = "|"
	SPACE     = " "
	HLINE     = "-"
)

type Decoder interface {
	Decode(str string) error
}

type ItemInfo struct {
	Id    int `client:"id"`
	Count int `client:"count"`
}

type PropInfo struct {
	Key   int `client:"key"`
	Value int `client:"value"`
}

type IntSlice []int
type ItemInfos []*ItemInfo

func (itemInfos *ItemInfos) Decode(cellString string) error {
	if nil == itemInfos {
		*itemInfos = make(ItemInfos, 0)
	}

	if len(cellString) == 0 {
		return nil // 不返回错误,表格单元格数据允许为空.
	}

	list := strings.Split(strings.Trim(strings.TrimSpace(cellString), SEMICOLON), SEMICOLON)
	if len(list) == 0 {
		return nil
	}

	for _, elem := range list {
		var itemInfo ItemInfo

		infos := strings.Split(strings.TrimSpace(elem), COMMA)
		if len(infos) != 2 {
			return fmt.Errorf("ItemInfo格式错误")
		}

		id, err := strconv.Atoi(infos[0])
		if err != nil {
			return err
		}
		itemInfo.Id = id

		count, err := strconv.Atoi(infos[1])
		if err != nil {
			return err
		}
		itemInfo.Count = count

		*itemInfos = append(*itemInfos, &itemInfo)
	}

	return nil
}

func (intSlice *IntSlice) Decode(cellString string) error {
	if nil == intSlice {
		*intSlice = make(IntSlice, 0)
	}

	if len(cellString) == 0 {
		return nil // 不返回错误,表格单元格数据允许为空.
	}

	list := strings.Split(strings.TrimSpace(cellString), COMMA)
	*intSlice = make(IntSlice, len(list))

	for _, elem := range list {
		if len(elem) == 0 {
			continue
		}

		value, err := strconv.Atoi(elem)
		if err != nil {
			return err
		}

		*intSlice = append(*intSlice, value)
	}

	return nil
}

func (propInfo *PropInfo) Decode(cellString string) error {
	if nil == propInfo {
		*propInfo = PropInfo{}
	}

	if len(cellString) == 0 {
		return nil // 不返回错误,表格单元格数据允许为空.
	}

	list := strings.Split(strings.TrimSpace(cellString), COMMA)
	if len(list) != 2 {
		return fmt.Errorf("属性信息格式错误:" + cellString)
	}

	key, err := strconv.Atoi(list[0])
	if err != nil {
		return err
	}

	value, err := strconv.Atoi(list[1])
	if err != nil {
		return err
	}

	propInfo.Key = key
	propInfo.Value = value
	return nil
}
