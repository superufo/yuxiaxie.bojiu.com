package create

import (
	"fmt"
	"sync"
	"yxxgame.bojiu.com/enum"
)

var YxxGame *YxxGameDesk

type YxxGameDesk struct {
	DeskId    int32                     // 桌子Id
	Status    int32                     // 当前牌桌的状态   1未下注  2下注中  3开牌中  4结算中
	NextTime  int32                     // 下次状态变化的时间
	Set       string                    // 局数
	Dice      []int                     // 骰子
	PlayerMap map[string]*enum.UserInfo // 玩家信息
	AllBet    map[int32]int64           // 所有区域的下注
	Ood       map[int]int               // 赔率     key:多少个一样的  value:赔率
	YxxLock   *sync.RWMutex
}

func NewYxxGame() *YxxGameDesk {
	var tmpOod = make(map[int]int)
	tmpOod[enum.One] = enum.OneOod
	tmpOod[enum.Two] = enum.TwoOod
	tmpOod[enum.Three] = enum.ThreeOod

	var tmpDeskBet = make(map[int32]int64)
	tmpDeskBet[enum.Gourd] = 0
	tmpDeskBet[enum.Fish] = 0
	tmpDeskBet[enum.Money] = 0
	tmpDeskBet[enum.Crab] = 0
	tmpDeskBet[enum.Cock] = 0
	tmpDeskBet[enum.Shrimp] = 0

	return &YxxGameDesk{
		DeskId:    0,
		Status:    0,
		NextTime:  0,
		Set:       "",
		Dice:      nil,
		PlayerMap: make(map[string]*enum.UserInfo),
		AllBet:    tmpDeskBet,
		Ood:       tmpOod,
		YxxLock:   new(sync.RWMutex),
	}
}

func StartDesk() {
	YxxGame = NewYxxGame()
	fmt.Println("------------------------------------", YxxGame)

	go waitTimer()

}
