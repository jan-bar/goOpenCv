package shanten

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"strconv"

	"github.com/jan-bar/goOpenCv/mahjong/janbar-helper/util"
	"github.com/jan-bar/tools"
)

/* 一些计算向听数的文章
https://www.bilibili.com/read/cv10974292
 x=9-2*m-d+c-q
  x为向听数
  m为面子(刻子或顺子)的数量
  d为搭子(面子少一张)的数量
  c为超载数
    当m+d<=5时,c=0
    当m+d>5时,c=m+d—5
  q为雀头函数
    m+d<=4时,q＝1
    m+d>4时
      有雀头(对子)时,q=1
      没有雀头时,q=0

https://github.com/skk294/mahjongg  上面理论的实际代码
https://www.bilibili.com/read/cv20401264

https://github.com/ibukisaar/JapaneseMahjong

包含麻将扑克等各种算法,生成表,判断时查表
https://github.com/yuanfengyun/q_algorithm
*/

// 使用该项目生成的数据文件,用查表法计算(向听数+1,注意我的算法有-1逻辑)
// https://github.com/tomohxx/shanten-number
// 下面是相关项目和相关文章
// https://www.cnblogs.com/syui-terra/p/16673262.html
// https://github.com/comssa56/snc_js/blob/main/source/shanten.js
// https://github.com/terralian/fastmaj/blob/master/src/main/java/com/github/terralian/fastmaj/tehai/FastSyantenCalculator.java
var mps, mph = loadIndexTarGz()

// 数据来源: https://github.com/tomohxx/shanten-number
// index.tar.gz 3ae692f68b1097eaf572039c99e8d200
//go:embed index.tar.gz
var shanTenTarGz []byte

func loadIndexTarGz() (mps, mph [][]int) {
	gr, err := gzip.NewReader(bytes.NewReader(shanTenTarGz))
	if err != nil {
		panic(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer gr.Close()

	// 指定max大小,避免内存浪费
	read := func(r io.Reader, max int) ([][]int, error) {
		mp, br := make([][]int, 0, max), bufio.NewScanner(r)
		for br.Scan() {
			tmp := bytes.Fields(br.Bytes())
			mpi := make([]int, len(tmp))
			for i, vv := range tmp {
				mpi[i], err = strconv.Atoi(tools.BytesToString(vv))
				if err != nil {
					return nil, err
				}
			}
			mp = append(mp, mpi)
		}
		return mp, br.Err()
	}

	tr := tar.NewReader(gr) // 读取[xx.tar.gz]文件
	for {
		th, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return
			}
			panic(err)
		}

		switch th.Name {
		case "index_h.txt":
			mph, err = read(tr, 78125)
		case "index_s.txt":
			mps, err = read(tr, 1953125)
		}
		if err != nil {
			panic(err)
		}
	}
}

func CalcShanTenStr(ts *string) (retNum, retMode int) {
	return CalcShanTenTile(util.StrToTile(ts))
}

/*CalcShanTenTile
计算手牌向听数
  -1: 表示已经糊了
   0: 表示听牌,或打一张后听牌 (龙七对=0)
*/
func CalcShanTenTile(tile []int) (retNum, retMode int) {
	all := util.CheckTile(tile)
	if all < 0 {
		retMode = all
		return
	}

	m := all / 3
	if m > 4 {
		m = 4 // 面子数最大为4
	}

	retNum, retMode = 1024, 0
	sht := calcNormal(tile, m)
	if sht < retNum {
		retNum, retMode = sht, 1
	}

	if m == 4 { // m=4 时才能凑够7对
		sht = calcSeven(tile)
		if sht < retNum {
			retNum, retMode = sht, 2
		} else if sht == retNum {
			retMode |= 2
		}
	}
	return
}

/**
 * 数牌
 *
 * @param lhs
 * @param rhs
 * @param m 面子数
 */
