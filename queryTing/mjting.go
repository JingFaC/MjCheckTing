package mjchecker

import (
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.bianfeng.com/gdmj/gameserver/config"
	"gitlab.bianfeng.com/gdmj/gameserver/library/nredis"
	"gitlab.bianfeng.com/gdmj/gameserver/models/game"
	"strconv"
	"strings"
)

// 查询当前手牌听牌列表
type MJTing struct {
	// 鬼牌数量
	ghostNum int

	// 无鬼加倍
	noGhostMulti int

	// 玩法
	category int

	// 房间类型
	homeType int

	// 鬼牌1
	ghostTips1 int

	// 鬼牌2
	ghostTips2 int

	// 四鬼胡牌听牌信息
	FourGhostWinPoint int

	// 真正的手牌不包括副露和鬼牌
	handCards []int

	// 副露list
	fuuroList []int

	// 规则配置
	ruleConfig *game.HomeConfigItem

	// 支持胡的番型
	supportFanList *config.SupportedFanList

	// 表格redis
	redisClient *redis.Client

	// 听牌查询结果
	NormalResultMap map[int]map[int]int
}

func (mj *MJTing) Init(handCards, fuuroCards, ghost []int) {
	mj.Reset()
	mj.fuuroList = fuuroCards

	mj.setGhost(ghost)
	mj.setHandCards(handCards)
	mj.setFourGhostWinTing()
	mj.setNoGhostMulti()
}

func (mj *MJTing) Reset() {
	mj.ghostNum = 0
	mj.noGhostMulti = 1
	mj.category = 0
	mj.homeType = 0
	mj.redisClient = nredis.GetTblClient()
	mj.NormalResultMap = make(map[int]map[int]int)
	mj.FourGhostWinPoint = 0
}

func (mj *MJTing) SetTingConfig(ruleConfig *game.HomeConfigItem, category int, supportFanList *config.SupportedFanList, homeType int) {
	mj.supportFanList = supportFanList
	mj.ruleConfig = ruleConfig
	mj.category = category
	mj.homeType = homeType
}

func (mj *MJTing) setGhost(ghost []int) {
	ghostNum := len(ghost)
	if ghostNum == 1 {
		mj.ghostTips1 = ghost[0]
	} else if ghostNum == 2 {
		mj.ghostTips1 = ghost[0]
		mj.ghostTips2 = ghost[1]
	}
}

// 判断是否是鬼牌
func (mj *MJTing) isGhost(cid int) bool {
	return cid == mj.ghostTips1 || cid == mj.ghostTips2
}

// 设置手牌cidList
func (mj *MJTing) setHandCards(cards []int) {
	mj.ghostNum = 0
	mj.handCards = make([]int, 0)
	for _, cid := range cards {
		if cid == 0 {
			continue
		}
		if mj.isGhost(cid) {
			mj.ghostNum++
			continue
		}
		mj.handCards = append(mj.handCards, cid)
	}
}

// 设置无鬼加倍倍数
func (mj *MJTing) setNoGhostMulti() {
	configNoGhostMulti := mj.ruleConfig.NoneGhostMultiple
	if configNoGhostMulti != 0 && mj.ghostNum == 0 {
		mj.noGhostMulti = configNoGhostMulti
	}
}

// 获取指定番型分值
func (mj *MJTing) getSupportFanPoint(fanType int) int {
	for index := range *mj.supportFanList {
		tmpFan := (*mj.supportFanList)[index]
		if int(tmpFan.FanType) == fanType {
			return tmpFan.Point
		}
	}
	return 0
}

// 查看是否是四鬼胡牌
func (mj *MJTing) setFourGhostWinTing() {
	// 不支持死鬼胡牌或者鬼牌数量小于四
	if mj.ghostNum < 4 || !mj.ruleConfig.FourGhostWin {
		return
	}

	// 分值
	mj.FourGhostWinPoint = mj.getSupportFanPoint(config.FourGhost)
}

