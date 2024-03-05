package gamedb

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"parser/gamelib/pcommon"
	"parser/util"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	textcensor "github.com/kai1987/go-text-censor"
	"github.com/tealeg/xlsx"
)

var startRow int = 3 // Sheet起始行
var startCol int = 2 // Sheet起始列

type fieldInfo struct {
	idx     int                  // 列索引
	field   *reflect.StructField // 列对应objs的field
	group   string               // 组名(暂时用不到)
	colName string               // sheet列名
}

type columnInfos map[int]*fieldInfo
type columnRecords map[string]bool

type sheetInfo struct {
	sheetName string
	obj       interface{}                        // 用于存放Sheet每行数据的数据结构
	loader    func(*GameDB, []interface{}) error // 填充objs到GameDB的方法(arrayLoader,mapLoader...)
}

func Load(basePath string) (*GameDB, error) {
	f, err := os.Stat(basePath)
	if err != nil {
		return nil, err
	}
	if f.IsDir() {
		return loadExcel(basePath, "gamedb.dat")
	}
	return loadFile(basePath)
}

func loadExcel(basePath string, datFileName string) (*GameDB, error) {

	gameDB = newGameDB()

	// 优先加载gamedb.dat文件
	datFilePath := filepath.Join(basePath, datFileName)
	if f, err := os.Stat(datFilePath); nil == err && !f.IsDir() {
		pcommon.PrintMemStats("loadExcel Alloc before loadDatFile: ")
		if err := gameDB.loadFile(datFilePath); err != nil {
			fmt.Printf("load .dat file error : %s\n", err.Error())
			gameDB = newGameDB()
		}
		runtime.GC()
		pcommon.PrintMemStats("loadExcel Alloc after loadDatFile GC: ")
	}

	hasChange, err := gameDB.loadExcels(filepath.Join(basePath, "excels"))
	if err != nil {
		return nil, err
	}

	// 动态数据(不以配置文件的形式加入客户端,在游戏运行时客户端动态请求服务器数据,onDemandData.json由jenkins生成)
	if temp, err := loadOnDemandData(basePath); nil == err {
		gameDB.OnDemandData = temp
	}

	if hasChange {
		gameDB.createFile(datFilePath)
	}

	// 组装(生成冗余的,但对于某些module方便的数据结构)
	gameDB.Patch()

	if err := gameDB.Check(); err != nil {
		return nil, err
	}

	if err := loadScenes(gameDB, basePath); err != nil {
		return nil, err
	}

	// 加载敏感词或短语
	if err := loadSensitivePhrases(filepath.Join(basePath, "filtertext.txt")); err != nil {
		return nil, err
	}

	fmt.Println("加载gameDB成功!")
	return gameDB, nil
}

