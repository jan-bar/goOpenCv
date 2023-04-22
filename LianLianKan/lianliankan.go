package main

import (
	"errors"
	"fmt"
	"image"
	"os"
	"time"
	"unsafe"

	"github.com/jan-bar/tools/screenshot"
	"github.com/lxn/win"
	"gocv.io/x/gocv"
	"golang.org/x/sys/windows"
)

func main() {
	l, err := NewLianLianKan()
	if err != nil {
		panic(err)
	}
	defer l.Release()

	for {
		if err = l.CaptureAndLoadPic(); err != nil {
			panic(err)
		}
		if err = l.LoadData(); err != nil {
			panic(err)
		}
		l.Print()

		l.ClickAllTwoPos()
		if l.IsWin() {
			break
		}
		// (全都不能连线了)需要等待5s让游戏重排完成
		time.Sleep(time.Second * 5)
	}
}

const (
	NeedHandle = 0  // 初始需要处理状态
	HandleOver = -1 // 已经处理状态
)

type (
	LianLianKan struct {
		X, Y       int32   // 游戏左上角坐标,X向右横坐标,Y向下纵坐标
		Cx, Cy     int     // 每个格子长Cx,宽Cy
		xMax, yMax int     // xMax列,yMax行
		data       [][]int // 保存每个点数据
		picCnt     int     // 相同图片编号

		image, temp gocv.Mat
	}

	LianLianKanPos struct {
		x, y int
	}
)

func NewLianLianKan() (*LianLianKan, error) {
	window := win.FindWindow(win.StringToBSTR("ShockwaveFlash"), nil)
	if window == 0 {
		return nil, errors.New("not run flashPlayer.exe")
	}

	win.SetWindowPos(window, win.HWND_TOPMOST, 0, 0, // 置顶窗口
		0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)

	// 相比于 win.GetWindowRect,win.GetClientRect,下面的方法更精确
	// 因为win10有毛玻璃特效等,所以最好用下面方案获取窗口坐标
	// https://learn.microsoft.com/zh-cn/windows/win32/api/dwmapi/ne-dwmapi-dwmwindowattribute
	// 根据上面注释,获取第DW MWA_EXTENDED_FRAME_BOUNDS项数据
	const ExtendedFrameBounds = 9
	var RectPos win.RECT
	err := windows.DwmGetWindowAttribute(windows.HWND(window), ExtendedFrameBounds,
		unsafe.Pointer(&RectPos), uint32(unsafe.Sizeof(RectPos)))
	if err != nil {
		return nil, err
	}

	l := &LianLianKan{
		X:    RectPos.Left + 57, // 窗口左上角X坐标偏移一定值到游戏界面左上角X坐标
		Cx:   44,                // 经过计算2个图标中间横向长度
		Y:    RectPos.Top + 110, // 窗口左上角Y坐标偏移一定值到游戏界面左上角Y坐标
		Cy:   40,                // 经过计算2个图标中间纵向宽度
		xMax: 14, yMax: 10,
	}
	l.data = make([][]int, l.yMax)
	for i := 0; i < l.yMax; i++ {
		l.data[i] = make([]int, l.xMax)
	}

	l.ClickLeft(LianLianKanPos{x: 6, y: 7})
	win.SetCursorPos(0, 0) // 点击开始游戏后把鼠标移开
	return l, nil
}

// ClickLeft 左键点击一次对应位置
func (l *LianLianKan) ClickLeft(pos LianLianKanPos) {
	win.SetCursorPos(l.X+int32(pos.x*l.Cx)+20, l.Y+int32(pos.y*l.Cy)+20)
	input := win.MOUSE_INPUT{
		Type: win.INPUT_MOUSE,
		Mi: win.MOUSEINPUT{
			DwFlags: win.MOUSEEVENTF_LEFTDOWN | win.MOUSEEVENTF_LEFTUP,
		},
	}
	time.Sleep(time.Millisecond * 100) // 避免点击太快,场面尴尬到失控
	win.SendInput(1, unsafe.Pointer(&input), int32(unsafe.Sizeof(input)))
	time.Sleep(time.Millisecond * 500) // 避免点击太快,场面尴尬到失控

	if win.GetKeyState(win.VK_SPACE) < 0 {
		os.Exit(0) // 按下空格键中断鼠标点击
	}
}

