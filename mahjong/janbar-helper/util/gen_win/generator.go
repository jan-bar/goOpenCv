package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
)

var special bool

func main() {
	// go run generator.go -s
	flag.BoolVar(&special, "s", false, "not calc special card")
	flag.Parse()

	err := genWin()
	if err != nil {
		panic(err)
	}
}

func genWin() error {
	// 生成指定源码到指定目录
	const winGo = "../win.txt"
	fw, err := os.Create(winGo)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer fw.Close()

	var b bytes.Buffer
	st := genMahjongTable()
	for i, g := range st.Groups {
		b.Reset()
		// 转换为16进制,方便观察,同时节省空间
		b.WriteString(strconv.FormatUint(uint64(g.Key), 16))
		for _, gr := range g.Result {
			b.WriteByte(' ')
			b.WriteString(strconv.FormatUint(uint64(gr), 16))
		}
		if i+1 < len(st.Groups) {
			b.WriteByte('\n')
		}
		_, err = fw.Write(b.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

// 原文地址: https://blog.csdn.net/ywloveb/article/details/86298697
// 一份归档: https://github.com/piaohua/mjalgorism

// 一组牌型
type mahjongGroup struct {
	// 牌型Key值
	Key uint32
	// 牌型结果
	Result []uint32
}

// 麻将牌型表
type mahjongTable struct {
	Groups []*mahjongGroup
}

func genMahjongTable() *mahjongTable {
	array := analyseQiDui() // 七对

	// 14张牌需要判断的牌型
	array = append(array, genData([][]int{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}, {1, 1, 1}, {2}})...)
	array = append(array, genData([][]int{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}, {3}, {2}})...)
	array = append(array, genData([][]int{{1, 1, 1}, {1, 1, 1}, {3}, {3}, {2}})...)
	array = append(array, genData([][]int{{1, 1, 1}, {3}, {3}, {3}, {2}})...)
	array = append(array, genData([][]int{{3}, {3}, {3}, {3}, {2}})...)

	// 去掉1个杠后11张牌的牌型
	array = append(array, genData([][]int{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}, {2}})...)
	array = append(array, genData([][]int{{1, 1, 1}, {1, 1, 1}, {3}, {2}})...)
	array = append(array, genData([][]int{{1, 1, 1}, {3}, {3}, {2}})...)
	array = append(array, genData([][]int{{3}, {3}, {3}, {2}})...)

	// 去掉2个杠后8张牌的牌型
	array = append(array, genData([][]int{{1, 1, 1}, {1, 1, 1}, {2}})...)
	array = append(array, genData([][]int{{1, 1, 1}, {3}, {2}})...)
	array = append(array, genData([][]int{{3}, {3}, {2}})...)

	// 去掉3个杠后5张牌的牌型
	array = append(array, genData([][]int{{1, 1, 1}, {2}})...)
	array = append(array, genData([][]int{{3}, {2}})...)

	// 去掉4个杠后,单吊将牌
	array = append(array, genData([][]int{{2}})...)

	table := &mahjongTable{}
	keyMap := make(map[uint32]struct{})
	for _, v := range array {
		if _, found := keyMap[v.Key]; !found {
			keyMap[v.Key] = struct{}{} // 多次去重
			table.Groups = append(table.Groups, v)
		}
	}

	// 计算龙七对这种特殊牌型,该牌型总数也是14张
	// 且4张相同牌不能杠出和碰出
	if special {
		var (
			legal [][][]int // 5个对子+4张相同牌(这个牌不能碰和杠出)
			cmp   = [5]int{14, 0, 5, 0, 1}
			parr  = ptn([][]int{{2}, {2}, {2}, {2}, {2}, {4}})
		)
		for _, v := range parr {
			var cnt [5]int
			for _, vv := range v {
				for _, vvv := range vv {
					if vvv > 0 {
						cnt[vvv]++
						cnt[0] += vvv
					}
				}
			}
			if cnt == cmp {
				legal = append(legal, v)
			}
		}
		for _, v := range legal {
			key := calKey(v)
			if _, found := keyMap[key]; !found {
				keyMap[key] = struct{}{}

				var tileNum, ret uint32
				for i := 0; i < len(v); i++ {
					for j := 0; j < len(v[i]); j++ {
						if v[i][j] == 4 {
							ret = tileNum << 6 // 存到雀头将的位置
						}
						tileNum++
					}
				}
				table.Groups = append(table.Groups, &mahjongGroup{
					Key: key,
					// 龙七对也算到七对里面,4张相同的牌存到雀头位置
					Result: []uint32{ret | 1<<26},
				})
			}
		}
	}

	// 结果按照key升序排序
	sort.Slice(table.Groups, func(i, j int) bool {
		return table.Groups[i].Key < table.Groups[j].Key
	})
	return table
}

