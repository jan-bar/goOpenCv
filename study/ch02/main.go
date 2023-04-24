package main

import (
	"fmt"
	"log"

	"gocv.io/x/gocv"
)

var window = gocv.NewWindow("janbar")

func main() {
	//goland:noinspection GoUnhandledErrorResult
	defer window.Close()

	log.Println(gocv.Version(), gocv.OpenCVVersion())

	err := cv1()
	if err != nil {
		log.Fatal(err)
	}
}

func showImg(p, str string, flag gocv.IMReadFlag) error {
	img := gocv.IMRead(p, flag)
	//goland:noinspection GoUnhandledErrorResult
	defer img.Close()
	if img.Empty() {
		return fmt.Errorf("%s empty", p)
	}
	fmt.Printf("行/高/Y:%d,列/宽/X:%d,通道:%d,类型:%s,读方式:%s\n",
		img.Rows(), img.Cols(), img.Channels(), img.Type(), str)
	window.IMShow(img)
	window.WaitKey(0)
	return nil
}

var IMReadFlagArr = []struct {
	g gocv.IMReadFlag
	s string
}{
	{g: gocv.IMReadUnchanged, // 按图像原样读出,RGBA这种,包含透明度(没有A则结果同 IMReadColor)
		s: "IMReadUnchanged"},
	{g: gocv.IMReadGrayScale, // 1通道灰度图,源码计算用:Gray = R*0.299 + G*0.587 + B*0.114
		s: "IMReadGrayScale"},
	{g: gocv.IMReadColor, // 转换为BGR这种3通道
		s: "IMReadColor"},
	{g: gocv.IMReadAnyDepth, s: "IMReadAnyDepth"},
	{g: gocv.IMReadAnyColor, s: "IMReadAnyColor"},
	{g: gocv.IMReadLoadGDAL, s: "IMReadLoadGDAL"},
	// 下面这几种方案会缩小图片尺寸
	{g: gocv.IMReadReducedGrayscale2, s: "IMReadReducedGrayscale2"},
	{g: gocv.IMReadReducedColor2, s: "IMReadReducedColor2"},
	{g: gocv.IMReadReducedGrayscale4, s: "IMReadReducedGrayscale4"},
	{g: gocv.IMReadReducedColor4, s: "IMReadReducedColor4"},
	{g: gocv.IMReadReducedGrayscale8, s: "IMReadReducedGrayscale8"},
	{g: gocv.IMReadReducedColor8, s: "IMReadReducedColor8"},

	{g: gocv.IMReadIgnoreOrientation, s: "IMReadIgnoreOrientation"},
}

func cv1() error {
	p := "../data/messi5.jpg"
	for _, v := range IMReadFlagArr {
		err := showImg(p, v.s, v.g)
		if err != nil {
			return err
		}
	}
	return nil
}