func addS(lhs, rhs []int, m int) {
	var j, k, sht int
	for j = m + 5; j >= 5; j-- {
		sht = tools.Min(lhs[j]+rhs[0], lhs[0]+rhs[j])

		for k = 5; k < j; k++ {
			sht = tools.Min(sht, lhs[k]+rhs[j-k], lhs[j-k]+rhs[k])
		}
		lhs[j] = sht
	}
	for j = m; j >= 0; j-- {
		sht = lhs[j] + rhs[0]

		for k = 0; k < j; k++ {
			sht = tools.Min(sht, lhs[k]+rhs[j-k])
		}
		lhs[j] = sht
	}
}

/**
 * 字牌
 *
 * @param lhs
 * @param rhs
 * @param m 面子数
 */
func addH(lhs, rhs []int, m int) {
	j := m + 5
	sht := tools.Min(lhs[j]+rhs[0], lhs[0]+rhs[j])
	for k := 5; k < j; k++ {
		sht = tools.Min(sht, lhs[k]+rhs[j-k], lhs[j-k]+rhs[k])
	}
	lhs[j] = sht
}

/**
 * 累积一个牌类型的值
 *
 * @param tile 值
 * @param from 下标从(花色的第二枚)
 * @param to 下标到,不包含自身
 * @param base 基础值(花色的第一枚)
 */
func accum(tile []int, from, to, base int) int {
	for i := from; i < to; i++ {
		base = 5*base + tile[i]
	}
	return base
}

/**
 * 计算通常向听
 *
 * @param tile 34编码手牌值
 * @param m 面子数
 */
func calcNormal(tile []int, m int) int {
	// 复制到ret,不能修改[st.mps,st.mph]原本数据,支持并发计算
	ret := append([]int(nil), mps[accum(tile, 1, 9, tile[0])]...)
	addS(ret, mps[accum(tile, 10, 18, tile[9])], m)
	addS(ret, mps[accum(tile, 19, 27, tile[18])], m)
	addH(ret, mph[accum(tile, 28, 34, tile[27])], m)
	return ret[5+m] - 1 // 普通牌面计算向听数
}

/**
 * 计算七对向听
 * http://ara.moo.jp/mjhmr/shanten.htm
 * 向听数 = 6-对子数+max(0,7-种类数)
 *
 * @param tile 34编码手牌值
 * @param m 面子数
 */
func calcSeven(tile []int) int {
	var pair, kind int
	for i := 0; i < util.TileMax; i++ {
		if tile[i] > 0 {
			kind++ // 牌种类数
			if tile[i] >= 2 {
				pair++ // 对子数
			}
		}
	}
	// 七对最大向听为6向听
	// 13枚牌,换6张必然听牌
	pair = 6 - pair
	if kind < 7 {
		// 七对至少7种牌才能听牌
		// 手上4个对子,原本只需要摸2枚其他类型的,即可听牌
		// 若是暗刻杠重复的话,则需要先补全其他类型的部分,这个操作即增加了向听数
		return pair + 7 - kind
	}
	return pair
}

//goland:noinspection GoUnusedFunction
func calcKoKuSi(tile []int) int {
	var pair, kind int
	// 国士无双,十三幺, 只满足下面这种特殊牌型
	// 1,9万 1,9筒 1,9条 东,南,西,北,白,发,中
	for _, i := range []int{0, 8, 9, 17, 18, 26, 27, 28, 29, 30, 31, 32, 33} {
		if tile[i] > 0 {
			kind++
			if tile[i] >= 2 {
				pair++
			}
		}
	}
	kind = 13 - kind
	if pair > 0 {
		return kind - 1
	}
	return kind
}

