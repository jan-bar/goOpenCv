package analyze

import (
	"testing"

	"github.com/jan-bar/goOpenCv/mahjong/janbar-helper/util"
)

// go test -v -run TestAnalyze -o analyze.exe
// .\analyze.exe -test.v
func TestAnalyze(t *testing.T) {
	tiles := util.LoadTileFromFile("tile.tmp.txt")
	for _, tt := range tiles {
		Analyze(tt.Tile)
	}
}