// 用当前手牌cidList 开始查询
func (mj *MJTing) FindCurHandCardTingInfo() {
	alreadyPut := make(map[int]bool)

	findMaxMulti := func(tmpRMap map[int]int) int {
		maxMulti := 0
		for _, tmpMulti := range tmpRMap {
			if tmpMulti > maxMulti {
				maxMulti = tmpMulti
			}
		}
		return maxMulti
	}

	for i := 0; i <= len(mj.handCards)-1; i++ {
		tmpCardCid := mj.handCards[i]

		if _, ok := alreadyPut[tmpCardCid]; ok {
			continue
		}
		alreadyPut[tmpCardCid] = true

		copyHandCards := make([]int, len(mj.handCards))
		copy(copyHandCards, mj.handCards)

		mjTurn := &MJTingTurn{}
		mjTurn.Init(tmpCardCid, append(copyHandCards[:i], copyHandCards[i+1:]...), mj.fuuroList, mj)

		mjTurn.findCurHandCardTingInfo()
		if len(mjTurn.finalResult) == 0 {
			continue
		}

		maxMulti := findMaxMulti(mjTurn.finalResult)
		findMaxMulti(mjTurn.finalResult)
		if mj.ghostTips1 != 0 {
			mjTurn.finalResult[mj.ghostTips1] = maxMulti
		}

		if mj.ghostTips2 != 0 {
			mjTurn.finalResult[mj.ghostTips2] = maxMulti
		}

		mj.NormalResultMap[tmpCardCid] = mjTurn.finalResult

		if 0 == tmpCardCid {
			return
		}
	}

	return
}

// 一轮查询
type MJTingTurn struct {
	Ting                      *MJTing
	putCardId                 int              // 打出的牌Id
	qiduiSyanten              int              // 七对向听数
	qiduiKongNum              int              // 七对豪华数量
	qiduiTingList             []int            // 七对听牌列表
	handCards                 []int            // 手牌list
	fuuroList                 []int            // 附录list
	cardTypeCountMapWithFuuro map[int]int      // 手牌每种牌数量 {1,4,5,4}
	cardStrListWithFuuro      []string         // 手牌每种牌字符串形式
	cardStrListWithoutFuuro   []string         // 手牌每种牌字符串形式
	ghostDistribution         [][]int          // 鬼牌分配情况
	cardCountListWithFuuro    [38]int          // 手牌数量列表
	cardCountListWithoutFuuro [38]int          // 手牌数量列表
	queryResult               []map[int]string // 查询结果
	finalResult               map[int]int      // 最终结果
}

func (turn *MJTingTurn) Init(cardId int, handCards []int, fuuroList []int, ting *MJTing) {
	turn.Reset()
	turn.putCardId = cardId
	turn.Ting = ting
	turn.fuuroList = fuuroList
	turn.handCards = handCards

	turn.setCardCount()
	turn.setCardStrList()
	turn.distributeGhost(turn.Ting.ghostNum+1, false, 0, []int{0, 0, 0, 0})
	turn.calcQiduiziSyanten()
}

func (turn *MJTingTurn) Reset() {
	turn.putCardId = 0
	turn.qiduiKongNum = 0
	turn.qiduiSyanten = 0
	turn.cardTypeCountMapWithFuuro = make(map[int]int)
	turn.cardStrListWithFuuro = make([]string, 4)
	turn.cardStrListWithoutFuuro = make([]string, 4)
	turn.ghostDistribution = make([][]int, 0)
	turn.cardCountListWithFuuro = [38]int{}
	turn.cardCountListWithoutFuuro = [38]int{}
	turn.finalResult = make(map[int]int)
	turn.queryResult = make([]map[int]string, 0)
	turn.qiduiTingList = make([]int, 0)
}

