package mjchecker

import (
	"gitlab.bianfeng.com/gdmj/gameserver/base/nlogs"
	"testing"
	"time"
)

func TestTing(t *testing.T) {
	type a struct {
		aa int
	}
	pa := &a{aa: 1}
	t1 := time.Now()
	tt := 0
	t2 := pa.aa
	for i := 1; i < 10000000000; i++ {
		tt = t2
	}
	//cards := []int{1, 2, 3, 11, 12, 13}
	//cards := []int{11, 11, 12, 12, 13, 13, 14, 14, 15, 15, 23, 23, 21, 21}
	//fuuro := []int{}
	//ghost := []int{}
	//// 手牌信息
	//
	//cfg := game.GetConfigLoader().GetHomeConfig(10001)
	//homeDefaultFanList := game.GetConfigLoader().GetFanDefaultList(int(home.HOME_TYPE_GOLD), int(cfg.Category))
	//t1 := time.Now()
	//// 查找
	//mjt := MJTing{}
	//mjt.Init(cards, fuuro, ghost, cfg, &homeDefaultFanList)
	//mjt.FindCurHandCardTingInfo()

	nlogs.Home.Infof("耗时 %d---%d", time.Since(t1).Nanoseconds()/1000000, tt)
}
