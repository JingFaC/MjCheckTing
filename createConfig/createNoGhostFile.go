package main

import (
	"fmt"
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
		Finders = append(Finders, Rule{
			KeziCount:   i,
			ShunZiCount: 12 - i,
			DuiZiCount:  1,
			AllCount:    5,
		})
	}
	//Finders = append(Finders, Rule{
	//	KeziCount:   0,
	//	ShunZiCount: 0,
	//	DuiZiCount:  7,
	//	AllCount:    7,
	//})
	//Finders = append(Finders, Rule{
	//	//	KeziCount:   0,
	//	//	ShunZiCount: 0,
	//	//	DuiZiCount:  7,
	//	//	AllCount:    7,
	//	//})
}

type ResultCard struct {
	desc CardXingType
	cid  int
}

var (
//instance *redis.Client
)
var CardWall []Card

func InitCardWall() {
	CardWall = []Card{
		Card{cid: 1, count: 12}, Card{cid: 2, count: 12}, Card{cid: 3, count: 12},
		Card{cid: 4, count: 12}, Card{cid: 5, count: 12}, Card{cid: 6, count: 12}, Card{cid: 7, count: 12}, Card{cid: 8, count: 12}, Card{cid: 9, count: 12},
		//Card{cid: 11, count: 12}, Card{cid: 12, count: 12}, Card{cid: 13, count: 12}, Card{cid: 14, count: 12}, Card{cid: 15, count: 12}, Card{cid: 16, count: 12}, Card{cid: 17, count: 12}, Card{cid: 18, count: 12}, Card{cid: 19, count: 12},
		//Card{cid: 21, count: 12}, Card{cid: 22, count: 12}, Card{cid: 23, count: 12}, Card{cid: 24, count: 12}, Card{cid: 25, count: 12}, Card{cid: 26, count: 12}, Card{cid: 27, count: 12}, Card{cid: 28, count: 12}, Card{cid: 29, count: 12},
		Card{cid: 31, count: 12}, Card{cid: 32, count: 12}, Card{cid: 33, count: 12}, Card{cid: 34, count: 12}, Card{cid: 35, count: 12}, Card{cid: 36, count: 12}, Card{cid: 37, count: 12},
	}
}

var HuFileNameMap map[int64]*os.File
var RepeatCover map[string]byte
var RepeatCover2 map[string]byte
var allRepeated map[string]map[string][]string

func main() {
	//	instance = GetClient()
	InitCardWall()
	InitFinder()

	//HuFileNameMap = make(map[int64]*os.File)
	RepeatCover = make(map[string]byte)
	RepeatCover2 = make(map[string]byte)
	allRepeated = make(map[string]map[string][]string)
	for _, tmpRule := range Finders {
		FindFuluByList(tmpRule, []ResultCard{}, CardWall)
	}
	//for _, v := range HuFileNameMap {
	//	v.Close()
	//}
	write()
	return
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
		CreateOneResult(ResultCopy)

	} else {
		// 继续查找
		FindFuluByList(LeftCopy, ResultCopy, WallCopy)
	}
}