// CaptureAndLoadPic 截屏游戏数据,保存成一张图片
func (l *LianLianKan) CaptureAndLoadPic() error {
	x0, y0 := int(l.X), int(l.Y) // 根据Picker.exe抓取坐标计算出来的值
	img, err := screenshot.CaptureRect(image.Rect(x0, y0, x0+615, y0+399))
	if err != nil {
		return err
	}
	l.Release() // 释放上一次资源
	l.image, err = gocv.ImageToMatRGB(img)
	if l.image.Empty() {
		return errors.New("LoadImage fail")
	}
	// 下面均为重置数据
	for ix := 0; ix < l.xMax; ix++ {
		for iy := 0; iy < l.yMax; iy++ {
			l.data[iy][ix] = NeedHandle
		}
	}
	l.picCnt = NeedHandle + 1 // 图片编号从1开始
	return nil
}

func (l *LianLianKan) LoadData() error {
	for ix := 0; ix < l.xMax; ix++ {
		for iy := 0; iy < l.yMax; iy++ {
			if l.data[iy][ix] == NeedHandle {
				if l.GetOnePic(ix, iy) {
					l.MatchTemp() // 该位置有图片,需要匹配
					l.picCnt++    // 图片编号增加
				} else {
					l.data[iy][ix] = HandleOver // 没有图片置为已处理
				}
			}
		}
	}
	return nil
}

// GetOnePic 从图片中截取一个小动物的图像
func (l *LianLianKan) GetOnePic(ix, iy int) bool {
	// 根据坐标计算每个格子的左上角坐标和宽高
	x, y, w, h := ix*l.Cx, iy*l.Cy, l.Cx-1, l.Cy-1
	// 根据上面的值计算出对应格子的左上角和右下角坐标
	// l.temp 底层和 l.image 共用,因此只需要l.image.Close()
	l.temp = l.image.Region(image.Rect(x, y, x+w, y+h))

	// 需要调试时,可以用下面方式保存单个图片
	// if !gocv.IMWrite("one.png", l.temp) {
	// 	panic("gocv.IMWrite fail")
	// }

	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			// row 表示从上到下和宽h对应,对应y垂直方向
			// col 表示从左到右和长w对应,对应x水平方向
			// 返回结果是BGR,判断R>0时说明这张图片有小动物
			if bgr := l.temp.GetVecbAt(j, i); bgr[2] > 0 {
				return true
			}
		}
	}
	return false
}

// MatchTemp 找到name图片在游戏区域所有匹配位置
func (l *LianLianKan) MatchTemp() {
	// 找到一篇文章,得到匹配多个图片的方案,通过自己的摸索理解,用不着该方案
	// https://stackoverflow.com/questions/58763007/opencv-equivalent-of-np-where
	// 一些教程文章
	// https://kebingzao.com/2021/06/02/opencv-match-template/

	result := gocv.NewMat()
	//goland:noinspection GoUnhandledErrorResult
	defer result.Close()
	m := gocv.NewMat() // mask不会空时才会执行掩码运算

	// 下面调用图片匹配,且只保留一定阈值的结果数据
	gocv.MatchTemplate(l.image, l.temp, &result, gocv.TmCcoeffNormed, m)
	_ = m.Close()

	// result 保存了原图所有像素点的匹配度,因此直接遍历这些像素点
	// 获取这些像素点的匹配度,筛选大于指定阈值的匹配度的像素点即可
	// 下面就是匹配多个结果的方案, 用这个 gocv.MinMaxLoc() 方法取的结果
	// 就是最大最小匹配度以及这两个像素点的位置
	for r, cm := 0, result.Cols(); r < result.Rows(); r++ {
		for c := 0; c < cm; c++ {
			if t := result.GetFloatAt(r, c); t > 0.8 {
				// 从图片坐标得到data的坐标,然后设置对应位置图片编号
				// 坐标除以每个图标的长度和宽度,得到数组下标
				ix, iy := (c+10)/l.Cx, (r+10)/l.Cy
				l.data[iy][ix] = l.picCnt
			}
		}
	}
}

