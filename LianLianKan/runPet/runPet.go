package main

import (
	"flag"
	"os"
	"path/filepath"

	. "github.com/lxn/walk/declarative"
)

// go build -i -ldflags "-s -w -H windowsgui"

func main() {
	w := flag.Int("w", 1500, "Width")
	h := flag.Int("h", 1000, "Height")
	flag.Parse()
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	_, err = MainWindow{
		Title:   "janbarLianLianKan",
		MinSize: Size{Width: *w, Height: *h},
		Size:    Size{Width: *w, Height: *h},
		Layout:  VBox{MarginsZero: true},
		Children: []Widget{
			WebView{
				Name: "wv",
				URL:  filepath.Join(dir, "pet.swf"),
			},
		},
	}.Run()
	if err != nil {
		panic(err)
	}
}
