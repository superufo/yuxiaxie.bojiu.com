package model

import (
	"math/rand"
	"time"
)

// LicensingFunc
func LicensingFunc() (Cards []int) {
	Cards = RandMath()
	return

}

// Cards 牌组
var (
	Cards = []int{1, 1, 1, // 葫芦
		2, 2, 2, // 鱼
		3, 3, 3, // 金钱
		4, 4, 4, // 蟹
		5, 5, 5, // 鸡
		6, 6, 6, // 虾
	}

	r = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// RandMath 随机取三个数
func RandMath() []int {
	// 拷贝
	AllPai := make([]int, 0, 0)
	tmpCards := make([]int, len(Cards))
	copy(tmpCards, Cards)
	tmp := Cards
	FaPaiCiShu := 3               // 初始化牌型
	SuiJiMap := make(map[int]int) // 记录随机数
	for i := 0; i < FaPaiCiShu; i++ {
		WeiZhi := r.Intn(18)
		_, ok := SuiJiMap[WeiZhi]
		if ok {
			FaPaiCiShu++
			continue
		}

		SuiJiMap[WeiZhi] = WeiZhi
		AllPai = append(AllPai, tmp[WeiZhi])

	}

	Result := AllPai[0:3]

	return Result
}