// IsWin 判断输赢,赢了返回true
func (l *LianLianKan) IsWin() bool {
	for iy := 0; iy < l.yMax; iy++ {
		for ix := 0; ix < l.xMax; ix++ {
			if l.data[iy][ix] >= NeedHandle {
				return false
			}
		}
	}
	return true
}

// Print 打印游戏区域每种类型动物所处位置的编号
// 相同编号为同一个动物
func (l *LianLianKan) Print() {
	for iy := 0; iy < l.yMax; iy++ {
		for ix := 0; ix < l.xMax; ix++ {
			fmt.Printf("%3d", l.data[iy][ix])
		}
		fmt.Println()
	}
	fmt.Println()
}

func (l *LianLianKan) Release() {
	_ = l.image.Close()
}

// ClickAllTwoPos 将当前所有能连线的全点了
func (l *LianLianKan) ClickAllTwoPos() {
	for {
		numList := make(map[int][]LianLianKanPos, l.picCnt)
		for iy := 0; iy < l.yMax; iy++ {
			for ix := 0; ix < l.xMax; ix++ {
				if tmp := l.data[iy][ix]; tmp > NeedHandle {
					numList[tmp] = append(numList[tmp], LianLianKanPos{
						x: ix, y: iy, // 相同图片坐标放到一起
					})
				}
			}
		} // 每次都产生新地列表,将已经消除的去掉

		isClick := true
		for _, val := range numList {
			for i1 := 0; i1 < len(val); i1++ {
				p1 := val[i1]
				if l.data[p1.y][p1.x] < NeedHandle {
					continue // 该点可能被消掉,需要判断
				}
				for i2 := i1 + 1; i2 < len(val); i2++ {
					p2 := val[i2]
					if l.data[p2.y][p2.x] < NeedHandle {
						continue // 该点可能被消掉,需要判断
					}
					if l.ConnectTwoPos(p1, p2) {
						isClick = false
						break // 这两点消掉,跳出循环循环
					}
				}
			}
		}

		if isClick {
			break // 当前没有可连线,需要重新截屏获取数据
		}
	}
}

// ConnectTwoPos 连接2个图片,如果能连接返回true
func (l *LianLianKan) ConnectTwoPos(p1, p2 LianLianKanPos) bool {
	ok := false
	switch {
	case p1.x == p2.x: // |
		ok = l.Link1(p1.x, p1.y, p2.y)
	case p1.y == p2.y: // -
		ok = l.Link2(p1.y, p1.x, p2.x)
	case p1.x < p2.x && p1.y < p2.y: // \
		ok = l.Link3(p1, p2)
	case p1.x > p2.x && p1.y > p2.y: // \
		ok = l.Link3(p2, p1)
	case p1.x < p2.x && p1.y > p2.y: // /
		ok = l.Link4(p1, p2)
	case p1.x > p2.x && p1.y < p2.y: // /
		ok = l.Link4(p2, p1)
	}
	if ok { // 点击这两点,并置为已处理
		l.ClickLeft(p1)
		l.ClickLeft(p2)
		l.data[p1.y][p1.x], l.data[p2.y][p2.x] = HandleOver, HandleOver
	}
	return ok
}

