package main

import (
	"strings"

	//"github.com/go-redis/redis"
	"os"
	"strconv"
)

type CardType int

const (
	WAN  = 1
	TONG = 2
	TIAO = 3
	ZI   = 4
)

type CardXingType int

const (
	DUI  = 1
	KE   = 2
	SHUN = 3
)

// 牌
type Card struct {
	cid   int
	count int
}

// 查找器
type Rule struct {
	KeziCount   int
	ShunZiCount int
	DuiZiCount  int
	AllCount    int
}

var Finders []Rule

// 初始化查找器
func InitFinder() {
	for i := 0; i <= 4; i++ {
		for j := 4 - i; j >= 0; j-- {
			for k := 0; k <= 1; k++ {
				Finders = append(Finders, Rule{
					KeziCount:   i,
					ShunZiCount: j,
					DuiZiCount:  k,
					AllCount:    i + j + k,
				})
			}
		}
	}
	//for i := 1; i <= 7; i++ {
	//	Finders = append(Finders, Rule{
	//		KeziCount:   0,
	//		ShunZiCount: 0,
	//		DuiZiCount:  i,
	//		AllCount:    i,
	//	})
	//}

}

type ResultCard struct {
	desc CardXingType
	cid  int
}

var CardWall []Card

func InitCardWall() {
	CardWall = []Card{
		Card{cid: 1, count: 4}, Card{cid: 2, count: 4}, Card{cid: 3, count: 4},
		Card{cid: 4, count: 4}, Card{cid: 5, count: 4}, Card{cid: 6, count: 4}, Card{cid: 7, count: 4}, Card{cid: 8, count: 4}, Card{cid: 9, count: 4},
		//Card{cid: 11, count: 12}, Card{cid: 12, count: 12}, Card{cid: 13, count: 12}, Card{cid: 14, count: 12}, Card{cid: 15, count: 12}, Card{cid: 16, count: 12}, Card{cid: 17, count: 12}, Card{cid: 18, count: 12}, Card{cid: 19, count: 12},
		//Card{cid: 21, count: 12}, Card{cid: 22, count: 12}, Card{cid: 23, count: 12}, Card{cid: 24, count: 12}, Card{cid: 25, count: 12}, Card{cid: 26, count: 12}, Card{cid: 27, count: 12}, Card{cid: 28, count: 12}, Card{cid: 29, count: 12},
	}
}

var HuFileNameMap map[int64]*os.File
var RepeatCover map[string]byte
var RepeatCover2 map[string]byte
var allRepeated map[string]map[string][]string

func main() {

	// 初始化牌墙 不同玩法牌可能不一样
	InitCardWall()

	// 胡牌时允许的原子牌型 几个对几个刻
	InitFinder()

	// 中间计算过程的存储容器
	InitContainer()

	// 递归计算各种胡牌可能 各个玩法都一样不用修改
	CalculateCard()

	// 根据牌型规律 配置玩法番型

	// 写入文件
	WriteToFile()

	return
}

func CalculateCard() {
	for _, tmpRule := range Finders {
		FindFuluByList(tmpRule, []ResultCard{}, CardWall)
	}
}

func InitContainer() {
	RepeatCover = make(map[string]byte)
	RepeatCover2 = make(map[string]byte)
	allRepeated = make(map[string]map[string][]string)
}

func ZhaoDuiZi(LeftRule Rule, LastResultCards []ResultCard, LeftCardWall []Card) {
	for index, _ := range LeftCardWall {
		if LeftCardWall[index].count >= 2 {
			HandleCardWallAndRule(DUI, index, LeftRule, LastResultCards, LeftCardWall)
		}
	}
}
func ZhaoKeZi(LeftRule Rule, LastResultCards []ResultCard, LeftCardWall []Card) {
	for index, _ := range LeftCardWall {
		if LeftCardWall[index].count >= 3 {
			HandleCardWallAndRule(KE, index, LeftRule, LastResultCards, LeftCardWall)
		}
	}
}

func ZhaoShunZi(LeftRule Rule, LastResultCards []ResultCard, LeftCardWall []Card) {
	for index, _ := range LeftCardWall {
		yushu := LeftCardWall[index].cid % 10
		if !(LeftCardWall[index].cid < 30 && yushu >= 2 && yushu <= 8) {
			continue
		}
		if !(LeftCardWall[index-1].count >= 1 && LeftCardWall[index].count >= 1 && LeftCardWall[index+1].count >= 1) {
			continue
		}
		HandleCardWallAndRule(SHUN, index, LeftRule, LastResultCards, LeftCardWall)
	}
}

func FindFuluByList(leftCardType Rule, tmpResult []ResultCard, cardList []Card) {

	if leftCardType.DuiZiCount > 0 {
		ZhaoDuiZi(leftCardType, tmpResult, cardList)
	} else if leftCardType.KeziCount > 0 {
		ZhaoKeZi(leftCardType, tmpResult, cardList)
	} else if leftCardType.ShunZiCount > 0 {
		ZhaoShunZi(leftCardType, tmpResult, cardList)
	}
}