//
func (turn *MJTingTurn) setCardCount() {
	turn.cardTypeCountMapWithFuuro = make(map[int]int)
	for _, cid := range append(turn.handCards, turn.fuuroList...) {
		turn.cardCountListWithFuuro[cid] += 1
		turn.cardTypeCountMapWithFuuro[cid/10] += 1
	}

	for _, cid := range turn.handCards {
		turn.cardCountListWithoutFuuro[cid] += 1
	}
}

func (turn *MJTingTurn) setCardStrList() {
	for i := range turn.cardCountListWithFuuro {
		if i%10 == 0 {
			continue
		}
		turn.cardStrListWithFuuro[i/10] += strconv.Itoa(turn.cardCountListWithFuuro[i])
	}
	for i := range turn.cardCountListWithoutFuuro {
		if i%10 == 0 {
			continue
		}
		turn.cardStrListWithoutFuuro[i/10] += strconv.Itoa(turn.cardCountListWithoutFuuro[i])
	}
}

// 计算七对子向听数和豪华数
func (turn *MJTingTurn) calcQiduiziSyanten() {
	cardCountMap := make(map[int]int)
	for _, cid := range turn.handCards {
		cardCountMap[cid] += 1
	}

	if len(turn.fuuroList) != 0 {
		return
	}

	singleCount := 0
	realDuiziCoung := 0
	duiziCount := 0
	tmpGhostNum := turn.Ting.ghostNum
	for k, num := range cardCountMap {
		if num == 4 {
			duiziCount += 2
			turn.qiduiKongNum += 1
			continue
		}

		if num == 3 {
			duiziCount++
			if tmpGhostNum >= 0 {
				turn.qiduiKongNum += 1
			}
		}

		if num == 2 {
			duiziCount++
			realDuiziCoung += 1
		}

		if num == 1 {
			singleCount += 1
		}

		if num%2 == 1 {
			turn.qiduiTingList = append(turn.qiduiTingList, k)
			if tmpGhostNum >= 0 {
				tmpGhostNum -= 1
			}
		}
	}

	turn.qiduiKongNum += (tmpGhostNum) / 2

	turn.qiduiSyanten = 5 - duiziCount - turn.Ting.ghostNum
}

// 分配鬼牌到不同手牌中
func (turn *MJTingTurn) distributeGhost(remainGhost int, jiangConfirm bool, index int, beforeresult []int) {
	if index+1 > 4 {
		if remainGhost == 0 {
			turn.ghostDistribution = append(turn.ghostDistribution, beforeresult)
		}
		return
	}

	cardCount := 0
	if _, ok := turn.cardTypeCountMapWithFuuro[index]; ok {
		cardCount = turn.cardTypeCountMapWithFuuro[index]
	}

	for i := 0; i <= remainGhost; i++ {
		result := make([]int, len(beforeresult))
		copy(result, beforeresult)
		tmpGhostNum := remainGhost

		left := (cardCount + i) % 3

		if left == 1 {
			i++
			left = (cardCount + i) % 3
		}

		if left == 0 {
			result[index] += i
			tmpGhostNum -= i
			turn.distributeGhost(tmpGhostNum, jiangConfirm, index+1, result)
		}

		if left == 2 && jiangConfirm == false {
			result[index] += i
			tmpGhostNum -= i
			turn.distributeGhost(tmpGhostNum, true, index+1, result)
		}
	}
}

// 计算听牌列表
func (turn *MJTingTurn) findCurHandCardTingInfo() {
	// 七对检测
	//turn.checkQiduizi()

	// 从db查询相应字符串
	turn.queryStrInDbClient()

	// 转化结果
	turn.transformResult()
}

