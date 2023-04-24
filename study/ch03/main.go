package main

import (
	"flag"
	"fmt"
	"log"

	"gocv.io/x/gocv"
)

/*
官网: https://www.dev47apps.com/

Android下载: https://play.google.com/store/apps/details?id=com.dev47apps.droidcam

windows下载: https://www.dev47apps.com/droidcam/windows/

手机USB连接后选择 [文件传输] ,并且开启USB调试
在window客户端选择USB方式,刷新时手机会提示USB调试授权
确定后,window上就可以预览了,window客户端需要一直开启
其他软件才可以正常使用摄像头,window客户端可以点右下角按钮静默不显示图片
此时其他软件仍可以正常使用摄像头,window客户端停止后其他软件无法打开摄像头

如果使用WiFi方式接入,需要window客户端能访问Android的ip地址

Android运行时需要摄像头等相关权限,也需要正常授权才能用，我已经把相关文件打包存到百度网盘了
*/

var (
	window = gocv.NewWindow("janbar")

	deviceID int
)

func main() {
	//goland:noinspection GoUnhandledErrorResult
	defer window.Close()

	flag.IntVar(&deviceID, "d", 1, "camera ID")
	flag.Parse()

	err := video()
	if err != nil {
		log.Fatal(err)
	}
}

//goland:noinspection GoUnhandledErrorResult
func video() error {
	// 一般笔记本内置摄像头deviceID=0
	// 使用DroidCam模拟摄像头从1开始
	// 貌似安装客户端以后不打开也能读到deviceID=1的摄像头只不过是纯色图片
	// 尝试打开deviceID=2的摄像头会失败
	cam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		return fmt.Errorf("error opening video capture device: %v\n", deviceID)
	}
	defer cam.Close()

	h := cam.Get(gocv.VideoCaptureFrameHeight)
	w := cam.Get(gocv.VideoCaptureFrameWidth)
	log.Println(h, w)

	// cam.Set(gocv.VideoCaptureFrameHeight, 640)
	// cam.Set(gocv.VideoCaptureFrameWidth, 480)

	img := gocv.NewMat()
	defer img.Close()

	if !cam.Read(&img) {
		return fmt.Errorf("cam.Read false")
	}
	if img.Empty() {
		return fmt.Errorf("img empty")
	}

	// 将摄像头图像保存为视频文件
	out, err := gocv.VideoWriterFile("output.avi", "MJPG", 25, img.Cols(), img.Rows(), true)
	if err != nil {
		return err
	}
	defer out.Close()

	for cam.IsOpened() {
		if !cam.Read(&img) {
			return fmt.Errorf("cam.Read false")
		}
		if img.Empty() {
			continue
		}

		// flipCode(翻转方向): 1:水平翻转,0:垂直翻转,-1:水平垂直翻转
		gocv.Flip(img, &img, 1)

		out.Write(img) // 将图像帧写入文件

		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
	return nil
}
