package mjchecker

import (
	"github.com/go-redis/redis"
	"gitlab.bianfeng.com/gdmj/gameserver/config"
	"gitlab.bianfeng.com/gdmj/gameserver/library/nredis"
	"gitlab.bianfeng.com/gdmj/gameserver/models/game"
	"strconv"
	"strings"
)

// 查询当前手牌听牌列表
type MJTing struct {
	ghostNum        int                      // 鬼牌数量
	noGhostMulti    int                      // 无鬼加倍
	category        int                      //玩法
	ghostList       []int                    // 列表
	handCards       []int                    // 真正的手牌没有副露
	checkCards      []int                    // 查听的牌包括手牌和副露
	fuuroList       []int                    // 副露list
	redisClient     *redis.Client            // redisConn
	supportFanList  *config.SupportedFanList // 支持胡的番型
	ruleConfig      *game.HomeConfigItem     // 玩法
	FourGhostWinMap map[int]int              // 四鬼胡牌听牌信息
	NormalResultMap map[int]map[int]int      // 听牌查询结果
}

func (mj *MJTing) Init(handCards, fuuroCards, ghost []int, ruleConfig *game.HomeConfigItem, category int, supportFanList *config.SupportedFanList) {
	mj.Reset()
	mj.ghostList = ghost
	mj.supportFanList = supportFanList
	mj.ruleConfig = ruleConfig
	mj.fuuroList = fuuroCards
	mj.category = category
	mj.setHandCards(handCards)
	mj.setFourGhostWinTing()
	mj.setNoGhostMulti()
}

func (mj *MJTing) Reset() {
	mj.ghostNum = 0
	mj.noGhostMulti = 1
	mj.category = 0
	mj.ghostList = make([]int, 0)
	mj.checkCards = make([]int, 0)
	mj.redisClient = nredis.GetTblClient()
	mj.NormalResultMap = make(map[int]map[int]int)
	mj.FourGhostWinMap = make(map[int]int)
}

func (mj *MJTing) isGhost(cid int) bool {
	for _, ghostCid := range mj.ghostList {
		if ghostCid != 0 && cid == ghostCid {
			return true
		}
	}
	return false
}

// 设置手牌cidList
func (mj *MJTing) setHandCards(cards []int) {
	mj.checkCards = make([]int, 0)
	mj.ghostNum = 0
	for _, cid := range cards {
		if cid == 0 {
			continue
		}
		if mj.isGhost(cid) {
			mj.ghostNum++
			continue
		}
		mj.checkCards = append(mj.checkCards, cid)
		mj.handCards = append(mj.handCards, cid)
	}
	mj.checkCards = append(mj.checkCards, mj.fuuroList...)
}

// 设置无鬼加倍
func (mj *MJTing) setNoGhostMulti() {
	configNoGhostMulti := mj.ruleConfig.NoneGhostMultiple
	if configNoGhostMulti != 0 && mj.ghostNum == 0 {
		mj.noGhostMulti = configNoGhostMulti
	}
}

func (mj *MJTing) checkCidValid(cid int) bool {
	deskConfig := mj.ruleConfig
	if deskConfig == nil {
		return false
	}
	if cid%10 == 0 {
		return false
	}
	if cid < 10 {
		return deskConfig.HasMan
	}
	if cid < 20 {
		return deskConfig.HasPin
	}
	if cid < 30 {
		return deskConfig.HasSou
	}
	if cid < 38 {
		if deskConfig.HasHonors {
			return true
		}
		if deskConfig.Category == config.CATEGORY_RED_DRAGON {
			// 没有字牌但如果是红中王，则必须有红中和白板
			return cid == 37 || cid == 35
		}
	}
	return false
}

func (mj *MJTing) getSupportFanPoint(fanType int) int {
	for index := range *mj.supportFanList {
		tmpFan := (*mj.supportFanList)[index]
		if int(tmpFan.FanType) == fanType {
			return tmpFan.Point
		}
	}
	return 0
}

func (mj *MJTing) setFourGhostWinTing() {
	// 不支持四鬼胡牌
	if mj.ghostNum < 4 || !mj.ruleConfig.FourGhostWin {
		return
	}

	fourGhostPoint := mj.getSupportFanPoint(config.FourGhost)
	if 13 == len(mj.checkCards)+mj.ghostNum {
		mj.FourGhostWinMap[0] = fourGhostPoint
		return
	}

	for index := range mj.checkCards {
		cid := mj.checkCards[index]
		if cid%10 == 0 || mj.isGhost(cid) {
			continue
		}
		mj.FourGhostWinMap[cid] = fourGhostPoint
	}

}

