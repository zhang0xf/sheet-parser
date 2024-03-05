package gamedb

type Item struct {
	Id           int       `col:"id" client:"id"`
	Name         string    `col:"name" client:"name"`                        //名称
	Note         string    `col:"note" client:"note"`                        //注解
	IconId       int       `col:"iconId" client:"iconId"`                    //图标
	ItemLvl      int       `col:"itemLvl"`                                   //物品等级
	Level        int       `col:"level" client:"level"`                      //等级需求
	Vip          int       `col:"vip"`                                       //VIP等级需求
	Color        int       `col:"color" client:"color"`                      //颜色
	Type         int       `col:"type" client:"type"`                        //类型
	BagTag       int       `col:"bagTag" client:"bagTag"`                    //背包类型
	Count        int       `col:"count" client:"count"`                      //是否叠加
	CanSell      int       `col:"canSell" client:"canSell"`                  //是否能出售给系统
	SellGet      ItemInfos `col:"sellGet" client:"sellGet"`                  //出售获得
	DropId       string    `col:"dropId"`                                    //掉落途径
	UseType      int       `col:"useType" client:"useType"`                  //使用类型
	UseTypePrams IntSlice  `col:"useTypePrams" client:"useTypePrams"`        //使用参数
	GetSource    IntSlice  `col:"getSource" client:"getSource"`              //获得途径
	Price        PropInfo  `col:"price" client:"price" checker:"itemOption"` //快捷购买代币类型,价格
	Cherish      int       `col:"cherish"`                                   //是否珍惜掉落
	InFly        int       `col:"inFly"`                                     //是否加入飞升榜
	Border       int       `col:"border"`                                    //边框
	Purpose      string    `col:"purpose" client:"purpose"`                  //物品说明
	Usefor       int       `col:"usefor" client:"usefor"`                    //用途
	IsAction     int       `col:"isAction" client:"isAction"`                //是否动态图标
}

type Scene struct {
	Id    int `col:"id" client:"id"`
	MapId int `col:"mapId" client:"mapId"` // 地图ID
}

type OtherData struct {
	Id   int    `col:"id" client:"id"`
	Data string `col:"data" client:"data"`
}