/*CalcShanTan
向听数计算方法,对于13张牌时和查表法一致
小于13张牌时不适用,还是用查表法高效准确
https://www.bilibili.com/read/cv12268871
https://github.com/skk294/mahjongg
*/
func CalcShanTan(tiles34 []int) (retNum, retMode int) {
	all := util.CheckTile(tiles34)
	if all < 0 {
		retMode = all
		return
	}

	m := all / 3
	if m > 4 {
		m = 4 // 面子数最大为4
	}

	// 该方法计算向听数和查表法比较,13张牌时基本没问题
	// 但是有些极端情况会计算为8向听,和查表的结果略微不同
	var (
		appM = []func([]int, *int){
			func(tp []int, cnt *int) {
				for j := 0; j < util.TileMax; j++ {
					if tp[j] >= 3 {
						tp[j] -= 3
						*cnt++ // 刻子数增加
					}
				}
			},
			func(tp []int, cnt *int) {
				var a, b, c int
				for a = 0; a < 3; a++ {
					for b = 0; b < 7; {
						c = 9*a + b
						if tp[c] >= 1 && tp[c+1] >= 1 && tp[c+2] >= 1 {
							tp[c]--
							tp[c+1]--
							tp[c+2]--
							*cnt++ // 顺子数增加
						} else {
							b++
						}
					}
				}
			},
		}

		appD = []func([]int, *int) bool{
			func(tp []int, cnt *int) bool {
				tc := *cnt // [2] 方案的搭子
				for j := 0; j < util.TileMax; j++ {
					if tp[j] >= 2 {
						tp[j] -= 2
						*cnt++
					}
				}
				return tc != *cnt // 有雀头
			},
			func(tp []int, cnt *int) bool {
				// [11] 方案的搭子,字牌无顺子
				var a, b, c int
				for a = 0; a < 3; a++ {
					for b = 0; b < 8; {
						c = 9*a + b
						if tp[c] >= 1 && tp[c+1] >= 1 {
							tp[c]--
							tp[c+1]--
							*cnt++
						} else {
							b++
						}
					}
				}
				return false
			},
			func(tp []int, cnt *int) bool {
				// [101] 方案的搭子,字牌无顺子
				var a, b, c int
				for a = 0; a < 3; a++ {
					for b = 0; b < 7; {
						c = 9*a + b
						if tp[c] >= 1 && tp[c+2] >= 1 {
							tp[c]--
							tp[c+2]--
							*cnt++
						} else {
							b++
						}
					}
				}
				return false
			},
		}

		// m:  面子数(刻子+顺子)
		// d:  搭子数(面子少1张,[2,11,101]这3种类型)
		// qt: true(有雀头)
		calcXT = func(m, d int, qt bool) int {
			var c, q int
			if m+d > 5 {
				c = m + d - 5
			}
			if m+d <= 4 || qt {
				q = 1
			}
			return 9 - 2*m - d + c - q
		}

		qt  bool
		tpm = make([]int, len(tiles34))
		tpd = make([]int, len(tiles34))

		mCnt, dCnt int
	)
	retNum = 1024 // 保留最小向听数

	// 顺子刻子先后顺序取出
	for _, am := range [][]int{{0, 1}, {1, 0}} {
		copy(tpm, tiles34)

		mCnt = 0 // 面子数
		for _, anv := range am {
			appM[anv](tpm, &mCnt)
		}

		// 搭子数,[2],[11],[101]这3种搭子方案先后顺序取出
		for _, ad := range [][]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2},
			{1, 2, 0}, {2, 0, 1}, {2, 1, 0}} {
			copy(tpd, tpm)

			dCnt, qt = 0, false // 搭子数,雀头
			for _, adv := range ad {
				if appD[adv](tpd, &dCnt) {
					qt = true // [2]方案有雀头
				}
			}

			if dCnt = calcXT(mCnt, dCnt, qt); dCnt < retNum {
				retNum = dCnt // 最小向听数
			}
		}
	}
	if retNum < 1024 {
		retMode = 1
	}

	if m == 4 {
		if dCnt = calcSeven(tiles34); dCnt < retNum {
			retNum, retMode = dCnt, 2 // 加上7对最小向听
		} else if dCnt == retNum {
			retMode |= 2
		}
	}
	return
}
