package model

// CheckBet 获取骰子掷出的结果 ------------> map[int32]int   区域:次数
func CheckBet(res []int) (request map[int32]int) {
	tmpMap := make(map[int32]int)
	tmpRes := make([]int, len(res))
	copy(tmpRes, res)
	for _, v := range tmpRes {
		tmpMap[int32(v)] += 1
	}

	return tmpMap
}

// Convert []int ---->  []int32
func Convert(in []int) (res []int32) {
	var tmp = make([]int, len(in))
	copy(tmp, in)
	for j := 0; j < len(in); j++ {
		res = append(res, int32(tmp[j]))
	}
	return res
}