// Link1 |
func (l *LianLianKan) Link1(x, yMin, yMax int) bool {
	if x == 0 || x == l.xMax-1 {
		return true // 最左边和最右边,直接就能连
	}
	if yMin > yMax {
		yMin, yMax = yMax, yMin
	}
	if l.CanLinkY(x, yMin+1, yMax-1) {
		return true
	}
	for tx := -1; tx <= l.xMax; tx++ {
		if tx == x { // 跳过本列
			continue
		}
		tx0, tx1 := tx, x-1 // 左侧
		if tx > x {
			tx0, tx1 = x+1, tx // 右侧
		}
		if !l.CanLinkY(tx, yMin, yMax) || !l.CanLinkX(yMin, tx0, tx1) || !l.CanLinkX(yMax, tx0, tx1) {
			continue
		}
		return true // 3条线都能连接
	}
	return false
}

// Link2 -
func (l *LianLianKan) Link2(y, xMin, xMax int) bool {
	if y == 0 || y == l.yMax-1 {
		return true // 最上边和最下边,直接就能连
	}
	if xMin > xMax {
		xMin, xMax = xMax, xMin
	}
	if l.CanLinkX(y, xMin+1, xMax-1) {
		return true
	}
	for ty := -1; ty <= l.yMax; ty++ {
		if ty == y {
			continue // 跳过本行
		}
		ty0, ty1 := ty, y-1 // 丄侧
		if ty > y {
			ty0, ty1 = y+1, ty // 下侧
		}
		if !l.CanLinkX(ty, xMin, xMax) || !l.CanLinkY(xMin, ty0, ty1) || !l.CanLinkY(xMax, ty0, ty1) {
			continue
		}
		return true // 3条线都能连接
	}
	return false
}

// Link3 \
func (l *LianLianKan) Link3(xyMin, xyMax LianLianKanPos) bool {
	for ix := -1; ix <= l.xMax; ix++ { // X方向
		switch {
		case ix < xyMin.x:
			if !l.CanLinkY(ix, xyMin.y, xyMax.y) || !l.CanLinkX(xyMin.y, ix, xyMin.x-1) || !l.CanLinkX(xyMax.y, ix, xyMax.x-1) {
				continue
			}
			return true
		case ix == xyMin.x:
			if !l.CanLinkY(ix, xyMin.y+1, xyMax.y) || !l.CanLinkX(xyMax.y, ix, xyMax.x-1) {
				continue
			}
			return true
		case ix < xyMax.x:
			if !l.CanLinkY(ix, xyMin.y, xyMax.y) || !l.CanLinkX(xyMin.y, xyMin.x+1, ix) || !l.CanLinkX(xyMax.y, ix, xyMax.x-1) {
				continue
			}
			return true
		case ix == xyMax.x:
			if !l.CanLinkX(xyMin.y, xyMin.x+1, xyMax.x) || !l.CanLinkY(xyMax.x, xyMin.y, xyMax.y-1) {
				continue
			}
			return true
		default:
			if !l.CanLinkY(ix, xyMin.y, xyMax.y) || !l.CanLinkX(xyMin.y, xyMin.x+1, ix) || !l.CanLinkX(xyMax.y, xyMax.x+1, ix) {
				continue
			}
			return true
		}
	}
	for iy := -1; iy <= l.yMax; iy++ { // Y方向
		if !l.CanLinkX(iy, xyMin.x, xyMax.x) {
			continue // 这条线连不上,下面也不用判断了
		}
		switch {
		case iy == xyMin.y || iy == xyMax.y: // 如果能成功上面X方向就返回了
		case iy < xyMin.y:
			if !l.CanLinkY(xyMin.x, iy, xyMin.y-1) || !l.CanLinkY(xyMax.x, iy, xyMax.y-1) {
				continue
			}
			return true
		case iy < xyMax.y:
			if !l.CanLinkY(xyMin.x, xyMin.y+1, iy) || !l.CanLinkY(xyMax.x, iy, xyMax.y-1) {
				continue
			}
			return true
		default:
			if !l.CanLinkY(xyMin.x, xyMin.y+1, iy) || !l.CanLinkY(xyMax.x, xyMax.y+1, iy) {
				continue
			}
			return true
		}
	}
	return false
}

