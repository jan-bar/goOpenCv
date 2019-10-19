package main

import (
	"errors"
	"image"
	"image/png"
	"os"
	"unsafe"

	"github.com/go-opencv/go-opencv/opencv"
	"github.com/lxn/win"
	"github.com/vova616/screenshot"
)

func main() {
	l := NewLianLianKan()
	l.ClickLeft(0, 0)
	l.CaptureAndLoadPic()
}

type LianLianKan struct {
	X, Y        int32
	CaptureName string

	image *opencv.IplImage
}

func NewLianLianKan() *LianLianKan {
	winHwnd := win.FindWindow(nil, win.StringToBSTR("janbarLianLianKan"))
	win.SetWindowPos(winHwnd, win.HWND_TOPMOST, // 居中且置顶窗口
		(win.GetSystemMetrics(win.SM_CXFULLSCREEN)-1500)/2,
		(win.GetSystemMetrics(win.SM_CYFULLSCREEN)-1000)/2,
		0, 0, win.SWP_NOSIZE)
	var RectPos, ClientPos win.RECT
	win.GetWindowRect(winHwnd, &RectPos)
	win.GetClientRect(winHwnd, &ClientPos)
	return &LianLianKan{ // 游戏左上角坐标
		X:           RectPos.Right - ClientPos.Right + 49,
		Y:           RectPos.Bottom - ClientPos.Bottom + 75,
		CaptureName: "lianliankan.png",
	}
}

// 左键点击一次
func (l *LianLianKan) ClickLeft(x, y int32) {
	win.SetCursorPos(l.X+x, l.Y+y)
	input := win.MOUSE_INPUT{
		Type: win.INPUT_MOUSE,
		Mi: win.MOUSEINPUT{
			DwFlags: win.MOUSEEVENTF_LEFTDOWN | win.MOUSEEVENTF_LEFTUP,
		},
	}
	win.SendInput(1, unsafe.Pointer(&input), int32(unsafe.Sizeof(input)))
}

// 截屏游戏数据,保存成一张图片
func (l *LianLianKan) CaptureAndLoadPic() error {
	x0, y0 := int(l.X), int(l.Y) /*47*14,*42*10*/
	img, err := screenshot.CaptureRect(image.Rect(x0, y0, x0+658, y0+420))
	if err != nil {
		return err
	}
	fw, err := os.Create(l.CaptureName)
	if err != nil {
		return err
	}
	err = png.Encode(fw, img)
	fw.Close() // 在读取之前关闭文件
	if err != nil {
		return err
	}

	l.image = opencv.LoadImage(l.CaptureName)
	if l.image == nil {
		return errors.New("LoadImage fail")
	}
	img, err = screenshot.CaptureRect(image.Rect(x0+10, y0+10, x0+100, y0+100))
	if err != nil {
		return err
	}
	fw, err = os.Create("t.png")
	if err != nil {
		return err
	}
	png.Encode(fw, img)
	fw.Close()
	return nil
}

func (l *LianLianKan) Close() error {
	if l.image != nil {
		l.image.Release()
	}
	return nil
}
