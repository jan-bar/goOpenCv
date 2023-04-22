package util

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const TileMax = 34

func StrToTile(s *string) []int {
	tile := make([]int, TileMax)
	if s == nil || len(*s) < 2 {
		*s = "" // 至少2个字符1张牌,不满足则置为空
		return tile
	}

	var (
		ok = false
		i  = len(*s) - 1
		si = -1
		ni int
	)
	for ; i >= 0; i-- {
		switch c := (*s)[i]; c {
		case 'm':
			si = 0 // [0,8]m = [1,9]万
		case 'p':
			si = 9 // [9,17]p = [1,9]筒
		case 's':
			si = 18 // [18,26]s = [1,9]条
		case 'z':
			si = 27 // [27,33]z = [1,7] = 东,南,西,北,白,发,中
		case '0':
			c = '5' // 替换赤宝牌(0m->5m,0p->5p,0s->5s)
			fallthrough
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if si >= 0 {
				if ni = int(c-'1') + si; ni < TileMax && tile[ni] < 4 {
					tile[ni]++
					ok = true
				}
			}
		}
	}

	if ok {
		*s = TileToStr(tile) // 有牌时序列化并排序
	} else {
		*s = "" // 无牌时置为空
	}
	return tile
}

func IsM(v int) bool { return v >= 0 && v <= 8 }
func IsP(v int) bool { return v >= 9 && v <= 17 }
func IsS(v int) bool { return v >= 18 && v <= 26 }
func IsZ(v int) bool { return v >= 27 && v <= 33 }
func IsSameType(a, b int) bool {
	return IsM(a) && IsM(b) || IsP(a) && IsP(b) || IsS(a) && IsS(b) || IsZ(a) && IsZ(b)
}

func LeftTiles34(tile34 []int) []int {
	left := make([]int, TileMax)
	for i, v := range tile34 {
		left[i] = 4 - v // 手牌以外可以摸的牌
	}
	return left
}

type TileLine struct {
	Tile []int
	Line []string
}

