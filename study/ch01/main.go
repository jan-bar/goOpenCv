package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gocv.io/x/gocv"
)

/*
有关图像通道的知识
tips1:
一个图像的通道数是N，就表明每个像素点处有N个数，一个a×b的N通道图像，其图像矩阵实际上是b行N×a列的数字矩阵。
OpenCV中图像的通道可以是1、2、3和4。其中常见的是1通道和3通道，2通道和4通道不常见。
	1通道的是灰度图。
	3通道的是彩色图像，比如RGB图像。
	4通道的图像是RGBA，是RGB加上一个A通道，也叫alpha通道，表示透明度。PNG图像是一种典型的4通道图像。alpha通道可以赋值0到1，或者0到255，表示透明到不透明。
	2通道的图像是RGB555和RGB565。2通道图在程序处理中会用到，如傅里叶变换，可能会用到，一个通道为实数，一个通道为虚数，主要是编程方便。
		RGB555是16位的，2个字节，5+6+5，第一字节的前5位是R，后三位+第二字节是G，第二字节后5位是B，可见对原图像进行压缩了

tips2:
OpenCV中用imshow( )来显示图像，只要Mat的数据矩阵符合图像的要求，就可以用imshow来显示。
	二通道好像不可以。超过了4通道，就不是图像了，imshow( )也显示不了。

tips3:
imshow( )显示单通道图像时一定是灰度图，如果我们想显示红色的R分量，还是应该按三通道图像显示，只不过G和B通道要赋值成0或255.

tips4:
通道分解用split( )，通道合成用merg( )，这俩函数都是mixchannel( )的特例。

RGB转GRAY是根据一个心理学公式来的：Gray = R*0.299 + G*0.587 + B*0.114
RGB转GRBA，默认A通道的数值是255，也就是不透明的。

注意rgb图像的通道排列是BGR
*/

var window *gocv.Window

func main() {
	window = gocv.NewWindow("janbar")
	defer window.Close()

	err := test()
	if err != nil {
		panic(err)
	}

	err = channels()
	if err != nil {
		panic(err)
	}
}

//goland:noinspection GoUnhandledErrorResult
func test() error {
	png := "../data/Lenna.png"
	img := gocv.IMRead(png, gocv.IMReadUnchanged)
	if img.Empty() {
		return fmt.Errorf("%s is empty", png)
	}
	defer img.Close()

	// 对应Python里面的数据,Rows <-> 高 <-> Y, Cols <-> 宽 <-> X
	// shape := []int{img.Rows(), img.Cols(), img.Channels()}
	fmt.Printf("img.shape: %d,%d,%d,%v\n",
		img.Rows(), img.Cols(), img.Channels(), img.Type())

	png = "../data/opencv_logo.png"
	logo := gocv.IMRead(png, gocv.IMReadUnchanged)
	if logo.Empty() {
		return fmt.Errorf("%s is empty", png)
	}
	defer logo.Close()
	gocv.Resize(logo, &logo, image.Point{X: 50, Y: 56}, 0, 0, 0)

	fmt.Printf("logo.shape: %d,%d,%d,%v,%v\n",
		logo.Rows(), logo.Cols(), logo.Channels(),
		logo.Size(), // Size返回(Y/h,X/w)
		logo.Type())

	png = "../data/butterfly.jpg"
	butterfly := gocv.IMRead(png, gocv.IMReadUnchanged)
	if butterfly.Empty() {
		return fmt.Errorf("%s is empty", png)
	}
	defer butterfly.Close()
	gocv.Resize(butterfly, &butterfly, image.Point{X: 50, Y: 50}, 0, 0, 0)

	fmt.Printf("butterfly.shape: %d,%d,%d,%v\n",
		butterfly.Rows(), butterfly.Cols(), butterfly.Channels(), butterfly.Type())

	// Cols <-> X <-> width, Rows <-> Y <-> height
	window.ResizeWindow(img.Cols(), img.Rows())
	window.MoveWindow(0, 0)
	window.IMShow(img) // 调整好窗口,然后显示图片
	window.WaitKey(0)

	y, x := 100, 50
	bgr := img.GetVecbAt(y, x) // 获取指定位置的(b,g,r)值
	fmt.Printf("bgr: %v\n", bgr)

	t := img.Region(image.Rect(100, 150, 100+butterfly.Cols(), 150+butterfly.Rows()))
	butterfly.CopyTo(&t) // t的底层共用img,因此CopyTo直接修改img内容

	dst := gocv.NewMat() // 由于logo是4通道,img是3通道,需要先做转换
	defer dst.Close()
	gocv.CvtColor(logo, &dst, gocv.ColorBGRAToBGR)
	t = img.Region(image.Rect(100, 300, 100+dst.Cols(), 300+dst.Rows()))
	dst.CopyTo(&t)

	// 显示当前时间字符串
	gocv.PutText(&img, time.Now().Format(time.DateTime), image.Pt(10, 100),
		gocv.FontHersheySimplex, 1, color.RGBA{G: 255}, 2)
	window.SetWindowTitle("img+logo")
	window.IMShow(img)
	window.WaitKey(0)
	return nil
}