// 用当前手牌cidList 开始查询
func (mj *MJTing) FindCurHandCardTingInfo() {
	handLintCount := len(mj.checkCards)
	if 13 == handLintCount+mj.ghostNum {
		mjTurn := &MJTingTurn{}
		mjTurn.Init(0, mj.checkCards, mj.handCards, mj)
		mjTurn.findCurHandCardTingInfo()
		if len(mjTurn.finalResult) == 0 {
			return
		}
		mj.NormalResultMap[0] = mjTurn.finalResult
		return
	}

	alreadyPut := make(map[int]bool)
	for index := range mj.fuuroList {
		flag := true
		for hCardIndex := range mj.handCards {
			if mj.handCards[hCardIndex] == mj.fuuroList[index] {
				flag = false
				break
			}
		}
		if flag {
			alreadyPut[mj.fuuroList[index]] = true
		}
	}

	for index := range mj.checkCards {
		tmpCardCid := mj.checkCards[index]

		if _, ok := alreadyPut[tmpCardCid]; ok {
			continue
		}
		alreadyPut[tmpCardCid] = true

		copyHandCards := make([]int, len(mj.checkCards))
		copy(copyHandCards, mj.checkCards)

		mjTurn := &MJTingTurn{}
		mjTurn.Init(tmpCardCid, append(copyHandCards[:index], copyHandCards[index+1:]...), mj.handCards, mj)
		mjTurn.findCurHandCardTingInfo()
		if len(mjTurn.finalResult) == 0 {
			continue
		}
		mj.NormalResultMap[tmpCardCid] = mjTurn.finalResult
	}

	return
}

// 一轮查询
type MJTingTurn struct {
	Ting              *MJTing
	putCardId         int              // 打出的牌Id
	qiduiSyanten      int              // 七对向听数
	qiduiKongNum      int              // 七对豪华数量
	qiduiTingList     []int            // 七对听牌列表
	cardTypeCountMap  map[int]int      // 手牌每种牌数量 {1,4,5,4}
	cardStrList       []string         // 手牌每种牌字符串形式
	ghostDistribution [][]int          // 鬼牌分配情况
	cardCountList     [38]int          // 手牌数量列表
	queryResult       []map[int]string // 查询结果
	finalResult       map[int]int      // 最终结果
}

func (turn *MJTingTurn) Init(cardId int, handCards, checkCards []int, ting *MJTing) {
	turn.Reset()
	turn.putCardId = cardId
	turn.Ting = ting

	turn.setCardCount(checkCards)
	turn.setCardStrList()
	turn.distributeGhost(turn.Ting.ghostNum+1, false, 0, []int{0, 0, 0, 0})
	turn.calcQiduiziSyanten(handCards)
}

func (turn *MJTingTurn) Reset() {
	turn.putCardId = 0
	turn.qiduiKongNum = 0
	turn.qiduiSyanten = 0
	turn.cardTypeCountMap = make(map[int]int)
	turn.cardStrList = make([]string, 4)
	turn.ghostDistribution = make([][]int, 0)
	turn.cardCountList = [38]int{}
	turn.finalResult = make(map[int]int)
	turn.queryResult = make([]map[int]string, 0)
	turn.qiduiTingList = make([]int, 0)
}

func (turn *MJTingTurn) setCardCount(handCards []int) {
	turn.cardTypeCountMap = make(map[int]int)
	for _, cid := range handCards {
		turn.cardCountList[cid] += 1
		turn.cardTypeCountMap[cid/10] += 1
	}
}

func (turn *MJTingTurn) setCardStrList() {
	for i := range turn.cardCountList {
		if i%10 == 0 {
			continue
		}
		turn.cardStrList[i/10] += strconv.Itoa(turn.cardCountList[i])
	}
}

func (turn *MJTingTurn) calcQiduiziSyanten(handCards []int) {
	cardCountMap := make(map[int]int)
	for _, cid := range handCards {
		cardCountMap[cid] += 1
	}
	duiziCount := 0
	for k, num := range cardCountMap {
		if num == 4 {
			duiziCount += 2
			turn.qiduiKongNum += 1
			continue
		}
		if num >= 2 {
			duiziCount++
		}
		if num%2 == 1 {
			turn.qiduiTingList = append(turn.qiduiTingList, k)
		}
	}

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
	if _, ok := turn.cardTypeCountMap[index]; ok {
		cardCount = turn.cardTypeCountMap[index]
		//turn.distributeGhost(remainGhost, jiangConfirm, index+1, beforeresult)
		//return
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
	turn.checkQiduizi()
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
		point := 0
		category := int(turn.Ting.category)
		qiduiMulti := turn.Ting.ruleConfig.SevenPairsMultiple
		if config.CATEGORY_PUSH_WIN == category {
			point = qiduiMulti
		} else if turn.qiduiKongNum == 0 {
			point = getFanPoint(config.SevenPairs)
		} else if turn.qiduiKongNum == 1 {
			point = getFanPoint(config.LuxurySeverPairs)
		} else if turn.qiduiKongNum == 2 {
			point = getFanPoint(config.DoubleLuxurySeverPairs)
		} else if turn.qiduiKongNum == 3 {
			point = getFanPoint(config.ThreeLuxurySeverPairs)
		}

		for index := range turn.qiduiTingList {
			turn.finalResult[turn.qiduiTingList[index]] = point
		}
	}
}

