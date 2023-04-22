package shanten

import (
	"strconv"
	"testing"

	"github.com/jan-bar/goOpenCv/mahjong/janbar-helper/util"
)

// go test -v -run TestCalculateShanTen -o shanten.exe
// .\shanten.exe -test.v
func TestCalculateShanTen(t *testing.T) {
	// 从文件中读取手牌和期望向听数
	// 33m 5555p 66s 556666z,1
	// 13579m 13579s 135p,4
	tiles := util.LoadTileFromFile("tile.tmp.txt")
	for _, v := range tiles {
		exp, err := strconv.Atoi(v.Line[1])
		if err != nil {
			continue
		}

		n0, m := CalcShanTenTile(v.Tile)
		if exp != n0 {
			// 与期望shan ten向听数不匹配时报错
			t.Fatal(v.Line[0], exp, n0, m)
		}
		t.Log("Calc", v.Line[0], n0, m)

		n1, m := CalcShanTan(v.Tile)
		t.Log("CalcShanTan", v.Line[0], n1, m)

		if n0 != n1 { // 只有小于13张牌时会不一样
			t.Log("not equ", n0, n1)
		}
		t.Log("-------------------------------")
	}
}
