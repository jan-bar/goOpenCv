package st_table

import (
	"testing"

	"mahjong/shanten"
)

func TestCalcTable(t *testing.T) {
	s := []struct {
		tile string
		exp  int
	}{
		{tile: "33m 5555p 66s 556666z", exp: 1},
		{tile: "13579m 13579s 135p", exp: 4},
		{tile: "13579m 12379s 135p", exp: 3},
		{tile: "123456789m 147s 14m", exp: 1},
		{tile: "123456789m 147s 1m", exp: 2},
		{tile: "258m 258s 258p 12345z", exp: 6},
		{tile: "123456789m 1134p", exp: 0},
		{tile: "123456789m 11345p", exp: -1},
		{tile: "1m", exp: 0},
		{tile: "1555m", exp: 0},
		{tile: "2247m", exp: 1},
		{tile: "11234m", exp: -1},
		{tile: "5555m", exp: 1},
		{tile: "5555z", exp: 1},
		{tile: "1111m 2222s 3333p 11115z", exp: 0},
		{tile: "111m 222s 333p 1115z", exp: 0},
		{tile: "135m 259s 368p 12345z", exp: 6},
	}
	for _, v := range s {
		n0, m := shanten.CalcShanTenStr(&v.tile)
		if v.exp != n0 {
			// 与期望shan ten向听数不匹配时报错
			t.Fatal(v.tile, v.exp, n0, m)
		}
		t.Log("Calc", v.tile, n0, m)

		n1 := CalcTableShanTen(v.tile)
		t.Log("CalcShanTan", v.tile, n1)

		if n0 != n1 { // 只有小于13张牌时会不一样
			t.Log("not equ", n0, n1)
		}
		t.Log("-------------------------------")
	}
}
