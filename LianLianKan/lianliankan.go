package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"time"
	"unsafe"

	"github.com/go-opencv/go-opencv/opencv"
	"github.com/lxn/win"
	"github.com/vova616/screenshot"
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
		picName    string  // 截屏图片名称
		xMax, yMax int     // xMax列,yMax行
		data       [][]int // 保存每个点数据
		picCnt     int     // 相同图片编号

		image, temp *opencv.IplImage
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
	// 根据上面注释,获取第DWMWA_EXTENDED_FRAME_BOUNDS项数据
	const ExtendedFrameBounds = 9
	var RectPos win.RECT
	err := windows.DwmGetWindowAttribute(windows.HWND(window), ExtendedFrameBounds,
		unsafe.Pointer(&RectPos), uint32(unsafe.Sizeof(RectPos)))
	if err != nil {
		return nil, err
	}

	l := &LianLianKan{
		X:       RectPos.Left + 57, // 窗口左上角X坐标偏移一定值到游戏界面左上角X坐标
		Cx:      44,                // 经过计算2个图标中间横向长度
		Y:       RectPos.Top + 110, // 窗口左上角Y坐标偏移一定值到游戏界面左上角Y坐标
		Cy:      40,                // 经过计算2个图标中间纵向宽度
		picName: "LianLianKan.png",
		xMax:    14, yMax: 10,
	}
	l.data = make([][]int, l.yMax)
	for i := 0; i < l.yMax; i++ {
		l.data[i] = make([]int, l.xMax)
	}
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
}

// CaptureAndLoadPic 截屏游戏数据,保存成一张图片
func (l *LianLianKan) CaptureAndLoadPic() error {
	x0, y0 := int(l.X), int(l.Y) // 根据Picker.exe抓取坐标计算出来的值
	err := SavePng(x0, y0, x0+615, y0+399, l.picName)
	if err != nil {
		return err
	}
	l.Release() // 释放上一次资源
	l.image = opencv.LoadImage(l.picName)
	if l.image == nil {
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
	// 通过 opencv.SaveImage 将单个动物图像来调整下面4个数据
	x, y, w, h := ix*l.Cx, iy*l.Cy, l.Cx-1, l.Cy-1

	if l.temp == nil {
		l.temp = opencv.CreateImage(w, h, l.image.Depth(), l.image.Channels())
	}
	l.image.SetROI(opencv.NewRect(x, y, w, h))
	opencv.Copy(l.image, l.temp, nil)

	// 需要调整单个小动物截图结果时,用下面代码保存图片,调试上面x,y,w,h的值
	// ok := opencv.SaveImage("one.png", l.temp, nil)
	// if ok != 1 {
	// 	panic(fmt.Sprint("opencv.SaveImage", ok))
	// }

	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			if l.temp.Get2DIndex(i, j, opencv.ScalarR) > 0 {
				return true // r > 0表示该位置有图片
			}
		}
	}
	return false
}

// MatchTemp 找到name图片在游戏区域所有匹配位置
func (l *LianLianKan) MatchTemp() {
	l.image.ResetROI() // 每次匹配时复位

	res := opencv.CreateImage(
		l.image.Width()-l.temp.Width()+1,
		l.image.Height()-l.temp.Height()+1,
		32, 1)

	// 下面调用图片匹配,且只保留一定阈值的结果数据
	opencv.MatchTemplate(l.image, l.temp, res, opencv.CV_TM_CCOEFF_NORMED)
	opencv.Threshold(res, res, 0.8, 1.0, opencv.CV_THRESH_TOZERO)

	var (
		minVal, maxVal float64
		minLoc, maxLoc opencv.CvPoint
	)
	for {
		opencv.MinMaxLoc(res, &minVal, &maxVal, &minLoc, &maxLoc, nil)
		if maxVal < 0.8 {
			break // 低于阈值,没有能匹配图片
		}
		// 从图片坐标得到data的坐标,然后设置对应位置图片编号
		pos := maxLoc.ToPoint()
		// 坐标除以每个图标的长度和宽度,得到数组下标
		ix, iy := (pos.X+10)/l.Cx, (pos.Y+10)/l.Cy
		l.data[iy][ix] = l.picCnt

		opencv.FloodFill(res, maxLoc, opencv.ScalarAll(0),
			opencv.ScalarAll(0.1), opencv.ScalarAll(1.0),
			nil, 4, nil)
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
	if l.image != nil {
		l.image.Release()
		l.image = nil
	}
	if l.temp != nil {
		l.temp.Release()
		l.temp = nil
	}
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
		} // 每次都产生新的列表,将已经消除的去掉

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

// SavePng 屏幕截图,保存为png图片
func SavePng(x0, y0, x1, y1 int, name string) error {
	img, err := screenshot.CaptureRect(image.Rect(x0, y0, x1, y1))
	if err != nil {
		return err
	}
	fw, err := os.Create(name)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer fw.Close()
	return png.Encode(fw, img) // 保存png图片
}
