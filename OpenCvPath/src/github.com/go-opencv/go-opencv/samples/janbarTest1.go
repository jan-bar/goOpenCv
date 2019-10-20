package main

import (
	"flag"

	"github.com/go-opencv/go-opencv/opencv"
)

func main() {
	s := flag.String("src", "", "src img")
	flag.Parse()
	src := opencv.LoadImage(*s)
	defer src.Release()

	r := opencv.NewRect(100, 100, 15, 10)
	tmp := opencv.CreateImage(r.Width(), r.Height(), src.Depth(), src.Channels())
	defer tmp.Release()

	src.SetROI(r)
	opencv.Copy(src, tmp, nil)
	src.ResetROI()
	opencv.SaveImage("janbar.jpg", tmp, nil)
}
