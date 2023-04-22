package analyze

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jan-bar/goOpenCv/mahjong/janbar-helper/shanten"
	"github.com/jan-bar/goOpenCv/mahjong/janbar-helper/util"
)

/*
计算向听数和打牌: https://github.com/skk294/mahjongg
*/

// Analyze 分析麻将
// tiles34不要包含杠和碰的牌,只计算能打出的牌
func Analyze(tiles34 []int) {
	all := util.CheckTile(tiles34)

	switch all % 3 {
	case 1: // 计算摸牌
		CalcInCard13(tiles34)
	case 2: // 计算打牌
		CalcOutCard14(tiles34)
	default:
		fmt.Printf("牌: %s,数量: %d,%d 不合法\n",
			util.TileToStr(tiles34), all, all%3) // 参数错误
	}
}

func CalcInCard13(tile []int) {
	left := util.LeftTiles34(tile)
	st, _ := shanten.CalcShanTenTile(tile)

	res := _search13(tile, left, st, _stopShanten(st))
	res.analysis(tile)
}

func CalcOutCard14(tile []int) {
	st, _ := shanten.CalcShanTenTile(tile)
	if util.Tile34IsHu(tile) {
		fmt.Println("已胡牌")
		return
	}

	left := util.LeftTiles34(tile)
	res := _search14(tile, left, st, _stopShanten(st))
	res.analysis(tile)
	return
}

const (
	shantenStateH = -1 // 胡牌向听数
	shantenStateT = 0  // 听牌向听数
)

func _stopShanten(shanten int) int {
	if shanten >= 3 {
		return shanten - 1
	}
	return shanten - 2
}

type (
	SearchNode struct {
		search13  bool // true: 表示摸牌,false: 表示打牌
		shanten   int  // 保存向听数
		tile, num int  // 牌的数值,牌的个数
		max       int  // count(子节点数), max(剩余牌数)
		count     int  // 当前节点所有子树节点个数
		children  SortNode
	}

	SortNode []*SearchNode // 子节点排序
)

func (sn SortNode) Len() int {
	return len(sn)
}
func (sn SortNode) Less(i, j int) bool {
	si, sj := sn[i].count, sn[j].count // 方案数越多越靠前
	if si == sj {
		si, sj = sn[i].max, sn[j].max // 剩余胡牌多或手牌少靠前
		if si == sj {
			si, sj = sn[i].num, sn[j].num // 手牌越多靠前,可以碰和杠
		}
	}
	return si > sj
}
func (sn SortNode) Swap(i, j int) {
	sn[i], sn[j] = sn[j], sn[i]
}

/*
计算打牌: https://tenhou.net/2/
得出结果和我通过统计得出结果一致
*/
func (sn *SearchNode) analysis(tile []int) {
	var (
		b strings.Builder
		s = util.TileToStr(tile)
	)
	b.WriteString(s)
	b.WriteByte('\n')
	for _, ch := range sn.children {
		nv := util.NumToTile(ch.tile)
		if sn.search13 {
			b.WriteString("摸")
			b.WriteByte(nv[0])
			b.WriteByte(nv[1])
			if ch.shanten == shantenStateH {
				b.WriteString("胡了,剩余: ")
				b.WriteString(strconv.Itoa(ch.max))
			} else {
				b.WriteString(",方案数: ")
				b.WriteString(strconv.Itoa(ch.count))
			}
			b.WriteByte('\n')
		} else {
			b.WriteString("打")
			b.WriteByte(nv[0])
			b.WriteByte(nv[1])
			// 胡了ch.max表示胡牌总数,没胡ch.max表示
			b.WriteString(fmt.Sprintf(" 数(%4d,%2d) 摸[", ch.count, ch.max))
			for chi, chh := range ch.children {
				if chi > 0 {
					b.WriteByte(',')
				}
				nvv := util.NumToTile(chh.tile)
				b.WriteByte(nvv[0])
				b.WriteByte(nvv[1])
				b.WriteByte(' ')
				if chh.shanten == shantenStateH {
					b.WriteString("胡(")
					b.WriteString(strconv.Itoa(chh.max))
				} else {
					switch chh.num {
					case 8: // 杠
						b.WriteString(fmt.Sprintf("杠(4,%3d", chh.count))
					case 7: // 碰
						b.WriteString(fmt.Sprintf("碰(3,%3d", chh.count))
					default:
						b.WriteString(fmt.Sprintf("数(%d,%3d", chh.num, chh.count))
					}
				}
				b.WriteString(")")
			}
			b.WriteString("]\n")
		}
	}

	fmt.Println(b.String())
}

func sortChildren(sn *SearchNode, st SortNode) *SearchNode {
	if len(st) > 0 {
		sort.Sort(st)
		for _, ch := range st {
			sn.count += ch.count
		}
		sn.children = st
	} else {
		sn.count = 1
	}
	return sn
}

func _search13(hand, left []int, cur, stop int) *SearchNode {
	var (
		tmp int
		isT = cur == shantenStateT // 当前听牌
		s14 = cur-1 >= stop        // 继续搜索14张牌

		children SortNode
	)

	for i := 0; i < util.TileMax; i++ {
		if hand[i] == 4 || left[i] == 0 {
			// 已经4张,没有剩余牌(剔除不能完成的方案),不能再摸牌
			continue
		}

		hand[i]++
		if isT {
			if util.Tile34IsHu(hand) {
				// 摸这张牌可以胡,该牌剩余越多排前面
				children = append(children, &SearchNode{
					shanten: shantenStateH, // 表示已胡牌
					tile:    i, count: 1,
					num: hand[i], // 记录当前牌数
					max: left[i], // 记录剩余牌数
				})
				tmp += left[i] // 记录胡牌总数
			}
		} else if s14 {
			if st, _ := shanten.CalcShanTenTile(hand); st < cur {
				left[i]-- // 摸牌必须减小向听才是合理摸牌
				sn := _search14(hand, left, cur-1, stop)
				sn.tile, sn.num = i, hand[i] // 记录牌和当前牌数
				children = append(children, sn)
				left[i]++
			}

			if hand[i] >= 3 {
				tmp = hand[i]
				hand[i] = 0
				// 摸牌杠碰保持向听或减小向听,合理摸牌
				if st, _ := shanten.CalcShanTenTile(hand); st <= cur {
					left[i]--
					sn := _search14(hand, left, cur-1, stop)
					sn.tile, sn.num = i, tmp+4 // 使向听减小的杠碰排序靠前
					children = append(children, sn)
					left[i]++
				}
				hand[i] = tmp
			}
		}
		hand[i]--
	}

	return sortChildren(&SearchNode{
		shanten:  cur,
		search13: true,
		max:      tmp,
	}, children)
}

func _search14(hand, left []int, cur, stop int) *SearchNode {
	var children SortNode

	for i := 0; i < util.TileMax; i++ {
		if hand[i] == 0 {
			continue // 没有可打出的牌
		}

		hand[i]--
		if st, _ := shanten.CalcShanTenTile(hand); st <= cur {
			sn := _search13(hand, left, cur, stop)
			sn.tile, sn.num = i, hand[i] // 合理打牌
			if sn.max == 0 {
				sn.max = 3 - hand[i]
			}
			children = append(children, sn)
		}
		hand[i]++
	}

	return sortChildren(&SearchNode{shanten: cur}, children)
}
