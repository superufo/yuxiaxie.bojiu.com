package model

import "github.com/shopspring/decimal"

/*
** 赔率
**
 */

// Settlement 结算
func Settlement(bet map[int32]int64, res map[int32]int, ood map[int]int) int64 {
	var x int64

	for k, v := range res {
		if _, ok := bet[k]; ok {
			x += int64(ood[v]) * bet[k]
		}
	}

	return addRet(x)
}

func addRet(i int64) int64 {
	var ret float32 = 1000
	var ood float32 = 0.05
	return (i*1000 - i*decimal.NewFromFloat(float64(ood*ret)).IntPart()) / 1000

}