// Link4 /
func (l *LianLianKan) Link4(xMinYMax, xMaxYMin LianLianKanPos) bool {
	for ix := -1; ix <= l.xMax; ix++ {
		switch {
		case ix < xMinYMax.x:
			if !l.CanLinkY(ix, xMaxYMin.y, xMinYMax.y) || !l.CanLinkX(xMinYMax.y, ix, xMinYMax.x-1) || !l.CanLinkX(xMaxYMin.y, ix, xMaxYMin.x-1) {
				continue
			}
			return true
		case ix == xMinYMax.x:
			if !l.CanLinkY(ix, xMaxYMin.y, xMinYMax.y-1) || !l.CanLinkX(xMaxYMin.y, ix, xMaxYMin.x-1) {
				continue
			}
			return true
		case ix < xMaxYMin.x:
			if !l.CanLinkY(ix, xMaxYMin.y, xMinYMax.y) || !l.CanLinkX(xMaxYMin.y, ix, xMaxYMin.x-1) || !l.CanLinkX(xMinYMax.y, xMinYMax.x+1, ix) {
				continue
			}
			return true
		case ix == xMaxYMin.x:
			if !l.CanLinkY(ix, xMaxYMin.y+1, xMinYMax.y) || !l.CanLinkX(xMinYMax.y, xMinYMax.x+1, ix) {
				continue
			}
			return true
		default:
			if !l.CanLinkY(ix, xMaxYMin.y, xMinYMax.y) || !l.CanLinkX(xMaxYMin.y, xMaxYMin.x+1, ix) || !l.CanLinkX(xMinYMax.y, xMinYMax.x+1, ix) {
				continue
			}
			return true
		}
	}
	for iy := -1; iy <= l.yMax; iy++ { // Y方向
		if !l.CanLinkX(iy, xMinYMax.x, xMaxYMin.x) {
			continue // 这条线连不上,下面也不用判断了
		}
		switch {
		case iy == xMaxYMin.y || iy == xMinYMax.y: // 如果能成功上面X方向就返回了
		case iy < xMaxYMin.y:
			if !l.CanLinkY(xMinYMax.x, iy, xMinYMax.y-1) || !l.CanLinkY(xMaxYMin.x, iy, xMaxYMin.y-1) {
				continue
			}
			return true
		case iy < xMinYMax.y:
			if !l.CanLinkY(xMinYMax.x, iy, xMinYMax.y-1) || !l.CanLinkY(xMaxYMin.x, xMaxYMin.y+1, iy) {
				continue
			}
			return true
		default:
			if !l.CanLinkY(xMinYMax.x, xMinYMax.y+1, iy) || !l.CanLinkY(xMaxYMin.x, xMaxYMin.y+1, iy) {
				continue
			}
			return true
		}
	}
	return false
}

// CanLinkX (y,xMin)和(y,xMax),包含这两个点,能连线返回true
func (l *LianLianKan) CanLinkX(y, xMin, xMax int) bool {
	if y < 0 || y >= l.yMax {
		return true // 边界以外直接可以连线
	}
	if xMin < 0 {
		xMin = 0
	}
	// 越过边界,可以连线,不判断
	if xMax >= l.xMax {
		xMax = l.xMax - 1
	}
	for x := xMin; x <= xMax; x++ {
		if l.data[y][x] > NeedHandle {
			return false
		}
	}
	return true
}

// CanLinkY (yMin,x)和(yMax,x),包含这两个点,能连线返回true
func (l *LianLianKan) CanLinkY(x, yMin, yMax int) bool {
	if x < 0 || x >= l.xMax {
		return true // 边界以外直接可以连线
	}
	if yMin < 0 {
		yMin = 0
	}
	// 越过边界,可以连线,不判断
	if yMax >= l.yMax {
		yMax = l.yMax - 1
	}
	for y := yMin; y <= yMax; y++ {
		if l.data[y][x] > NeedHandle {
			return false
		}
	}
	return true
}