// 从Redis中查询
func (turn *MJTingTurn) queryStrInDbClient() {
	for _, ghostDis := range turn.ghostDistribution { // 遍历鬼牌分配情况
		flag := true
		tmpResultMap := make(map[int]string)

		for typeIndex := range turn.cardStrList { // 四种手牌分别查表
			cardStr := turn.cardStrList[typeIndex]

			ghostCount := ghostDis[typeIndex]
			if (cardStr == "000000000" || cardStr == "0000000") && ghostCount == 0 {
				continue
			}

			key := "1"
			if typeIndex == 3 { //字
				key = "2"
			}

			key += strconv.Itoa(ghostCount)

			findResult := turn.Ting.redisClient.HGet(key, cardStr)
			fan, err := findResult.Result()
			if err != nil {
				flag = false
				break
			}

			tmpResultMap[typeIndex] = fan
		}

		if flag {
			turn.queryResult = append(turn.queryResult, tmpResultMap)
		}
	}
}

// 分析番型string是否符合指定i的番，若符合则返回听的牌
func (turn *MJTingTurn) analyzeFan(fanType int, fanMap map[int]string) (map[int]byte, int) {
	tmpMap := make(map[int]byte)
	tmpFlag := 0
	tag := "|" + strconv.Itoa(fanType) + "|"
	for key, lineValue := range fanMap { //  当前所有种类的手牌
		if !strings.Contains(lineValue, tag) {
			break
		}

		tmpFlag++
		num := strings.Split(lineValue, tag)
		if len(num) < 2 {
			continue
		}

		fanNum := num[1]
		numIndex := strings.Index(fanNum, " ")
		numStr := fanNum[:numIndex]
		if numStr == "" {
			//tmpMap = append(tmpMap, key*10)
			continue
		}
		for _, v := range strings.Split(numStr, ",") {
			n, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			cid := key*10 + n
			if turn.Ting.checkCidValid(cid) {
				tmpMap[cid] = 1
			}
		}
	}
	return tmpMap, tmpFlag
}

// 检查番型
func (turn *MJTingTurn) checkFan(index int) (int, map[int]byte) {
	// 最大番型和听牌列表
	maxFan := 0
	fanTing := make(map[int]byte)

	oneQueryMap := turn.queryResult[index]
	lenOfMap := len(oneQueryMap)

	_, hasHonors := oneQueryMap[3]
	noGhostMulti := turn.Ting.noGhostMulti
	category := int(turn.Ting.category)
	qiduiMulti := turn.Ting.ruleConfig.SevenPairsMultiple

	for _, value := range *turn.Ting.supportFanList {
		fanType := int(value.FanType)
		point := value.Point * noGhostMulti
		if point == 0 {
			continue
		}

		// 推倒胡默认两倍
		if config.CATEGORY_PUSH_WIN == category {
			point = qiduiMulti
		}

		tmpMap, tmpFlag := turn.analyzeFan(fanType, oneQueryMap)

		if (fanType == config.AllHonors || fanType == config.GreatWinds || fanType == config.GreatDragons || fanType == config.LittleWinds ||
			fanType == config.LittleDragons || fanType == config.PureOneSuit || fanType == config.PurePong) && lenOfMap == 1 && tmpFlag == 1 { // 一个成立
			if fanType > maxFan {
				maxFan = point
				fanTing = tmpMap
			}
		} else if (fanType == config.MixedOneSuit || fanType == config.MixedPong) &&
			lenOfMap == 2 && tmpFlag == 2 && hasHonors { // 两个成立
			if fanType > maxFan {
				maxFan = point
				fanTing = tmpMap
			}
		} else if tmpFlag == lenOfMap && tmpFlag != 0 && fanType != 5 && fanType != 8 { // 多个成立
			if fanType > maxFan {
				maxFan = point
				fanTing = tmpMap
			}
		}
	}
	return maxFan, fanTing
}

func (turn *MJTingTurn) transformResult() {
	// 遍历查询结果
	for index := range turn.queryResult {
		maxFan, tmpResult := turn.checkFan(index)

		for cid := range tmpResult {
			if maxFan > turn.finalResult[cid] {
				turn.finalResult[cid] = maxFan
			}
		}
	}
}
