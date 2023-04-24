package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"gocv.io/x/gocv"
)

var (
	window = gocv.NewWindow("janbar")

	red = color.RGBA{R: 255}
)

func main() {
	//goland:noinspection GoUnhandledErrorResult
	defer window.Close()

	err := matchMax()
	if err != nil {
		log.Fatal(err)
	}

	err = matchMulti()
	if err != nil {
		log.Fatal(err)
	}
}

func matchMax() error {
	p := "../data/messi5.jpg"
	img := gocv.IMRead(p, gocv.IMReadUnchanged)
	if img.Empty() {
		return fmt.Errorf("%s is empty", p)
	}
	defer img.Close()

	p = "../data/messi_face.jpg"
	tpl := gocv.IMRead(p, gocv.IMReadGrayScale)
	if tpl.Empty() {
		return fmt.Errorf("%s is empty", p)
	}
	defer tpl.Close()
	w, h := tpl.Cols(), tpl.Rows()

	methods := []gocv.TemplateMatchMode{
		gocv.TmSqdiff,       // 平方差匹配算法,它计算每个像素点的差值,然后求平方和,匹配的结果是差值最小的位置
		gocv.TmSqdiffNormed, // 标准平方差匹配算法,它计算每个像素点的差值,并将其归一化,然后求平方和,匹配的结果是归一化差值最小的位置
		gocv.TmCcorr,        // 相关性匹配算法,它计算模板图像和待匹配图像的相关性,匹配的结果是相关性最大的位置
		gocv.TmCcorrNormed,  // 标准相关性匹配算法,它计算模板图像和待匹配图像的相关性,并将其归一化,匹配的结果是归一化相关性最大的位置
		gocv.TmCcoeff,       // 相关系数匹配算法,它计算模板图像和待匹配图像的相关系数,匹配的结果是相关系数最大的位置
		gocv.TmCcoeffNormed, // 标准相关系数匹配算法,它计算模板图像和待匹配图像的相关系数,并将其归一化,匹配的结果是归一化相关系数最大的位置
	}

	mask := gocv.NewMat()
	defer mask.Close()

	for _, method := range methods {
		rgb := img.Clone()
		img2 := gocv.NewMat() // 转换灰度图进行匹配
		gocv.CvtColor(rgb, &img2, gocv.ColorBGRToGray)

		res := gocv.NewMat() // 匹配结果也只32位浮点型单通道数据
		gocv.MatchTemplate(img2, tpl, &res, method, mask)

		minVal, maxVal, minLoc, maxLoc := gocv.MinMaxLoc(res)
		log.Println(method, minVal, maxVal, minLoc, maxLoc, res.Type(), res.Channels())

		topLeft := maxLoc
		if method == gocv.TmSqdiff || method == gocv.TmSqdiffNormed {
			topLeft = minLoc // 取最小值位置
		}

		// 在彩色图像上圈出匹配人脸,弹窗显示结果
		gocv.Rectangle(&rgb, image.Rect(topLeft.X, topLeft.Y, topLeft.X+w, topLeft.Y+h), red, 2)
		window.IMShow(rgb)
		window.WaitKey(0)

		_ = res.Close()
		_ = img2.Close()
		_ = rgb.Close()
	}
	return nil
}

func matchMulti() error {
	p := "../data/mario.png"
	rgb := gocv.IMRead(p, gocv.IMReadUnchanged)
	if rgb.Empty() {
		return fmt.Errorf("%s is empty", p)
	}
	defer rgb.Close()

	mario := gocv.NewMat()
	defer mario.Close()
	gocv.CvtColor(rgb, &mario, gocv.ColorBGRToGray)

	p = "../data/mario_coin.png"
	tpl := gocv.IMRead(p, gocv.IMReadGrayScale)
	if tpl.Empty() {
		return fmt.Errorf("%s is empty", p)
	}
	defer tpl.Close()
	w, h := tpl.Cols(), tpl.Rows()

	res := gocv.NewMat()
	defer res.Close()
	mask := gocv.NewMat()
	defer mask.Close()
	// 使用灰度图进行匹配,结果圈在彩色图里面
	// res 结果是 CV32F 这种float32的浮点数矩阵,每个点都是相似度
	gocv.MatchTemplate(mario, tpl, &res, gocv.TmCcoeffNormed, mask)

	// 合适的阈值,有利于减少范围
	FindResult(res, 0.8, w, h, func(rect image.Rectangle) {
		gocv.Rectangle(&rgb, rect, red, 1)
	})

	window.IMShow(rgb)
	window.WaitKey(0)
	return nil
}

func FindResult(img gocv.Mat, threshold float32, w, h int, f func(image.Rectangle)) {
	var (
		pos = map[[2]int]struct{}{}
		ok  bool
	)
	for r, c, rm, cm := 0, 0, img.Rows(), img.Cols(); r < rm; r++ {
		for c = 0; c < cm; c++ {
			if s := img.GetFloatAt(r, c); s >= threshold {
				xy := [2]int{c / w, r / h}
				if _, ok = pos[xy]; ok {
					continue // 不要重复计算该区域
				}
				pos[xy] = struct{}{}

				f(image.Rect(c, r, c+w, r+h))
				c += w - 1 // 下个区域,减少计算
			}
		}
	}
}