// 分析七对
//goland:noinspection ALL
func analyseQiDui() []*mahjongGroup {
	// 先输入七 对子
	chidui := ptn([][]int{{2}, {2}, {2}, {2}, {2}, {2}, {2}})
	newChidui := make([][][]int, 0)
	for _, v := range chidui {
		valid := true
		for _, vv := range v {
			for _, vvv := range vv {
				if vvv != 2 {
					valid = false
					break
				}
			}
			if !valid {
				break
			}
		}
		if valid { // 只保留全是2个的组合
			newChidui = append(newChidui, v)
		}
	}
	var (
		keyMap = make(map[uint32]struct{})
		array  []*mahjongGroup
	)
	for _, v := range newChidui {
		key := calKey(v)
		if _, found := keyMap[key]; !found {
			keyMap[key] = struct{}{}
			array = append(array, &mahjongGroup{
				Key:    key,
				Result: findHaiPos(v),
			})
		}
	}
	return array
}

func genData(a [][]int) []*mahjongGroup {
	r := ptn(a)
	array := make([]*mahjongGroup, 0, len(r))
	for _, v := range r {
		array = append(array, &mahjongGroup{
			Key:    calKey(v),
			Result: findHaiPos(v),
		})
	}
	return array
}

func copyArr(src [][]int) [][]int {
	ret := make([][]int, len(src))
	for i, v := range src {
		ret[i] = make([]int, len(v))
		for iv, vv := range v {
			ret[i][iv] = vv
		}
	}
	return ret
}

func ptn(a [][]int) (ret [][][]int) {
	if a == nil {
		ret = make([][][]int, 0)
		return
	}
	size := len(a)
	if size <= 1 {
		ret = [][][]int{a}
		return
	}
	ret = append(ret, perms(a)...)
	keyMap := make(map[string]struct{})
	for i := 0; i < size; i++ {
		for j := i + 1; j < size; j++ {
			key := fmt.Sprintf("%v0%v", a[i], a[j])
			if _, found := keyMap[key]; found {
				continue // 两个元素中间补0时重复则跳过
			}
			keyMap[key] = struct{}{}

			tMap := make(map[string]struct{})
			lj := len(a[j])
			al := len(a[i]) + lj
			for k := 0; k <= al; k++ {
				t := make([]int, al+lj)
				for l := lj; l < al; l++ {
					t[l] = a[i][l-lj]
				}
				for m := 0; m < lj; m++ {
					t[k+m] += a[j][m]
				}
				var tmp []int
				valid := true
				for _, v := range t {
					if v > 4 {
						valid = false
						break // 大于4的不合法
					}
					if v > 0 { // 删除为0的元素
						tmp = append(tmp, v)
					}
				}
				if !valid || len(tmp) > 9 || len(tmp) <= 0 {
					continue
				}
				tmpStr := fmt.Sprintf("%v", tmp)
				if _, found := tMap[tmpStr]; !found {
					tMap[tmpStr] = struct{}{}
					b := copyArr(a) // 克隆切片,不能直接用copy会有引用问题
					b = append(b[:i], b[i+1:]...)
					b = append(b[:j-1], b[j:]...)
					c := [][]int{tmp}
					if len(b) > 0 {
						c = append(c, b...)
					}
					ret = append(ret, ptn(c)...)
				}
			}
		}
	}
	return
}