//goland:noinspection GoUnhandledErrorResult
func channels() error {
	// 研究opencv通道的相关文章
	// https://blog.csdn.net/GDFSG/article/details/50927257

	// 将通道拆开
	rgb := gocv.NewMatWithSizesWithScalar([]int{3, 4}, gocv.MatTypeCV8UC3, gocv.NewScalar(1, 2, 3, 4))
	defer rgb.Close()

	d, err := rgb.DataPtrUint8()
	if err != nil {
		return err
	}
	fmt.Println("RGB", d)

	chs := gocv.Split(rgb) // 将每个通道的数据拆分出来,注意BGR排列
	tmp := map[int]string{0: "B", 1: "G", 2: "R"}
	for i, ch := range chs {
		d, err = ch.DataPtrUint8()
		if err != nil {
			return err
		}
		fmt.Println(tmp[i], d)
	}
	fmt.Println("-----------------------------------")

	// 合并通道
	r := gocv.NewMatWithSizesWithScalar([]int{3, 4}, gocv.MatTypeCV8UC1, gocv.Scalar{Val1: 3})
	defer r.Close()
	g := gocv.NewMatWithSizesWithScalar([]int{3, 4}, gocv.MatTypeCV8UC1, gocv.Scalar{Val1: 2})
	defer g.Close()
	b := gocv.NewMatWithSizesWithScalar([]int{3, 4}, gocv.MatTypeCV8UC1, gocv.Scalar{Val1: 1})
	defer b.Close()

	rgb = gocv.NewMat()
	defer rgb.Close()
	gocv.Merge([]gocv.Mat{b, g, r}, &rgb)

	for i, ch := range []gocv.Mat{b, g, r} {
		d, err = ch.DataPtrUint8()
		if err != nil {
			return err
		}
		fmt.Println(tmp[i], d)
	}

	d, err = rgb.DataPtrUint8()
	if err != nil {
		return err
	}
	fmt.Println("RGB", d)

	// 混合通道
	rgb = gocv.NewMatWithSizesWithScalar([]int{3, 4}, gocv.MatTypeCV8UC3, gocv.NewScalar(1, 2, 3, 4))
	defer rgb.Close()
	a := gocv.NewMatWithSizesWithScalar([]int{3, 4}, gocv.MatTypeCV8UC1, gocv.Scalar{Val1: 6})
	defer a.Close()

	rgba := gocv.NewMatWithSizesWithScalar([]int{3, 4}, gocv.MatTypeCV8UC4, gocv.Scalar{})
	defer rgba.Close()

	// rgb[0] -> rgba[0], rgb[1] -> rgba[1]
	// rgb[2] -> rgba[2], a[0] -> rgba[4]
	fromTo := []int{0, 0, 1, 1, 2, 2, 3, 3}
	gocv.MixChannels([]gocv.Mat{rgb, a}, []gocv.Mat{rgba}, fromTo)

	d, err = rgba.DataPtrUint8()
	if err != nil {
		return err
	}
	fmt.Println("RGB+A = RGBA", d)

	r = gocv.NewMatWithSizesWithScalar([]int{3, 4}, gocv.MatTypeCV8UC1, gocv.Scalar{})
	defer r.Close()
	gb := gocv.NewMatWithSizesWithScalar([]int{3, 4}, gocv.MatTypeCV8UC2, gocv.Scalar{})
	defer gb.Close()

	// rgb[0] -> gb[1], rgb[1] -> gb[0], rgb[2] -> r[0]
	fromTo = []int{0, 2, 1, 1, 2, 0}
	gocv.MixChannels([]gocv.Mat{rgb}, []gocv.Mat{r, gb}, fromTo)

	d, err = r.DataPtrUint8()
	if err != nil {
		return err
	}
	fmt.Println("RGB -> R", d)
	d, err = gb.DataPtrUint8()
	if err != nil {
		return err
	}
	fmt.Println("RGB -> GB", d)
	return nil
}
