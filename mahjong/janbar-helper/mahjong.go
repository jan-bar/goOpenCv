package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"

	"github.com/jan-bar/goOpenCv/mahjong/janbar-helper/analyze"
	"github.com/jan-bar/goOpenCv/mahjong/janbar-helper/util"
	"github.com/jan-bar/tools/screenshot"
	"gocv.io/x/gocv"
)

const (
	scoreDefault = 0.9

	rectTmpTxt = "rect.tmp.txt"
)

var (
	mjH, mjW int // 每个麻将的高宽
	// 27个麻将牌的mat数据
	mjP = make([]gocv.Mat, 0, 27)

	// 匹配结果是32位浮点数,因此要用float32进行比较
	matchScore float32 = scoreDefault
)

func main() {
	pic := flag.String("p", "", "27张麻将牌图片路径")
	mac := flag.Bool("m", false, "全屏截图匹配,生成区域坐标")
	sco := flag.Float64("s", scoreDefault, "匹配图片分数,[0.8,1],合理设置值")
	flag.Parse()

	if *sco >= 0.8 && *sco <= 1 {
		// 通过传参修改图像匹配度
		// 限定在某个范围,超过范围匹配也没啥意义
		matchScore = float32(*sco)
	}

	temp := gocv.IMRead(*pic, gocv.IMReadColor)
	if temp.Empty() {
		log.Fatalf("%s pic read error", *pic)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer temp.Close()

	hw := temp.Size()
	mjH, mjW = hw[0]/3, hw[1]/9 // 得到每个麻将的高和宽

	for i := 0; i < 3; i++ {
		for j := 0; j < 9; j++ {
			mjP = append(mjP, temp.Region(image.Rect(j*mjW, i*mjH, j*mjW+mjW, i*mjH+mjH)))
		}
	}

	var err error
	if *mac {
		err = analyzeOne()
	} else {
		err = mahjong()
	}
	if err != nil {
		log.Fatal(err)
	}
}

var (
	red   = color.RGBA{R: 255}
	white = color.RGBA{R: 255, G: 255, B: 255}
)

func fillPoly(src *gocv.Mat, r, c, h, w int) {
	// 擦除匹配区域麻将,避免后续重复匹配
	pts := [][]image.Point{{
		image.Pt(c+5, r+5),
		image.Pt(c+5, r+h-5),
		image.Pt(c+w-5, r+h-5),
		image.Pt(c+w-5, r+5),
	}}
	// 该方法会将上面的点围成的多边形内部填充指定颜色
	pv := gocv.NewPointsVectorFromPoints(pts)
	gocv.FillPoly(src, pv, white)
	pv.Close()
}

func mahjong() error {
	var (
		ok   bool
		tile = make([]int, util.TileMax)
		last = make([]int, util.TileMax)
		rect image.Rectangle

		window = gocv.NewWindow("janbar")
	)
	//goland:noinspection GoUnhandledErrorResult
	defer window.Close()

	b, err := os.ReadFile(rectTmpTxt)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &rect)
	if err != nil {
		return err
	}

	window.ResizeWindow(rect.Dx(), rect.Dy()) // 窗口和正常游戏相同大小

	for {
		err = matchOne(rect, tile, window)
		if err != nil {
			return err
		}

		ok = false
		for i, v := range last {
			if tile[i] != v {
				ok = true
				break
			}
		}

		if ok { // 手牌发生变化时重新分析
			analyze.Analyze(tile)
			copy(last, tile)
		}
	}
}

func matchOne(rect image.Rectangle, tile []int, window *gocv.Window) error {
	img, err := screenshot.CaptureRect(rect)
	if err != nil {
		return err
	}

	src, err := gocv.ImageToMatRGB(img)
	if err != nil {
		return err
	}
	saveImg := src.Clone()
	mask := gocv.NewMat()
	defer func() {
		_ = src.Close()
		_ = saveImg.Close()
		_ = mask.Close()
	}()

	for i := range tile {
		tile[i] = 0
	}

	for im, mat := range mjP {
		ims := util.NumToTileStr(im)

		result := gocv.NewMat()
		// 下面调用图片匹配,且只保留一定阈值的结果数据
		gocv.MatchTemplate(src, mat, &result, gocv.TmCcoeffNormed, mask)

		for c, rm := 0, result.Rows(); c < result.Cols(); c++ {
			for r := 0; r < rm; r++ {
				score := result.GetFloatAt(r, c)
				if score > matchScore {
					gocv.PutText(&saveImg, ims, image.Pt(c, r+22), gocv.FontHersheyPlain, 1.5, red, 1)
					tile[im]++

					// 擦除当前匹配麻将,避免后续相似牌的错误匹配
					fillPoly(&src, r, c, mjH, mjW)

					c += mjW - 1
					break // 匹配后跳到下个位置,避免重复统计当前麻将牌
				}
			}
		}
		_ = result.Close()
	}

	window.IMShow(saveImg) // 显示图片,按Esc键退出程序
	if window.WaitKey(1000) == 27 {
		return fmt.Errorf("exit")
	}
	return nil
}

