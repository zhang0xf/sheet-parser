package gamedb

type fileInfo struct {
	excelName  string
	sheetInfos []sheetInfo
}

type fileRecords map[string]int64

var fileInfos []fileInfo    // 所有表格文件
var fileModTime fileRecords // 文件修改时间记录

// package func init() before main
func init() {
	fileInfos = []fileInfo{
		{"item.xlsx", []sheetInfo{
			{"item", &Item{}, mapLoader("Items", "Id")},
		}},
		{"otherData.xlsx", []sheetInfo{
			{"otherData", &OtherData{}, arrayLoader("Items")},
		}},
	}
}

func getfileModTime(filePath string) int64 {
	if nil == fileModTime {
		return 0
	}
	return fileModTime[filePath]
}

func setfileModTime(filePath string, nanoTime int64) {
	if nil == fileModTime {
		fileModTime = make(fileRecords)
	}
	fileModTime[filePath] = nanoTime
}
