package enum

const (
	CMD_ENTER_GAME_3         uint16 = 3101 // 进入游戏,回复场次信息
	CMD_INIT_GAME_3          uint16 = 3102 // 获取游戏数据
	CMD_GAME_3_BETS          uint16 = 3103 // 玩家投注,返回投注的信息
	CMD_GAME_3_STATUS_CHANGE uint16 = 3104 // 状态改变

	CMDS string = "3101,3102,3103,3104"
)