func loadOnDemandData(basePath string) (onDemand, error) {
	onDemandData := make(map[string]map[string]interface{})

	fInfo, err := os.Stat(basePath)
	if err != nil {
		fmt.Printf("loadOnDemandData() basePath not found\n")
		return nil, fmt.Errorf("loadOnDemandData() basePath not found")
	}

	var onDemandFilePath string
	if fInfo.IsDir() {
		onDemandFilePath = basePath + "/onDemandData.json"
	} else {
		onDemandFilePath = filepath.Dir(basePath) + "/onDemandData.json"
	}

	f, err := os.Open(onDemandFilePath)
	if err != nil {
		fmt.Printf("loadOnDemandData() open %s error : %v\n", onDemandFilePath, err)
		return nil, fmt.Errorf("loadOnDemandData() open %s error : %v", onDemandFilePath, err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("loadOnDemandData() Read %s error : %v", onDemandFilePath, err)
	}

	var demandData map[string]interface{}
	json.Unmarshal(b, &demandData)
	for k, v := range demandData {
		if assertion, ok := v.(map[string]interface{}); ok {
			onDemandData[k] = assertion
		}
	}

	return onDemandData, nil
}

func loadScenes(gameDB *GameDB, basePath string) error {
	if nil == gameDB {
		return fmt.Errorf("loadScenes() invalid param")
	}

	scenePath := filepath.Join(basePath, "scenes/map_%d.json")
	mapIds := gameDB.getSceneMapIds()
	var info string = ""
	var waiter sync.WaitGroup
	var errChan chan error = make(chan error, len(mapIds))
	var dataChan chan *SceneMap = make(chan *SceneMap, len(mapIds))

	waiter.Add(len(mapIds))
	for _, id := range mapIds {
		go func(mapId int) {
			if sceneMap, err := loadSceneMap(makeMapPath(scenePath, mapId), mapId); err != nil {
				errChan <- err
			} else {
				dataChan <- sceneMap
			}
			waiter.Done()
		}(id)
	}
	waiter.Wait()

	go func() {
		close(errChan)
	}()

	// 关闭通道后,不会阻塞,也不会无限循环. range对chan操作有封装
	for err := range errChan {
		info += fmt.Sprintf("%s,\n", err.Error())
	}

	if info != "" {
		return fmt.Errorf(info)
	}

	go func() {
		close(dataChan)
	}()

	for sceneMap := range dataChan {
		if nil == sceneMaps {
			sceneMaps = make(map[int]*SceneMap)
		}
		sceneMaps[sceneMap.Id] = sceneMap
	}

	return nil
}

// 加载map_****.json
func loadSceneMap(scenePath string, sceneId int) (*SceneMap, error) {
	b, err := ioutil.ReadFile(scenePath)
	if err != nil {
		fmt.Printf("loadSceneMap() read file %s, error : %v\n", scenePath, err)
		return nil, err
	}

	sceneMap := SceneMap{}
	// Unmarshal() 会对 SceneMap{} 中的map结构分配内存,详见F12.
	if err := json.Unmarshal(b, &sceneMap); err != nil {
		fmt.Printf("loadSceneMap() json.Unmarshal() file %s, err : %v\n", scenePath, err)
		return nil, err
	}

	if len(sceneMap.RoadFlags) < 1 {
		fmt.Printf("loadSceneMap() Load no Walkable point in map : %s\n", scenePath)
		return nil, fmt.Errorf("config has no roadFlags")
	}

	sceneMap.Width = int(math.Ceil(float64(sceneMap.Width) / CellWidth))
	sceneMap.Height = int(math.Ceil(float64(sceneMap.Height) / CellHeight))

	// 重新计算并生成walkableMap
	walkableMap := make(map[int32]bool, len(sceneMap.RoadFlags))
	for k, v := range sceneMap.RoadFlags {
		x, y := k/1000, k%1000
		walkableMap[x<<16|y] = v&1 == 1
		if x<<16|y < 0 {
			fmt.Printf("x : %d, y : %d, k : %d\n", x, y, k)
		}
		delete(sceneMap.RoadFlags, k)
	}
	sceneMap.walkableMap = walkableMap

	// 文件不提供Id和Name,则生成
	if sceneMap.Id < 1 {
		sceneMap.Id = sceneId
	}

	if len(sceneMap.Name) < 1 {
		sceneMap.Name = fmt.Sprintf("scene_%d", sceneMap.Id)
	}

	return &sceneMap, nil
}

func makeMapPath(basePath string, sceneId int) string {
	return fmt.Sprintf(basePath, sceneId)
}

func loadSensitivePhrases(filterFilePath string) error {
	err := textcensor.InitWordsByPath(filterFilePath, false)
	if err != nil {
		return err
	}

	defaultPunctuation := "0123456789abcdefghijklmnopqrstuvwxyz !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~，。？；：”’￥（）——、！……"
	textcensor.SetPunctuation(defaultPunctuation) // 忽略标点符号
	return nil
}

func loadFile(basePath string) (*GameDB, error) {

	gameDB = newGameDB()

	if err := gameDB.loadFile(basePath); err != nil {
		return nil, err
	}

	if temp, err := loadOnDemandData(basePath); nil == err {
		gameDB.OnDemandData = temp
	}

	gameDB.Patch()

	if err := gameDB.Check(); err != nil {
		return nil, err
	}

	dirPath := filepath.Dir(basePath)

	if err := loadScenes(gameDB, dirPath); err != nil {
		return nil, err
	}

	if err := loadSensitivePhrases(filepath.Join(dirPath, "filtertext.txt")); err != nil {
		return nil, err
	}

	return gameDB, nil
}

func (gameDB *GameDB) loadFile(datFilePath string) error {
	startTime := time.Now()

	defer func() {
		fmt.Println("GameDB loadFile used time(seconds) : ", time.Since(startTime).Seconds())
	}()

	f, err := os.Open(datFilePath)
	if err != nil {
		return err
	}

	defer f.Close() // 系统资源，不被GC,手动释放

	reader := bufio.NewReader(f)
	decoder := gob.NewDecoder(reader)
	return decoder.Decode(&gameDB)
}

func (gameDB *GameDB) createFile(filePath string) error {
	now := time.Now()

	defer func() {
		fmt.Printf("create %s use time : %f\n", filePath, time.Since(now).Seconds())
	}()

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	enc := gob.NewEncoder(w)
	enc.Encode(gameDB)

	return w.Flush()
}

func (gameDB *GameDB) loadExcels(basePath string) (bool, error) {
	var counter int
	var threadError error
	var waiter sync.WaitGroup
	startTime := time.Now()

	defer func() {
		fmt.Printf("GameDB loadExcels used time(seconds) : %f\n", time.Since(startTime).Seconds())
	}()

	for _, excelInfo := range fileInfos {
		excelPath := filepath.Join(basePath, excelInfo.excelName)

		f, err := os.Stat(excelPath)
		if err != nil {
			return false, fmt.Errorf("stat file ( %s ) has err : %s", excelPath, err)
		}

		modifyTime := f.ModTime().UnixNano() // 时间戳（纳秒）

		if getfileModTime(excelInfo.excelName) == modifyTime {
			fmt.Printf("file ( %s ) not modified.\n", excelInfo.excelName)
			continue
		}

		setfileModTime(excelInfo.excelName, modifyTime)
		counter++

		waiter.Add(1)
		go func(excelPath string, excelInfo fileInfo) {
			startTime := time.Now()
			if err := gameDB.loadExcel(excelPath, excelInfo.sheetInfos); err != nil {
				threadError = err
				fmt.Printf("GameDB load %s has error, error : %s.\n", excelInfo.excelName, err)
			}
			fmt.Printf("GameDB load %s complete, used time(seconds) : %s.\n", excelInfo.excelName, time.Since(startTime))
			waiter.Done()
		}(excelPath, excelInfo)
	}

	waiter.Wait()

	if threadError != nil {
		return false, threadError
	}

	if counter == 0 {
		return false, fmt.Errorf("no excels be loaded")
	}

	fmt.Printf("excels totally loaded : %d.\n", counter)
	return true, nil
}

func (gameDB *GameDB) loadExcel(excelPath string, sheetInfos []sheetInfo) error {
	xlsxFile, err := xlsx.OpenFile(excelPath)
	if err != nil {
		return err
	}

	for _, sheetInfo := range sheetInfos {
		if _, ok := xlsxFile.Sheet[sheetInfo.sheetName]; !ok {
			return fmt.Errorf("no sheet ( %s ) found", sheetInfo.sheetName)
		}

		sheet := xlsxFile.Sheet[sheetInfo.sheetName]

		objs, err := gameDB.readSheet(sheet, sheetInfo.obj, startRow, startCol)
		if err != nil {
			return err
		}

		if err := sheetInfo.loader(gameDB, objs); err != nil {
			return err
		}
	}

	return nil
}

// 表格填充格式:
// 从第3行,第2列开始填写.
func (gameDB *GameDB) readSheet(sheet *xlsx.Sheet, obj interface{}, startRow int, startCol int) ([]interface{}, error) {

	objT := reflect.TypeOf(obj)
	var result []interface{} = make([]interface{}, 0)

	if !(objT.Kind() == reflect.Ptr && objT.Elem().Kind() == reflect.Struct) {
		return nil, fmt.Errorf("obj must be a struct")
	}

	if len(sheet.Rows) <= startRow || len(sheet.Cols) <= startCol {
		return nil,
			fmt.Errorf("sheet ( %s ) not meets the request, rows ( %d ), cols ( %d )", sheet.Name, len(sheet.Rows), len(sheet.Cols))
	}

	colInfos, colRecords, maxCol := gameDB.collectColumnInfo(sheet, objT, startRow, startCol)

	if len(colInfos) == 0 {
		return nil, fmt.Errorf("no column found for current sheet : %s", sheet.Name)
	}

	isPass, pField := gameDB.checkAllFieldFoundColumn(objT, colRecords)
	if !isPass {
		if pField != nil {
			return nil, fmt.Errorf("sheet ( %s ) not found column : %s\n, update excels and try again", sheet.Name, pField.Name)
		}
	}

	for i, row := range sheet.Rows {
		if i < startRow { // 真正的数据在colName(title)下一行
			continue
		}

		if nil == row || len(row.Cells) == 0 { // 表格不允许有空行.
			return nil, fmt.Errorf("empty row : sheet ( %s ) empty row at %d row", sheet.Name, i+1)
		}

		// 利用反射创建obj对象,每行数据都需要一个obj,否则数据会覆盖
		objStruct := reflect.New(objT.Elem())

		for j, cell := range row.Cells {
			if j < startCol-1 {
				continue
			}

			if j > maxCol {
				break // 最大列后,可能存在诸多注释列不需要解析.
			}

			if _, ok := colInfos[j]; !ok {
				continue
			}
			fieldInfo := colInfos[j]
			cellString, err := cell.FormattedValue()
			if err != nil {
				return nil, fmt.Errorf("sheet ( %s ), cell (row : %d, col : %d) fomatted err : %s", sheet.Name, i, j, err.Error())
			}
			cellString = strings.TrimSpace(cellString)

			// 自增列(e.g : Id列)不能为空
			if j == startCol-1 && i >= startRow && (nil == cell || len(cellString) == 0) {
				return nil, fmt.Errorf("sheet ( %s ), cell (row : %d, col : %d) autoincrease column should not empty", sheet.Name, i, j)
			}

			// 获取结构体某个Field
			fieldV := objStruct.Elem().Field(fieldInfo.idx)

			// struct field 是否 addressable(可取地址的) 和 exported(可导出的:用大小写区分)
			if !fieldV.CanSet() {
				return nil, fmt.Errorf("sheet ( %s ), cell (row : %d, col : %d) can not set to field( %s )",
					sheet.Name, i, j, objT.Elem().Field(fieldInfo.idx).Name)
			}

			// 自定义类型解析
			// .(Decoder)类型断言,判断类型(*type)是否实现了Decoder接口
			// 即使无数据,也需要在Decode()中为自定义fieldV分配内存,使其拥有零值,否则具体逻辑使用时需要nil判断,极易出错.
			if decoder, ok := fieldV.Addr().Interface().(Decoder); ok {
				if err := decoder.Decode(cellString); err != nil {
					return nil, fmt.Errorf("sheet ( %s ), cell (row : %d, col : %d) decode err : %s", sheet.Name, i, j, err.Error())
				}
				continue
			}

			// 基础类型解析
			// 无数据不解析,使用默认零值
			if len(cellString) == 0 {
				continue
			}

			switch objT.Elem().Field(fieldInfo.idx).Type.Kind() {
			case reflect.Bool:
				cellBool, err := strconv.ParseBool(cellString)
				if err != nil {
					return nil, fmt.Errorf("sheet ( %s ), cell (row : %d, col : %d) ParseBool err : %s", sheet.Name, i, j, err.Error())
				}
				fieldV.SetBool(cellBool)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				cellFloat, err := strconv.ParseFloat(cellString, 64)
				if err != nil {
					return nil, fmt.Errorf("sheet ( %s ), cell (row : %d, col : %d) ParseFloat for Int err : %s", sheet.Name, i, j, err.Error())
				}
				fieldV.SetInt(int64(util.RoundFloat(cellFloat, 0)))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				cellUint, err := strconv.ParseUint(cellString, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("sheet ( %s ), cell (row : %d, col : %d) ParseUint err : %s", sheet.Name, i, j, err.Error())
				}
				fieldV.SetUint(cellUint)
			case reflect.Float32, reflect.Float64:
				cellFloat, err := strconv.ParseFloat(cellString, 64)
				if err != nil {
					return nil, fmt.Errorf("sheet ( %s ), cell (row : %d, col : %d) ParseFloat err : %s", sheet.Name, i, j, err.Error())
				}
				fieldV.SetFloat(cellFloat)
			case reflect.String:
				str := regexp.MustCompile("\n").ReplaceAllString(cellString, "")
				fieldV.SetString(strings.Replace(str, `"`, `\"`, -1))
			default:
				return nil, fmt.Errorf("sheet ( %s ), cell (row : %d, col : %d) field type error", sheet.Name, i, j)
			}
		}

		result = append(result, objStruct.Interface()) // 需要转换为interface类型
	}

	return result, nil
}

// sheet colName为集合A, struct field为集合B (A应>=B)
func (gameDB *GameDB) collectColumnInfo(sheet *xlsx.Sheet, objT reflect.Type, startRow int, startCol int) (columnInfos, columnRecords, int) {
	var maxCol int = 0
	infos := make(columnInfos)
	records := make(columnRecords)

	for idx, cell := range sheet.Rows[startRow-1].Cells {
		if idx < startCol-1 {
			continue
		}

		if cell == nil {
			break
		}

		cellString := strings.TrimSpace(cell.Value)

		if len(cellString) == 0 {
			break
		}

		maxCol = idx

		// objs struct field
		for i := 0; i < objT.Elem().NumField(); i++ {
			field := objT.Elem().Field(i)
			if len(field.Tag.Get("col")) == 0 { //struct的field没有tag(col),即该字段不需解析sheet
				continue
			}

			if field.Tag.Get("col") == cellString {
				infos[idx] = &fieldInfo{
					idx:     i,
					field:   &field,
					group:   field.Tag.Get("group"), // 组名(暂时用不到)
					colName: cellString,
				}
				records[cellString] = true
				// break // 具有相同col的struct field: 后面覆盖前面
			}
		}
	}

	return infos, records, maxCol
}

// 检查所有的field在sheet中均有对应的column
func (gameDB *GameDB) checkAllFieldFoundColumn(objT reflect.Type, records columnRecords) (bool, *reflect.StructField) {
	for i := 0; i < objT.Elem().NumField(); i++ {
		field := objT.Elem().Field(i)
		col := field.Tag.Get("col")
		if len(col) == 0 {
			continue
		}
		if !records[col] {
			return false, &field
		}
	}
	return true, nil
}

func (gameDB *GameDB) getSceneMapIds() []int {
	var sceneIdsSet map[int]struct{} = make(map[int]struct{})

	if len(gameDB.Scenes) < 1 {
		return nil
	}

	for _, scene := range gameDB.Scenes {
		if _, ok := sceneIdsSet[scene.MapId]; ok {
			fmt.Printf("scene的MapId可以相同, TODO ...")
		}
		sceneIdsSet[scene.MapId] = struct{}{}
	}

	var temp []int = make([]int, 0, len(sceneIdsSet))

	for k := range sceneIdsSet {
		temp = append(temp, k)
	}

	return temp
}

// arrayLoader() 和 mapLoader() 返回一个loader函数以供使用.
// 可以将不同形参(eg.fieldName)传递给loader函数
func arrayLoader(fieldName string) func(*GameDB, []interface{}) error {
	return func(gameDB *GameDB, objs []interface{}) error {
		fieldV := reflect.ValueOf(gameDB).Elem().FieldByName(fieldName)
		var keySet map[int64]struct{} = make(map[int64]struct{})
		switch fieldV.Kind() {
		case reflect.Slice:
			// slice动态扩展,需要重新分配适当内存
			if fieldV.IsNil() || fieldV.Len() > 0 {
				fieldV.Set(reflect.MakeSlice(fieldV.Type(), 0, len(objs)))
			}

			for _, obj := range objs {
				objV := reflect.ValueOf(obj)
				fieldV.Set(reflect.Append(fieldV, objV))
				if err := checkKeyUnique(fieldName, keySet, objV); err != nil {
					return err
				}
			}
		case reflect.Array:
			// 数组创建时指定大小
			for i, obj := range objs {
				objV := reflect.ValueOf(obj)
				fieldV.Index(i).Set(objV)
				if err := checkKeyUnique(fieldName, keySet, objV); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("field %s is not an array", fieldName)
		}
		return nil
	}
}

// 允许自定义key
// fieldName 与 sheetname 有关联方便查错.
func mapLoader(fieldName string, key string) func(*GameDB, []interface{}) error {
	return func(gameDB *GameDB, objs []interface{}) error {
		fieldV := reflect.ValueOf(gameDB).Elem().FieldByName(fieldName)
		if fieldV.Kind() != reflect.Map {
			return fmt.Errorf("field %s is not a map", fieldName)
		}

		// make map
		if fieldV.IsNil() || fieldV.Len() > 0 {
			fieldV.Set(reflect.MakeMap(fieldV.Type()))
		}

		for _, obj := range objs {
			objV := reflect.ValueOf(obj)
			keyFieldV := objV.Elem().FieldByName(key)
			if !keyFieldV.IsValid() {
				return fmt.Errorf("key field %s wrong, when setting %s", key, fieldName)
			}
			if fieldV.MapIndex(keyFieldV).IsValid() {
				return fmt.Errorf("表 %s 列 %s 值->%v 重复了", fieldName, key, keyFieldV)
			}
			fieldV.SetMapIndex(keyFieldV, objV)
		}

		return nil
	}
}

// objV is a struct which defined in objs.go (eg.Item)
func checkKeyUnique(fieldName string, keySet map[int64]struct{}, objV reflect.Value) error {
	for _, v := range []string{"Id", "Lvl"} { // 填表格式固定有好处
		keyFieldV := objV.Elem().FieldByName(v)
		if keyFieldV.IsValid() {
			key := keyFieldV.Int()
			if _, ok := keySet[key]; ok {
				return fmt.Errorf("表　%s 字段 %s, %d, %v 重复了", fieldName, v, key, objV)
			}
			keySet[key] = struct{}{}
			break
		}
	}
	return nil
}