func analyzeOne() error {
	var (
		rect image.Rectangle
		pos  *[]int
	)
	b, err := os.ReadFile(rectTmpTxt)
	if err == nil {
		err = json.Unmarshal(b, &rect) // 已有小范围区域坐标
	}
	if err != nil {
		rect, err = screenshot.ScreenRect()
		if err != nil {
			return err
		}

		// 没有小范围区域坐标,直接全屏截图,然后计算
		if cnt := screenshot.GetSystemMetrics(screenshot.SM_CMONITORS); cnt > 1 {
			rect.Max.X *= cnt // 按照显示器个数扩展宽度
		}
		pos = &[]int{0, 0, math.MaxInt}
	}

	img, err := screenshot.CaptureRect(rect)
	if err != nil {
		return err
	}

	src, err := gocv.ImageToMatRGB(img)
	if err != nil {
		return err
	}

	saveImg := src.Clone()
	mask := gocv.NewMat()
	defer func() {
		_ = src.Close()
		_ = mask.Close()
		_ = saveImg.Close()
	}()

	tile := make([]int, 34)
	for im, mat := range mjP {
		ims := util.NumToTileStr(im) // 低于0.9的匹配可以丢弃了

		result := gocv.NewMat()
		// 下面调用图片匹配,且只保留一定阈值的结果数据
		gocv.MatchTemplate(src, mat, &result, gocv.TmCcoeffNormed, mask)

		for c, rm := 0, result.Rows(); c < result.Cols(); c++ {
			for r := 0; r < rm; r++ {
				score := result.GetFloatAt(r, c)
				if score > matchScore {
					if pos != nil {
						if (*pos)[0] < r {
							(*pos)[0] = r // 最大Y坐标
						}
						if (*pos)[1] < c {
							(*pos)[1] = c // 最大X坐标
						}
						if (*pos)[2] > c {
							(*pos)[2] = c // 最小X坐标
						}
					}
					gocv.PutText(&saveImg, ims, image.Pt(c, r+22), gocv.FontHersheyPlain, 1.5, red, 1)
					fillPoly(&src, r, c, mjH, mjW)
					tile[im]++

					c += mjW - 1
					break // 匹配后跳到下个位置,避免重复统计
				}
			}
		}
		_ = result.Close()
	}
	gocv.IMWrite("src.tmp.png", src)
	gocv.IMWrite("out.tmp.png", saveImg)

	analyze.Analyze(tile)

	const minPng = "min.tmp.png"
	if pos == nil { // 已有坐标
		_ = os.Remove(minPng)
		return nil
	}

	rect.Min.X = (*pos)[2] - mjW
	if rect.Min.X < 0 {
		rect.Min.X = 0 // 遇到屏幕最左边
	}
	rect.Max.X = (*pos)[1] + 2*mjW
	if xm := (*pos)[2] + mjW*15; rect.Max.X < xm {
		rect.Max.X = xm // 识别的牌太少,凑够15张牌的宽度
	}

	rect.Min.Y = (*pos)[0] - 10 // Y方向一般不会越界
	rect.Max.Y = (*pos)[0] + mjH + 10
	// 保存区域截图的坐标到文件
	b, err = json.Marshal(rect)
	if err != nil {
		return err
	}
	// 这个分析主要是得到小范围区域截图坐标
	err = os.WriteFile(rectTmpTxt, b, os.ModePerm)
	if err != nil {
		return err
	}

	gocv.IMWrite(minPng, saveImg.Region(rect))
	return nil
}