func perms(a [][]int) (ret [][][]int) {
	ret = make([][][]int, 0)
	r := perm(a) // 将排列结果去重后返回
	keyM := make(map[string]struct{})
	for _, v := range r {
		key := fmt.Sprintf("%v", v)
		if _, found := keyM[key]; !found {
			ret = append(ret, v)
			keyM[key] = struct{}{}
		}
	}
	return
}

// 将a进行全排列
func perm(a [][]int) (ret [][][]int) {
	ret = make([][][]int, 0)
	if a == nil {
		return
	}
	if len(a) <= 1 {
		ret = append(ret, a)
		return
	}
	for k, v := range a {
		tmp := copyArr(a)
		tmp = append(tmp[:k], tmp[k+1:]...)
		// 将a去掉其中一项,然后再和去掉的v组合
		// 最终将a的每个元素全排列结果返回
		for _, tv := range perm(tmp) {
			tv = append([][]int{v}, tv...)
			ret = append(ret, tv)
		}
	}
	return
}

/*
按照如下编码,下一个数字为'0'则添加'10',非'0'时情况下添加'0'
[1] -> []       [10] -> [10]
[2] -> [11]     [20] -> [1110]
[3] -> [1111]   [30] -> [111110]
[4] -> [111111] [40] -> [11111110]

bit位需要从右往左看,2个数字中间会塞0,研究哈弗曼编码
2        -> 0x7     -> 0111
23       -> 0xFB    -> 1111 1011
32       -> 0xEF    -> 1110 1111
233      -> 0x1F7B  -> 0001 1111 0111 1011
2303     -> 0x3EFB  -> 0011 1110 1111 1011
21110330 -> 0x1F7A3 -> 0001 1111 0111 1010 0011
*/
func calKey(a [][]int) (ret uint32) {
	l := -1
	ret = 0
	for _, b := range a {
		for _, v := range b {
			l++
			switch v {
			case 2:
				ret |= 0x3 << uint(l)
				l += 2
			case 3:
				ret |= 0xF << uint(l)
				l += 4
			case 4:
				ret |= 0x3F << uint(l)
				l += 6
			}
		}
		ret |= 0x1 << uint(l)
		l++
	}
	return
}

