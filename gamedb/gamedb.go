package gamedb

const CellWidth = 72
const CellHeight = 48

type onDemand map[string]map[string]interface{}

var sceneMaps map[int]*SceneMap

var gameDB *GameDB

func newGameDB() *GameDB {
	return &GameDB{}
}

type GameDB struct {
	OnDemandData onDemand

	Items      map[int]*Item  `client:"items,map" mapKey:"Id"`
	Scenes     map[int]*Scene `client:"scenes,map" mapKey:"Id"`
	OtherDatas []*OtherData   `client:"OtherDatas,array" mapKey:"Id"`
}

type SceneMap struct {
	Id          int
	Name        string
	Width       int
	Height      int
	RoadFlags   map[int32]int8
	walkableMap map[int32]bool
}
