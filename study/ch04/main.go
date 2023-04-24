package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"time"

	"gocv.io/x/gocv"
)

var (
	window = gocv.NewWindow("janbar")
)

func main() {
	//goland:noinspection GoUnhandledErrorResult
	defer window.Close()

	err := draw()
	if err != nil {
		log.Fatal(err)
	}

	err = drawOpencv()
	if err != nil {
		log.Fatal(err)
	}
}

//goland:noinspection GoUnhandledErrorResult
func draw() error {
	img := gocv.NewMatWithSizesWithScalar([]int{100, 100},
		gocv.MatTypeCV8UC4, gocv.NewScalar(0, 0, 0, 0))
	defer img.Close()

	red := color.RGBA{R: 255}

	// 在两个点之间画直线
	gocv.Line(&img, image.Pt(10, 10), image.Pt(30, 30), red, 1)
	// 画直线,终点位置带箭头
	gocv.ArrowedLine(&img, image.Pt(10, 80), image.Pt(30, 40), red, 1)

	pvs := gocv.NewPointsVectorFromPoints([][]image.Point{{
		image.Pt(40, 40),
		image.Pt(40, 50),
		image.Pt(50, 50),
		image.Pt(50, 40),
	}})
	// 按照上面点画直线,可以选择isClosed=true时,闭合区域
	gocv.Polylines(&img, pvs, true, red, 1)
	pvs.Close()

	// 画矩形区域,thickness = -1时,区域内部全都填充
	gocv.Rectangle(&img, image.Rect(50, 60, 60, 70), red, -1)

	// 画圆形,给定圆心和半径
	gocv.Circle(&img, image.Pt(80, 80), 10, red, 1)

	// 画椭圆
	gocv.Ellipse(&img, image.Pt(80, 20), image.Pt(20, 5),
		90, 0, 360, red, 1)

	// 输出文字
	gocv.PutText(&img, time.Now().Format(time.DateTime), image.Pt(3, 90),
		gocv.FontHersheyPlain, 1, red, 1)

	window.IMShow(img)
	window.WaitKey(0)

	td := img.Clone()
	defer td.Close()
	gocv.Threshold(img, &td, 0.9, 1, gocv.ThresholdToZero)
	window.IMShow(td)
	window.WaitKey(0)

	dst := img.Clone()
	defer dst.Close()
	rot := gocv.GetRotationMatrix2D(image.Pt(10, 10), 1.0, 1.0)
	defer rot.Close()
	// 仿射变换
	gocv.WarpAffine(img, &dst, rot, image.Pt(img.Cols(), img.Rows()))

	window.IMShow(dst)
	window.WaitKey(0)

	ms := gocv.IMRead("../data/messi5.jpg", gocv.IMReadGrayScale)
	defer ms.Close()
	cd := gocv.NewMat()
	defer cd.Close()
	gocv.Canny(ms, &cd, 100, 100) // 边缘检测

	window.IMShow(cd)
	window.WaitKey(0)

	// 从一个 分辨率大尺寸的图像向上构建一个金子塔,尺寸变小 分辨率降低
	// gocv.PyrDown()
	// 从一个低分 率小尺寸的图像向下构建一个 子塔 尺 寸变大 但分 率不会增加
	// gocv.PyrUp()
	return nil
}

//goland:noinspection GoUnhandledErrorResult
func drawOpencv() error {
	r1, r2, angf, d := 70, 30, float64(60), 170
	h := int(float64(d) / 2.0 * math.Sqrt(3))

	dot_red := image.Pt(256, 128)
	dot_green := image.Pt(dot_red.X-d/2, dot_red.Y+h)
	dot_blue := image.Pt(dot_red.X+d/2, dot_red.Y+h)

	red := color.RGBA{R: 255}
	green := color.RGBA{G: 255}
	blue := color.RGBA{B: 255}
	black := color.RGBA{}

	full := -1

	// 创建指定大小的黑色背景
	img := gocv.NewMatWithSizeFromScalar(gocv.Scalar{}, 512, 512, gocv.MatTypeCV8UC3)
	defer img.Close()

	// 画3个颜色的实心圆形
	gocv.Circle(&img, dot_red, r1, red, full)
	gocv.Circle(&img, dot_green, r1, green, full)
	gocv.Circle(&img, dot_blue, r1, blue, full)
	// 中心填充黑色实心圆形
	gocv.Circle(&img, dot_red, r2, black, full)
	gocv.Circle(&img, dot_green, r2, black, full)
	gocv.Circle(&img, dot_blue, r2, black, full)
	// 3个缺口处,画60度的部分圆形,填充黑色
	gocv.Ellipse(&img, dot_red, image.Pt(r1, r1), angf, 0, angf, black, full)
	gocv.Ellipse(&img, dot_green, image.Pt(r1, r1), 360-angf, 0, angf, black, full)
	gocv.Ellipse(&img, dot_blue, image.Pt(r1, r1), 360-2*angf, angf, 0, black, full)
	// 显示相关文字
	gocv.PutText(&img, "OpenCV", image.Pt(15, 450),
		gocv.FontHersheySimplex, 4, color.RGBA{R: 255, G: 255, B: 255}, 10)

	// todo gocv库没实现 utf8 字符的绘制
	//   检查已实现列表 https://github.com/hybridgroup/gocv/blob/release/ROADMAP.md\
	window.IMShow(img)
	window.WaitKey(0)
	return nil
}
