package main

import (
	"flag"
	"fmt"

	"github.com/go-opencv/go-opencv/opencv"
)

// https://stackoverflow.com/questions/32737420/multiple-results-in-opencvsharp3-matchtemplate

func main() {
	s := flag.String("src", "", "src img")
	t := flag.String("templ", "", "templ img")
	flag.Parse()

	src := opencv.LoadImage(*s)
	temp := opencv.LoadImage(*t)

	w, h := src.Width()-temp.Width()+1, src.Height()-temp.Height()+1
	ftmp := opencv.CreateImage(w, h, 32, 1)

	var minVal, maxVal float64
	var minloc, maxLoc opencv.CvPoint
	opencv.MatchTemplate(src, temp, ftmp, opencv.CV_TM_CCOEFF_NORMED)
	opencv.Threshold(ftmp, ftmp, 0.8, 1.0, opencv.CV_THRESH_TOZERO)

	for {
		opencv.MinMaxLoc(ftmp, &minVal, &maxVal, &minloc, &maxLoc, nil)
		fmt.Println(minVal, maxVal, minloc, maxLoc)
		if maxVal >= 0.8 {
			pt1 := maxLoc.ToPoint()
			pt2 := opencv.Point{X: pt1.X + temp.Width(), Y: pt1.Y + temp.Height()}
			opencv.Rectangle(src, pt1, pt2, opencv.NewScalar(0, 0, 255, 0))
			opencv.FloodFill(ftmp, maxLoc, opencv.ScalarAll(0),
				opencv.ScalarAll(0.1), opencv.ScalarAll(1.0),
				nil, 4, nil)
		} else {
			break
		}
	}
	win := opencv.NewWindow("src", 1)
	defer win.Destroy()

	win.ShowImage(src)

	opencv.WaitKey(0)
}