func CreateHuFile(huList []ResultCard, line string) {
	cardTypeCountMap := make(map[CardType]int)
	for _, value := range huList {
		cardType := value.cid/10 + 1
		switch value.desc {
		case DUI:
			cardTypeCountMap[CardType(cardType)] += 2
		default:
			cardTypeCountMap[CardType(cardType)] += 3
		}
	}
	// 生成文件名
	FileName := int64(0)
	index := 0
	for key, count := range cardTypeCountMap {
		if index != 0 {
			FileName = FileName << 7
		}
		index++
		switch key {
		case WAN:
			FileName += int64(1<<4 + count)
		case TONG:
			FileName += int64(2<<4 + count)
		case TIAO:
			FileName += int64(3<<4 + count)
		case ZI:
			FileName += int64(4<<4 + count)
		}
	}
	if _, ok := HuFileNameMap[FileName]; ok {
		file := HuFileNameMap[FileName]
		file.WriteString(line)
		return
	}

	file1, err := os.OpenFile("D:/HuFiles/"+strconv.FormatInt(FileName, 10)+".txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	HuFileNameMap[FileName] = file1

	//instance.SAdd("Hu:"+strconv.FormatInt(FileName, 10), fmt.Sprintf("%s", line))
	file1.WriteString(line)
	return

}

func Maopao(xing CardXingType, finalList []ResultCard) []ResultCard {
	for index := 0; index < len(finalList); index++ {
		if finalList[index].desc != xing {
			continue
		}
		for index2 := 0; index2 < len(finalList)-1; index2++ {
			if finalList[index2].desc != xing {
				continue
			}
			if finalList[index].cid > finalList[index2].cid {
				tmpNum := finalList[index].cid
				finalList[index].cid = finalList[index2].cid
				finalList[index2].cid = tmpNum
			}
		}
	}
	return finalList
}

/// 牌型枚举
type FanType int

const (
	_                      FanType = iota
	WolfWin                        = 1  // 鸡胡
	Common                         = 2  // 平胡
	FourGhost                      = 3  // 四鬼胡牌
	AllPong                        = 4  // 碰碰胡
	MixedOneSuit                   = 5  // 混一色
	SevenPairs                     = 6  // 七小对
	PureOneSuit                    = 7  // 清一色
	MixedPong                      = 8  // 混碰【混一色+碰碰胡】
	LuxurySeverPairs               = 9  // 豪华七小对
	LittleDragons                  = 10 // 小三元
	LittleWinds                    = 11 // 小四喜
	PurePong                       = 12 // 清碰【清一色+碰碰胡】
	MixedTerminals                 = 13 // 混幺九【四副幺九/字牌刻子 + 一副幺九/字牌将】
	GreatDragons                   = 14 // 大三元
	GreatWinds                     = 15 // 大四喜
	AllHonors                      = 16 // 字一色
	PureTerminals                  = 17 // 清幺九【四副幺九刻子 + 一副幺九将】
	DoubleLuxurySeverPairs         = 18 // 双豪华七小对
	ThreeLuxurySeverPairs          = 19 // 三豪华七小对
	AllKong                        = 20 // 十八罗汉
	Heavenly                       = 21 // 天胡
	Earthly                        = 22 // 地胡
	ThirteenOrphans                = 23 // 十三幺
)

func CreateOneResult(tmpList []ResultCard) {
	finalNum := []int{0, 0, 0, 0, 0, 0, 0, 0, 0}
	finalNum1 := []int{0, 0, 0, 0, 0, 0, 0}
	duiziCount := 0
	shunziCount := 0
	keziCount := 0

	duiziCount2 := 0
	keziCount2 := 0

	duiziCount3 := 0
	keziCount3 := 0

	yaojiu_jiang := 0
	yaojiu_ke := 0
	yaojiu_zi_dui := 0
	yaojiu_zi_Ke := 0

	for _, value := range tmpList {
		if value.cid < 10 {
			switch value.desc {
			case DUI:
				duiziCount++
				finalNum[value.cid-1] += 2
				if value.cid%10 == 1 || value.cid%10 == 9 {
					yaojiu_jiang += 1
				}

			case SHUN:
				shunziCount++
				finalNum[value.cid-1] += 1
				finalNum[value.cid-2] += 1
				finalNum[value.cid] += 1
			case KE:
				keziCount++
				if value.cid%10 == 1 || value.cid%10 == 9 {
					yaojiu_ke += 1
				}

				finalNum[value.cid-1] += 3
			}
		}

		if value.cid > 30 {
			switch value.desc {
			case DUI:
				finalNum1[value.cid%10-1] += 2
				yaojiu_zi_dui += 1
				if value.cid > 34 {
					duiziCount3++
				}
				if value.cid > 30 && value.cid <= 34 {
					duiziCount2++
				}
			case SHUN:
				finalNum1[value.cid%10-1] += 1
				finalNum1[value.cid%10-2] += 1
				finalNum1[value.cid%10] += 1
			case KE:
				finalNum1[value.cid%10-1] += 3
				yaojiu_zi_Ke += 1
				if value.cid > 34 {
					keziCount3++
				}
				if value.cid > 30 && value.cid <= 34 {
					keziCount2++
				}
			}
		}
	}

	i := 0
	countMap := make(map[int]int)
	countMap[1] = duiziCount
	countMap[2] = shunziCount
	countMap[3] = keziCount

	countMap[4] = duiziCount2
	countMap[5] = keziCount2

	countMap[6] = duiziCount3
	countMap[7] = keziCount3

	countMap[8] = yaojiu_ke
	countMap[9] = yaojiu_jiang

	countMap[10] = yaojiu_zi_Ke
	countMap[11] = yaojiu_zi_dui

	normalCount := 0
	for _, value := range finalNum {
		normalCount += value
	}
	ziCount := 0
	for _, value := range finalNum1 {
		ziCount += value
	}
	fan1 := createFan(normalCount, countMap, "1")

	fan2 := createFan(ziCount, countMap, "2")
	write2File(i, "1", finalNum, fan1)

	write2File(i, "2", finalNum1, fan2)

}
func createFan(tmpList int, countInfo map[int]int, fileType string) string {
	i := ""

	if fileType == "1" {
		if tmpList < 0 {
			return ""
		}
		// 鸡胡 刻子顺子不能为4
		if countInfo[2] != 4 && countInfo[3] != 4 {
			i += " |" + strconv.Itoa(WolfWin) + "|"
		}
		// 平胡 没有刻子
		if countInfo[3] == 0 {
			i += " |" + strconv.Itoa(Common) + "|"
		}
		// 碰碰 没有顺子
		if countInfo[2] == 0 {
			i += " |" + strconv.Itoa(AllPong) + "|"
			if tmpList < 14 {
				i += " |" + strconv.Itoa(MixedPong) + "|"
			}
		}
		// 清一色
		if tmpList >= 14 {
			i += " |" + strconv.Itoa(PureOneSuit) + "|"
			// 清碰
			if countInfo[2] == 0 {
				i += " |" + strconv.Itoa(PurePong) + "|"
			}
		}
		// 混一色
		if tmpList < 14 {
			i += " |" + strconv.Itoa(MixedOneSuit) + "|"
		}
		// 混幺九
		if countInfo[1] == countInfo[9] && countInfo[8] == countInfo[3] && countInfo[2] == 0 {
			if tmpList < 14 {
				i += " |" + strconv.Itoa(MixedTerminals) + "|"
			}

			i += " |" + strconv.Itoa(PureTerminals) + "|"
		}
	}
	if fileType == "2" {
		if tmpList < 0 {
			return ""
		}
		i += " |" + strconv.Itoa(WolfWin) + "|"
		// 碰碰 没有顺子
		if countInfo[5]+countInfo[7] != 0 {
			if tmpList < 14 {
				i += " |" + strconv.Itoa(MixedPong) + "|"
			}
			i += " |" + strconv.Itoa(AllPong) + "|"
		}
		// 清一色
		if tmpList >= 14 {
			i += " |" + strconv.Itoa(PureOneSuit) + "|"
			// 清碰
			if countInfo[5]+countInfo[7] == 4 {
				i += " |" + strconv.Itoa(PurePong) + "|"
			}
		}
		// 混一色
		if tmpList < 14 {
			i += " |" + strconv.Itoa(MixedOneSuit) + "|"
		}
		// 混幺九
		if (countInfo[5]+countInfo[7] != 0 || countInfo[6]+countInfo[4] != 0) && tmpList < 14 {
			i += " |" + strconv.Itoa(MixedTerminals) + "|"
		}

		if countInfo[7] == 2 && countInfo[6] == 1 {
			i += " |" + strconv.Itoa(LittleDragons) + "|"
		}
		if countInfo[7] == 3 {
			i += " |" + strconv.Itoa(GreatDragons) + "|"
		}

		if countInfo[5] >= 3 && countInfo[4] == 1 {
			i += " |" + strconv.Itoa(LittleWinds) + "|"
		}

		if countInfo[5] == 4 {
			i += " |" + strconv.Itoa(GreatWinds) + "|"
		}

		if tmpList >= 14 {
			i += " |" + strconv.Itoa(AllHonors) + "|"
		}

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

	//tmpRepeated := allRepeated[fileName]
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

		//file1.Close()
		return
	}
	//for j := next; j < len(srcList); j++ {
	//	if srcList[j] == 0 {
	//		continue
	//	}
	//	for i := 1; i < 4; i++ {
	//		if i > endNum {
	//			break
	//		}
	//		tmpList := make([]int, len(srcList))
	//		leftNum := endNum
	//		copy(tmpList, srcList)
	//		if srcList[j] >= i {
	//			tmpList[j] -= i
	//			leftNum -= i
	//			reduceNumFun(tmpList, leftNum, j+1, name, fan)
	//			continue
	//		}
	//		if srcList[j] < i {
	//			tmpList[j] -= srcList[j]
	//			leftNum -= srcList[j]
	//			reduceNumFun(tmpList, leftNum, j+1, name, fan)
	//		}
	//	}
	//}
}

func write() {

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

//
//// 获取RedisClient
//func GetClient() *redis.Client {
//	once.Do(func() {
//		instance = redis.NewClient(&redis.Options{
//			Addr:     "127.0.0.1:6379",
//			Password: "",
//			DB:       1,
//		})
//	})
//
//	return instance
//}