func LoadTileFromFile(p string) (res []TileLine) {
	fr, err := os.Open(p)
	if err != nil {
		panic(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer fr.Close()

	br := bufio.NewScanner(fr)
	for br.Scan() {
		s := strings.Split(br.Text(), ",")
		if t := StrToTile(&s[0]); s[0] != "" {
			for i := 1; i < len(s); i++ {
				s[i] = strings.TrimSpace(s[i]) // 处理除了第1个以为的元素
			}
			// 第1个转换为手牌,Line保存全部数据
			res = append(res, TileLine{Tile: t, Line: s})
		}
	}
	if err = br.Err(); err != nil {
		panic(err)
	}
	return
}

func TileToStr(tile []int) string {
	var (
		i, j int
		lt   = len(tile) // 支持lt<34的不完全入参
		tmp  strings.Builder
	)

	ok := false // 拼接万牌
	for i = 0; i < 9 && i < lt; i++ {
		for j = tile[i]; j > 0; j-- {
			tmp.WriteByte(byte(i + '1'))
			ok = true
		}
	}
	if ok {
		tmp.WriteByte('m')
	}

	ok = false // 拼接筒牌
	for i = 9; i < 18 && i < lt; i++ {
		for j = tile[i]; j > 0; j-- {
			tmp.WriteByte(byte(i - 9 + '1'))
			ok = true
		}
	}
	if ok {
		tmp.WriteByte('p')
	}

	ok = false // 拼接条牌
	for i = 18; i < 27 && i < lt; i++ {
		for j = tile[i]; j > 0; j-- {
			tmp.WriteByte(byte(i - 18 + '1'))
			ok = true
		}
	}
	if ok {
		tmp.WriteByte('s')
	}

	ok = false // 拼接字牌
	for i = 27; i < 34 && i < lt; i++ {
		for j = tile[i]; j > 0; j-- {
			tmp.WriteByte(byte(i - 27 + '1'))
			ok = true
		}
	}
	if ok {
		tmp.WriteByte('z')
	}
	return tmp.String()
}

func CheckTile(tile []int) int {
	if len(tile) < TileMax {
		return -1 // 参数错误
	}

	all, max := 0, 14 // 牌总数,允许最多牌数
	for i := 0; i < TileMax; i++ {
		switch tile[i] {
		case 0:
		case 4:
			// 1个杠允许总牌数多1张
			// 将杠减1个变成刻子,向听数结果一样,因此支持4张牌
			max++
			fallthrough
		case 1, 2, 3:
			all += tile[i]
		default:
			return -2 // 参数错误
		}
	}
	if all > max {
		return -3 // 超过胡牌最大值
	}
	return all // 检查正常,返回牌总数
}

func NumToTile(v int) []byte {
	if v < 0 || v >= TileMax {
		return []byte{'0', '-'} // 不合法数据
	}
	if v >= 27 {
		return []byte{byte(v - 27 + '1'), 'z'}
	}
	if v >= 18 {
		return []byte{byte(v - 18 + '1'), 's'}
	}
	if v >= 9 {
		return []byte{byte(v - 9 + '1'), 'p'}
	}
	return []byte{byte(v + '1'), 'm'}
}

func NumToTileStr(v int) string {
	if v < 0 || v >= TileMax {
		return "0-" // 不合法数据
	}
	if v >= 27 {
		return strconv.Itoa(v-26) + "z"
	}
	if v >= 18 {
		return strconv.Itoa(v-17) + "s"
	}
	if v >= 9 {
		return strconv.Itoa(v-8) + "p"
	}
	return strconv.Itoa(v+1) + "m"
}

var (
	//go:embed win.txt
	winData  []byte
	winTable = loadWinTableData(bytes.NewReader(winData))
)

func loadWinTableData(r io.Reader) map[int][]int {
	var (
		re = make(map[int][]int, 9540)
		br = bufio.NewScanner(r)
		kv int
	)
	for br.Scan() {
		arr := strings.Fields(br.Text())
		lvi := len(arr) - 1
		if lvi < 1 {
			continue
		}

		val := make([]int, 0, lvi)
		for i, v := range arr {
			tv, ev := strconv.ParseUint(v, 16, 0)
			if ev == nil {
				if i == 0 {
					kv = int(tv) // 第1个是key
				} else {
					val = append(val, int(tv)) // 后面的是多种牌型
				}
			}
		}

		if len(val) > 0 {
			re[kv] = val
		}
	}
	if err := br.Err(); err != nil {
		panic(err)
	}
	return re
}

func _calKey(tiles34 []int, res *[]*MahjongResult) (isHu bool) {
	// 杠牌和碰牌属于已经出过的特殊牌,判断胡牌时不要传进来
	// http://hp.vector.co.jp/authors/VA046927/mjscore/AgariIndex.java
	var (
		i, j, ti, c, key int

		bitPos, idx = -1, -1 // 数牌
		prevInHand  bool

		tiles14 = make([]int, 14)
	)
	for i = 0; i < 3; i++ {
		prevInHand = false // 上张牌是否在手牌中
		for j = 0; j < 9; j++ {
			idx++
			if c = tiles34[idx]; c > 0 {
				tiles14[ti] = idx
				ti++

				prevInHand = true
				bitPos++
				switch c {
				case 2:
					key |= 0x3 << uint(bitPos)
					bitPos += 2
				case 3:
					key |= 0xF << uint(bitPos)
					bitPos += 4
				case 4:
					key |= 0x3F << uint(bitPos)
					bitPos += 6
				}
			} else if prevInHand {
				prevInHand = false
				key |= 0x1 << uint(bitPos)
				bitPos++
			}
		}
		if prevInHand {
			key |= 0x1 << uint(bitPos)
			bitPos++
		}
	}

	// 字牌
	for i = 27; i < 34; i++ {
		if c = tiles34[i]; c > 0 {
			tiles14[ti] = idx
			ti++

			bitPos++
			switch c {
			case 2:
				key |= 0x3 << uint(bitPos)
				bitPos += 2
			case 3:
				key |= 0xF << uint(bitPos)
				bitPos += 4
			case 4:
				key |= 0x3F << uint(bitPos)
				bitPos += 6
			}
			key |= 0x1 << uint(bitPos)
			bitPos++
		}
	}

	var tileType []int
	tileType, isHu = winTable[key]
	if !isHu || res == nil {
		return
	}

	// 查表法获取牌型,需要返回时才进行下面的处理
	results := make([]*MahjongResult, 0, len(tileType))
	for _, tt := range tileType {
		tmp := &MahjongResult{
			Jiang:   tiles14[(tt>>6)&0xF], // 雀头
			NumKe:   tt & 0x7,             // 刻子
			NumShun: (tt >> 3) & 0x7,      // 顺子的第一张牌
		}

		tmp.ArrayKe = make([]int, tmp.NumKe)
		for i = 0; i < tmp.NumKe; i++ {
			tmp.ArrayKe[i] = tiles14[(tt>>uint(10+i*4))&0xF]
		}

		tmp.ArrayShun = make([]int, tmp.NumShun)
		for i = 0; i < tmp.NumShun; i++ {
			tmp.ArrayShun[i] = tiles14[(tt>>uint(10+(tmp.NumKe+i)*4))&0xF]
		}

		tmp.IsQiDui = tt&(1<<26) != 0
		tmp.IsChurchmenPout = tt&(1<<27) != 0
		tmp.IsTongTan = tt&(1<<28) != 0
		tmp.IsRyanPeiKou = tt&(1<<29) != 0
		tmp.IsIiPeiKou = tt&(1<<30) != 0
		results = append(results, tmp)
	}
	*res = results
	return
}

func Tile34IsHu(tiles34 []int) bool {
	// 3k+2 张牌,是否和牌(不检测 国士无双)
	return _calKey(tiles34, nil)
}

type MahjongResult struct {
	NumKe     int   // 刻子数量
	NumShun   int   // 顺子数量
	Jiang     int   // 将牌值(0-8{1-9万},9-17{1-9筒},18-26{1-9条})
	ArrayKe   []int // 刻子数组
	ArrayShun []int // 顺子数组

	IsQiDui         bool // 七对,包含龙七对(此时将牌值存了4张龙头牌)
	IsChurchmenPout bool // 九莲宝灯
	IsTongTan       bool // 通天(注意: 未考虑副露)
	IsRyanPeiKou    bool // 两杯口(IsRyanPeiKou == true 时 IsIiPeiKou == false)
	IsIiPeiKou      bool // 一杯口
}

func (mr *MahjongResult) String() string {
	b := bytes.NewBufferString("{")
	nv := NumToTile(mr.Jiang)
	if nv[1] == '-' {
		return "" // 将不合法
	}
	b.WriteByte(nv[0])
	b.WriteByte(nv[0])

	if mr.IsQiDui {
		if mr.Jiang > 0 {
			b.WriteByte(nv[0])
			b.WriteByte(nv[0])
			b.WriteString(",龙七对}")
			return b.String()
		}
		return "七对"
	}

	b.WriteByte(nv[1])
	b.WriteByte(',')
	for _, v := range mr.ArrayKe {
		nv = NumToTile(v)
		b.WriteByte(nv[0])
		b.WriteByte(nv[0])
		b.WriteByte(nv[0])
		b.WriteByte(nv[1])
		b.WriteByte(',')
	}
	for _, v := range mr.ArrayShun {
		nv = NumToTile(v)
		b.WriteByte(nv[0])
		b.WriteByte(nv[0] + 1)
		b.WriteByte(nv[0] + 2)
		b.WriteByte(nv[1])
		b.WriteByte(',')
	}
	// 移除最后1个','号,并将结果括起来
	b.Truncate(b.Len() - 1)
	b.WriteByte('}')
	return b.String()
}

func Tiles34Result(tiles34 []int) (res []*MahjongResult) {
	_calKey(tiles34, &res)
	return
}

func Tiles34Backtrack(tiles34 []int) (res []*MahjongResult) {
	// 计算麻将胡牌方案
	// 杠牌和碰牌属于已经出过的特殊牌,判断胡牌时不要传进来
	// http://hp.vector.co.jp/authors/VA046927/mjscore/mjalgorism.html
	// 回溯法,下面方案参照回溯法
	// http://hp.vector.co.jp/authors/VA046927/mjscore/AgariBacktrack.java
	var (
		all int
		cnt [5][]int
	)
	for i, v := range tiles34 {
		if v > 0 {
			cnt[v] = append(cnt[v], i)
			all += v
		}
	}
	if all == 14 {
		if len(cnt[2]) == 7 {
			return []*MahjongResult{{
				IsQiDui: true, // 七对
			}}
		}
		if len(cnt[2]) == 5 && len(cnt[4]) == 1 {
			return []*MahjongResult{{
				IsQiDui: true,
				Jiang:   cnt[4][0], // 龙七对,4张牌做将(没有杠出和碰出)
			}}
		}
	}

	var (
		appNum = []func(tp []int, k, s *[]int){
			func(tp []int, k, s *[]int) {
				for j := 0; j < TileMax; j++ {
					if tp[j] >= 3 {
						tp[j] -= 3 // 取刻子
						*k = append(*k, j)
					}
				}
			},
			func(tp []int, k, s *[]int) {
				var a, b, c int
				for a = 0; a < 3; a++ {
					for b = 0; b < 7; {
						c = 9*a + b
						if tp[c] >= 1 && tp[c+1] >= 1 && tp[c+2] >= 1 {
							tp[c]--
							tp[c+1]--
							tp[c+2]-- // 取顺子
							*s = append(*s, c)
						} else {
							b++
						}
					}
				}
			},
		}

		tp = make([]int, len(tiles34))
		mp = make(map[string]struct{})
	)
	for i := 0; i < TileMax; i++ {
		if tiles34[i] < 2 {
			continue // 跳过不能做雀头的牌
		}

		// 取刻子和取顺子,按照先后顺序进行排列
		for _, an := range [][]int{{0, 1}, {1, 0}} {
			copy(tp, tiles34)
			tp[i] -= 2 // 取雀头

			var keNum, shunNum []int
			for _, anv := range an {
				appNum[anv](tp, &keNum, &shunNum)
			}

			ok := true
			for _, vt := range tp {
				if vt != 0 {
					ok = false
					break
				}
			}
			if ok { // 胡牌,记录牌型,结果去重
				key := fmt.Sprintf("%d%v%v", i, keNum, shunNum)
				if _, ok = mp[key]; !ok {
					res = append(res, &MahjongResult{
						NumKe:     len(keNum),
						NumShun:   len(shunNum),
						Jiang:     i,
						ArrayKe:   keNum,
						ArrayShun: shunNum,
					})
					mp[key] = struct{}{}
				}
			}
		}
	}
	return
}
