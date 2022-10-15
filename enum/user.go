package enum

type UserInfo struct {
	SId      string
	Name     string
	Sex      int32
	Nickname string
	Platform string
	Agent    string
	Coin     int64           // 我的金币（筹码）
	MyBet    map[int32]int64 // 我的下注
}