// 分析牌型(暴力拆解)
// 3bit  0: 刻子数(0-4)
// 3bit  3: 顺子数(0-4)
// 4bit  6: 雀头位置(1-13)
// 4bit 10: 面子位置1(0-13) 刻子在前,顺子在后
// 4bit 14: 面子位置2(0-13)
// 4bit 18: 面子位置3(0-13)
// 4bit 22: 面子位置4(0-13)
// 1bit 26: 七 对子
// 1bit 27: 九莲宝灯
// 1bit 28: 一气通贯
// 1bit 29: 两杯口
// 1bit 30: 一杯口
//goland:noinspection ALL
func findHaiPos(a [][]int) []uint32 {
	size := len(a)
	p_atama := 0
	ret_map := map[int]struct{}{}
	ret_array := make([]uint32, 0)
	for i := 0; i < size; i++ {
		for j := 0; j < len(a[i]); j++ {
			// 拆解将牌
			if a[i][j] >= 2 {
				for kotsu_shuntus := 0; kotsu_shuntus <= 1; kotsu_shuntus++ {
					t := copyArr(a)
					t[i][j] -= 2 // 复制a,取雀头
					p := 0
					p_kotsu := make([]int, 0)
					p_shuntsu := make([]int, 0)
					for k := 0; k < len(t); k++ {
						for m := 0; m < len(t[k]); m++ {
							if kotsu_shuntus == 0 {
								// 先取刻子
								if t[k][m] >= 3 {
									t[k][m] -= 3
									p_kotsu = append(p_kotsu, p)
								}
								// 再取顺子
								for len(t[k])-m >= 3 &&
									t[k][m] >= 1 &&
									t[k][m+1] >= 1 &&
									t[k][m+2] >= 1 {
									t[k][m] -= 1
									t[k][m+1] -= 1
									t[k][m+2] -= 1
									p_shuntsu = append(p_shuntsu, p)
								}
							} else {
								// 先取顺子
								for len(t[k])-m >= 3 &&
									t[k][m] >= 1 &&
									t[k][m+1] >= 1 &&
									t[k][m+2] >= 1 {
									t[k][m] -= 1
									t[k][m+1] -= 1
									t[k][m+2] -= 1
									p_shuntsu = append(p_shuntsu, p)
								}
								// 再取刻子
								if t[k][m] >= 3 {
									t[k][m] -= 3
									p_kotsu = append(p_kotsu, p)
								}
							}
							p += 1
						}
					}

					hu := true
					for _, v := range t {
						for _, vv := range v {
							if vv != 0 {
								hu = false
								break
							}
						}
					}
					if hu { // 所有牌取完则胡牌,进行特殊值判定
						ret := len(p_kotsu) + (len(p_shuntsu) << 3) + (p_atama << 6)
						l := 10
						for _, ke := range p_kotsu {
							ret |= ke << uint(l)
							l += 4
						}
						for _, shun := range p_shuntsu {
							ret |= shun << uint(l)
							l += 4
						}
						if len(a) == 1 {
							// 九莲宝灯
							key := fmt.Sprintf("%v", a[0])
							if key == "[4 1 1 1 1 1 1 1 3]" ||
								key == "[3 2 1 1 1 1 1 1 3]" ||
								key == "[3 1 2 1 1 1 1 1 3]" ||
								key == "[3 1 1 2 1 1 1 1 3]" ||
								key == "[3 1 1 1 2 1 1 1 3]" ||
								key == "[3 1 1 1 1 2 1 1 3]" ||
								key == "[3 1 1 1 1 1 2 1 3]" ||
								key == "[3 1 1 1 1 1 1 2 3]" ||
								key == "[3 1 1 1 1 1 1 1 4]" {
								ret |= 1 << 27
							}
						}
						// 通天
						if len(a) <= 3 && len(p_shuntsu) >= 3 {
							p_ikki := 0
							for _, c := range a {
								if len(c) == 9 {
									b_ikki1 := false
									b_ikki2 := false
									b_ikki3 := false
									for _, x_ikki := range p_shuntsu {
										if x_ikki == p_ikki {
											b_ikki1 = true
										}
										if x_ikki == p_ikki+3 {
											b_ikki2 = true
										}
										if x_ikki == p_ikki+6 {
											b_ikki3 = true
										}
									}
									if b_ikki1 && b_ikki2 && b_ikki3 {
										ret |= 1 << 28
									}
								}
								p_ikki += len(c)
							}
						}
						// 二杯口
						if len(p_shuntsu) == 4 &&
							p_shuntsu[0] == p_shuntsu[1] &&
							p_shuntsu[2] == p_shuntsu[3] {
							ret |= 1 << 29
						} else if len(p_shuntsu) >= 2 &&
							(len(p_kotsu)+len(p_shuntsu)) == 4 {
							// 一杯口
							m := map[int]struct{}{}
							for _, v := range p_shuntsu {
								m[v] = struct{}{}
							}
							if len(p_shuntsu)-len(m) >= 1 {
								ret |= 1 << 30
							}
						}

						_, ok := ret_map[ret]
						if !ok { // 结果去重,先后顺序和回溯法保持一致
							ret_array = append(ret_array, uint32(ret))
							ret_map[ret] = struct{}{}
						}
					}
				}
			}
			p_atama++
		}
	}

	if len(ret_array) > 0 {
		return ret_array
	}

	total := 0 // 七对子
	for _, v := range a {
		for _, vv := range v {
			if vv != 2 {
				return nil // 不是对子
			}
			total += vv
		}
	}
	if total == 14 {
		return []uint32{1 << 26} // 七对没有雀头
	}
	return nil
}