func (turn *MJTingTurn) checkQiduizi() {

	getFanPoint := func(fanType int) int {
		for index := range *turn.Ting.supportFanList {
			fanItem := (*turn.Ting.supportFanList)[index]
			if int(fanItem.FanType) == fanType {
				return fanItem.Point
			}
		}
		return 0
	}

	if turn.qiduiSyanten <= -1 {
		basePoint := 4
		if turn.Ting.homeType == config.SUB_CATEGORY_FRIEND {
			basePoint = 2
		}

		qiduiKongNum := turn.qiduiKongNum
		category := int(turn.Ting.category)
		qiduiMulti := turn.Ting.ruleConfig.SevenPairsMultiple
		if qiduiKongNum > 3 {
			return
		}
		sevenMulti := []int{0, 0, 0, 0}

		sevenMulti[0] = getFanPoint(config.SevenPairs)
		sevenMulti[1] = getFanPoint(config.LuxurySeverPairs)
		sevenMulti[2] = getFanPoint(config.DoubleLuxurySeverPairs)
		sevenMulti[3] = getFanPoint(config.ThreeLuxurySeverPairs)

		noGhostMulti := turn.Ting.noGhostMulti
		for index := range turn.qiduiTingList {
			tingCid := turn.qiduiTingList[index]

			if config.CATEGORY_PUSH_WIN == category {
				turn.finalResult[tingCid] = qiduiMulti * noGhostMulti * basePoint
				continue
			}

			turn.finalResult[tingCid] = sevenMulti[qiduiKongNum] * noGhostMulti
		}
	}
}

// 从Redis中查询
func (turn *MJTingTurn) queryStrInDbClient() {

	reduceUnfitStr := func(src map[int][]string) map[int]string {
		tmpResultMap2 := make(map[int]string)
		for key, _ := range src {
			value := src[key]
			//if !strings.Contains(value[1], "|2|") && strings.Contains(value[0], "|2|") {
			//	value[0] = strings.Replace(value[0], "|2|", "|1|", 1)
			//}
			for i := 1; i <= 9; i++ {
				str1 := strconv.Itoa(i) + " "
				str2 := strconv.Itoa(i) + ","
				if !strings.Contains(value[0], str1) && !strings.Contains(value[0], str2) {
					continue
				}
				if value[1] == "" {
					break
				}
				if !strings.Contains(value[1], str1) && !strings.Contains(value[1], str2) {
					value[0] = strings.Replace(value[0], ","+str1, " ", 9)
					value[0] = strings.Replace(value[0], str1, " ", 9)
					value[0] = strings.Replace(value[0], str2, "", 9)
				}
			}

			tmpResultMap2[key] = value[0]
		}
		return tmpResultMap2
	}

	for _, ghostDis := range turn.ghostDistribution { // 遍历鬼牌分配情况
		flag := true
		tmpResultMap := make(map[int][]string)

		for typeIndex := range turn.cardStrListWithFuuro { // 四种手牌分别查表
			cardStr := turn.cardStrListWithFuuro[typeIndex]

			ghostCount := ghostDis[typeIndex]
			if (cardStr == "000000000" || cardStr == "0000000") && ghostCount == 0 {
				continue
			}

			key := "1"
			if typeIndex == 3 { //字
				key = "2"
			}

			key += strconv.Itoa(ghostCount)
			realKey := fmt.Sprintf("%s:%s", key, cardStr)

			fan, err := turn.Ting.redisClient.Get(realKey).Result()
			if err != nil || fan == "" {
				flag = false
				break
			}

			fan2 := ""
			cardStrWithoutFuuro := turn.cardStrListWithoutFuuro[typeIndex]
			realKey = fmt.Sprintf("%s:%s", key, cardStrWithoutFuuro)
			if cardStrWithoutFuuro != cardStr {
				fan2, err = turn.Ting.redisClient.Get(realKey).Result()
				if err != nil {
					flag = false
					break
				}
			}

			tmpResultMap[typeIndex] = []string{fan, fan2}
		}

		if flag {
			turn.queryResult = append(turn.queryResult, reduceUnfitStr(tmpResultMap))
		}
	}
}