func HandleCardWallAndRule(ReduceType CardXingType, index int, LeftRule Rule, LastResultCards []ResultCard, LeftCardWall []Card) {
	// 剩余查找
	LeftCopy := LeftRule

	// 结果
	ResultCopy := make([]ResultCard, len(LastResultCards))
	copy(ResultCopy, LastResultCards)

	// 剩余牌
	WallCopy := make([]Card, len(LeftCardWall))
	copy(WallCopy, LeftCardWall)

	switch ReduceType {
	case DUI:
		WallCopy[index].count -= 2
		LeftCopy.DuiZiCount -= 1
	case KE:
		WallCopy[index].count -= 3
		LeftCopy.KeziCount -= 1
	case SHUN:
		WallCopy[index-1].count -= 1
		WallCopy[index].count -= 1
		WallCopy[index+1].count -= 1
		LeftCopy.ShunZiCount -= 1
	}

	// 牌墙扣除两张牌
	LeftCopy.AllCount -= 1

	// 添加到临时结果
	ResultCopy = append(ResultCopy, ResultCard{desc: ReduceType, cid: LeftCardWall[index].cid})

	if LeftCopy.AllCount == 0 {
		// 尺子没了 输出一次结果
		createOneResult(ResultCopy)

	} else {
		// 继续查找
		FindFuluByList(LeftCopy, ResultCopy, WallCopy)
	}
}

/// 牌型枚举
type FanType int

const (
	_              FanType = iota
	Common                 = 1 // 平胡
	DaDui                  = 2 // 大对子
	SeverPairs             = 3 // 七对
	Pure                   = 4 // 清一色
	DanDiao                = 5 // 单吊
	LongSeverPairs         = 6 // 龙七对

	PureDaDui   = 7 // 清大对
	PureDanDiao = 8 // 请单吊
	PurePairs   = 9 // 清七对

	PureLongBei = 10 // 清龙背
)

func createOneResult(tmpList []ResultCard) {
	finalNum := []int{0, 0, 0, 0, 0, 0, 0, 0, 0}

	duiziCount := 0
	shunziCount := 0
	keziCount := 0

	for _, value := range tmpList {
		switch value.desc {
		case DUI:
			duiziCount++
			finalNum[value.cid-1] += 2
		case SHUN:
			shunziCount++
			finalNum[value.cid-1] += 1
			finalNum[value.cid-2] += 1
			finalNum[value.cid] += 1
		case KE:
			keziCount++

			finalNum[value.cid-1] += 3
		}
	}

	i := 0
	countMap := make(map[int]int)
	countMap[DUI] = duiziCount
	countMap[SHUN] = shunziCount
	countMap[KE] = keziCount

	normalCount := 0
	for _, value := range finalNum {
		normalCount += value
	}

	fan1 := createFan(&finalNum, normalCount, countMap)
	write2File(i, "1", finalNum, fan1)
}

func createFan(finalNum *[]int, tmpList int, countInfo map[int]int) string {
	i := ""

	if tmpList < 0 {
		return ""
	}

	kongNum := func() int {
		kongCount := 0
		for _, cardCount := range *finalNum {
			if cardCount >= 4 {
				kongCount++
			}
		}
		return kongCount
	}

	// 平胡 鸡胡
	i += " |" + strconv.Itoa(Common) + "|"

	// 七对 没有刻子 没有顺子
	if countInfo[SHUN] == 0 && countInfo[KE] == 0 {
		i += " |" + strconv.Itoa(SeverPairs) + "|"
		// 清七对
		if tmpList >= 14 {
			i += " |" + strconv.Itoa(PurePairs) + "|"
		}
		// 龙七对 有至少一个杠
		if kongNum() == 1 {
			i += " |" + strconv.Itoa(LongSeverPairs) + "|"
			// 清龙背
			if tmpList >= 14 {
				i += " |" + strconv.Itoa(PureLongBei) + "|"
			}
		}
	}

	// 单吊 大对 碰碰胡 没有顺子
	if countInfo[SHUN] == 0 && countInfo[DUI] <= 1 {
		i += " |" + strconv.Itoa(DanDiao) + "|"
		i += " |" + strconv.Itoa(DaDui) + "|"
		// 清单吊
		if tmpList >= 14 {
			i += " |" + strconv.Itoa(PureDanDiao) + "|"
			i += " |" + strconv.Itoa(PureDaDui) + "|"
		}
	}

	// 清一色
	if tmpList >= 14 {
		i += " |" + strconv.Itoa(Pure) + "|"
	}

	return i
}

func write2File(reduceNum int, fileName string, finalNum []int, fan string) {
	sum := 0
	for _, value := range finalNum {
		sum += value
	}

	if sum < reduceNum {
		return
	}
	reduceNumFun(finalNum, reduceNum, 0, fileName+strconv.Itoa(reduceNum), fan)
}

func reduceNumFun(srcList []int, endNum, next int, name string, fan string) {
	if endNum <= 0 {

		if _, ok := allRepeated[name]; !ok {
			allRepeated[name] = make(map[string][]string)
		}

		line := ""
		for _, value := range srcList {
			line += strconv.Itoa(value)
		}
		// 查找
		if _, ok := allRepeated[name][line]; !ok {
			allRepeated[name][line] = make([]string, 0)
			tmpList := strings.Split(fan, " ")
			for _, value := range tmpList {
				allRepeated[name][line] = append(allRepeated[name][line], value)
			}
		} else {
			for _, value2 := range strings.Split(fan, " ") {
				flag := false
				for _, value := range allRepeated[name][line] {
					if value == value2 {
						flag = true
					}
				}
				if flag == false {
					allRepeated[name][line] = append(allRepeated[name][line], value2)
				}
			}
		}

		return
	}
}

func WriteToFile() {
	for fileName, value := range allRepeated {
		file1, _ := os.OpenFile("D:/HuFiles/"+fileName+".txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		for key, value1 := range value {

			str := key + "*"
			for _, lValue := range value1 {
				str += lValue + " "
			}

			file1.WriteString(str + "\n")
		}
		file1.Close()
	}
}
