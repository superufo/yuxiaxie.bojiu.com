package create

import (
	"runtime"
	"time"
	"yxxgame.bojiu.com/enum"
)

// waitTimer 未下注  5S
func waitTimer() {
	YxxGame.YxxLock.Lock()
	YxxGame.Status = enum.Bet_Before
	YxxGame.NextTime = 5
	YxxGame.YxxLock.Unlock()
	ticker := time.NewTicker(time.Second * 1)
	count := 0
	go func() {
		for {
			<-ticker.C
			count++
			YxxGame.YxxLock.Lock()
			YxxGame.NextTime = int32(5 - count)
			YxxGame.YxxLock.Unlock()

			// 触发条件
			if count == 5 {
				ticker.Stop()
				go bettingTimer()
				runtime.Goexit()
			}
		}
	}()
}

// bettingTimer 下注中  10S
func bettingTimer() {
	YxxGame.YxxLock.Lock()
	YxxGame.Status = enum.Betting
	YxxGame.NextTime = 10
	YxxGame.YxxLock.Unlock()

	ticker := time.NewTicker(time.Second * 1)
	count := 0
	go func() {
		for {
			<-ticker.C
			count++
			YxxGame.YxxLock.Lock()
			YxxGame.NextTime = int32(10 - count)
			YxxGame.YxxLock.Unlock()

			// 触发条件
			if count == 10 {
				ticker.Stop()
				go openResTimer()
				runtime.Goexit()
			}
		}
	}()
}

// openResTimer 开牌中  5S
func openResTimer() {
	YxxGame.YxxLock.Lock()
	YxxGame.Status = enum.Show_Result
	YxxGame.NextTime = 8
	YxxGame.YxxLock.Unlock()
	ticker := time.NewTicker(time.Second * 1)
	count := 0
	go func() {
		for {
			<-ticker.C
			count++
			YxxGame.YxxLock.Lock()
			YxxGame.NextTime = int32(8 - count)
			YxxGame.YxxLock.Unlock()

			// 触发条件
			if count == 8 {
				// 清除桌子数据
				YxxGame.YxxLock.Lock()
				YxxGame.AllBet = nil
				YxxGame.Dice = nil
				for _, v := range YxxGame.PlayerMap {
					v.MyBet = nil
				}
				YxxGame.YxxLock.Unlock()

				ticker.Stop()
				go waitTimer()
				runtime.Goexit()
			}
		}
	}()
}
