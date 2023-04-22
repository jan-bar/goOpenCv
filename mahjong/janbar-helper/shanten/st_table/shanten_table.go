package st_table

import (
	"archive/zip"
	"bufio"
	"bytes"
	_ "embed"
	"strconv"
	"strings"

	"mahjong/util"
)

var (
	//go:embed ShantenTable.zip
	shanTenTableZip []byte

	shanTenMap = loadShanTenTableTxt(shanTenTableZip)
)

func loadShanTenTableTxt(b []byte) map[int]int {
	r := bytes.NewReader(b)
	zr, err := zip.NewReader(r, r.Size())
	if err != nil {
		panic(err)
	}
	// 读取zip中指定查表文件,还原相关map
	fr, err := zr.Open("ShantenTable.txt")
	if err != nil {
		return nil
	}
	//goland:noinspection GoUnhandledErrorResult
	defer fr.Close()

	br := bufio.NewScanner(fr)
	br.Scan()
	k, err := strconv.Atoi(br.Text())
	if err != nil {
		panic(err)
	}
	res := make(map[int]int, k)

	for br.Scan() {
		line := strings.Fields(br.Text())
		if len(line) == 2 {
			k, err = strconv.Atoi(line[0])
			if err != nil {
				panic(err)
			}
			res[k], err = strconv.Atoi(line[1])
			if err != nil {
				panic(err)
			}
		}
	}
	if err = br.Err(); err != nil {
		panic(err)
	}
	return res
}

func CalcTableShanTen(s string) int {
	// 摘抄下面这段代码
	// https://github.com/shimbe/SilverBird/blob/795f34a5c7823b6c21d4d0873a36a8b943951eba/SilverBirdClient/src/org/sb/client/shizimily7/PlayerUtil.java#L110
	// 该方案用到 http://ara.moo.jp/mjhmr/ ,这个软件里面的 ShantenTable.txt 文件
	// 使用查表法计算向听数,但是对于小于13张牌的情况仍然不能很好计算向听数
	// 因此最好还是用外层那个计算向听数的方案

	var (
		min     int
		tile    = make([]int, 0, 14)
		pairs   []int
		tileSet = make(map[int]struct{})
	)
	for i, v := range util.StrToTile(&s) {
		if v > 0 {
			for j := 0; j < v; j++ {
				tile = append(tile, i)
			}
			if v == 2 {
				pairs = append(pairs, i)
			}
			tileSet[i] = struct{}{}
		}
	}

	normalNum := calcTableScore(tile)
	for _, pair := range pairs {
		var tmpTile []int
		for _, vt := range tile {
			if vt != pair {
				tmpTile = append(tmpTile, vt)
			}
		}
		if min = calcTableScore(tmpTile) - 1; min < normalNum {
			normalNum = min
		}
	}

	// 计算七对的向听数
	if min = len(tileSet); 7 > min {
		min = 6 - len(pairs) + 7 - min
	} else {
		min = 6 - len(pairs)
	}

	if min < normalNum {
		return min
	}
	return normalNum
}

func calcTableScore(tile []int) int {
	var tmpTile []int
	table := 0

	for _, t := range tile {
		if len(tmpTile) > 0 {
			tt := tmpTile[len(tmpTile)-1]
			if util.IsZ(t) && t != tt || (t-tt) > 2 || !util.IsSameType(t, tt) {
				table += getTable(tmpTile)
				tmpTile = tmpTile[:0]
			}
		}
		tmpTile = append(tmpTile, t)
	}
	table += getTable(tmpTile)

	a1, a2, b1, b2 := table/1000, (table%1000)/100, (table%100)/10, table%10

	if a1+a2 > 4 {
		a2 = 4 - a1
	}
	if b1+b2 > 4 {
		b2 = 4 - b1
	}

	a1 = 8 - a1*2 - a2
	b1 = 8 - b1*2 - b2
	if a1 < b1 {
		return a1
	}
	return b1
}

func getTable(tile []int) int {
	x, last := 0, 1000
	for _, v := range tile {
		diff := v - last
		if diff > 1 {
			x *= 100
		} else if diff == 1 {
			x *= 10
		}
		x++
		last = v
	}
	// 查表,不存在返回0,存在返回对应值
	return shanTenMap[x]
}