// 分析番型string是否符合指定的番，若符合则返回听的牌
func (turn *MJTingTurn) analyzeFan(fanType int, fanMap map[int]string) ([]int, int) {
	tmpMap := make([]int, 0)
	tmpFlag := 0
	tag := "|" + strconv.Itoa(fanType) + "|"

	splitTingList := func(key int, str string) {
		for _, v := range strings.Split(str, ",") {
			n, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			cid := key*10 + n
			if cid%10 != 0 {
				tmpMap = append(tmpMap, cid)
			}
		}
	}

	for key, lineValue := range fanMap { //  当前所有种类的手牌
		if strings.Contains(lineValue, tag) {
			tmpFlag++
		} else if fanType == config.GreatWinds || fanType == config.GreatDragons || fanType == config.LittleWinds || fanType == config.LittleDragons {
			allFan := strings.Split(lineValue, " ")
			for index := range allFan {
				tmpTingStr := allFan[index]
				index := strings.LastIndex(tmpTingStr, "|")
				splitTingList(key, tmpTingStr[index+1:])
			}
		}

		num := strings.Split(lineValue, tag)
		if len(num) < 2 {
			continue
		}

		fanNum := num[1]
		numIndex := strings.Index(fanNum, " ")
		if -1 == numIndex {
			continue
		}
		numStr := fanNum[:numIndex]
		if numStr == "" {
			continue
		}
		splitTingList(key, numStr)
	}
	return tmpMap, tmpFlag
}

// 检查番型
func (turn *MJTingTurn) checkFan(index int) map[int]int {
	// 最大番型和听牌列表
	fanTing := make(map[int]int)

	oneQueryMap := turn.queryResult[index]
	lenOfMap := len(oneQueryMap)

	_, hasHonors := oneQueryMap[3]
	noGhostMulti := turn.Ting.noGhostMulti
	category := int(turn.Ting.category)

	basePoint := 4
	if turn.Ting.homeType == config.SUB_CATEGORY_FRIEND {
		basePoint = 2
	}

	addToFanMap := func(tmpList []int, fanType int) {
		for index := range tmpList {
			cid := tmpList[index]
			if _, ok := fanTing[cid]; !ok {
				fanTing[cid] = fanType
				continue
			}
			if fanType > fanTing[cid] {
				fanTing[cid] = fanType
			}
		}
	}

	for _, value := range *turn.Ting.supportFanList {
		fanType := int(value.FanType)
		point := value.Point * noGhostMulti
		if point == 0 {
			continue
		}

		// 推倒胡分值
		if config.CATEGORY_PUSH_WIN == category {
			point = basePoint * noGhostMulti
		}

		tmpList, tmpFlag := turn.analyzeFan(fanType, oneQueryMap)

		if (fanType == config.AllHonors || fanType == config.PureOneSuit || fanType == config.PurePong) && tmpFlag == 1 { // 一个成立
			addToFanMap(tmpList, point)
		} else if (fanType == config.GreatWinds || fanType == config.GreatDragons || fanType == config.LittleWinds || fanType == config.LittleDragons) && tmpFlag != 0 && hasHonors {
			addToFanMap(tmpList, point)
		} else if (fanType == config.MixedOneSuit || fanType == config.MixedPong) && lenOfMap == 2 && tmpFlag == 2 && hasHonors { // 两个成立
			addToFanMap(tmpList, point)
		} else if tmpFlag == lenOfMap && tmpFlag != 0 && fanType != config.MixedOneSuit && fanType != config.MixedPong { // 多个成立
			addToFanMap(tmpList, point)
		}
	}
	return fanTing
}

func (turn *MJTingTurn) transformResult() {
	// 遍历查询结果
	for index := range turn.queryResult {
		tmpResult := turn.checkFan(index)

		for cid, fan := range tmpResult {
			mFan, ok := turn.finalResult[cid]
			if !ok {
				turn.finalResult[cid] = fan
				continue
			}
			if fan > mFan {
				turn.finalResult[cid] = fan
			}
		}
	}
}
