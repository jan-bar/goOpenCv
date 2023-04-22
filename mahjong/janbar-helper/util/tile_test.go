package util

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

// 编译为可执行程序,输出从文件读取,方便快速测试
// go test -v -run TestTile -o tile.exe
// .\tile.exe -test.v  可执行程序进行测试
func TestTile(t *testing.T) {
	rbData, err := os.ReadFile("win_rb.tmp.txt")
	if err == nil {
		// 当ruby生成文件存在时才进行比较
		winTableRb := loadWinTableData(bytes.NewReader(rbData))
		// 打印两种方案结果长度
		t.Log(len(winTable), len(winTableRb))
		for k, v := range winTable {
			vv, ok := winTableRb[k]
			if ok {
				if fmt.Sprint(v) != fmt.Sprint(vv) {
					// 两种方案相同key但不同val时需要报错
					t.Fatalf("key:%x,go:%x,ruby:%x", k, v, vv)
				}
			} else if v[0] != 1<<26 && (v[0]&1<<26) != 0 {
				// go生成的不同于ruby数据只有龙七对的key
				// 当发现不是七对(1<<26)的值,但是(1<<26 bit)没有置为七对类型报错
				t.Fatalf("key:%x,go:%x,ruby:%x", k, v, vv)
			}
		}
	}

	tiles := LoadTileFromFile("tile.tmp.txt")
	for _, tt := range tiles {
		if Tile34IsHu(tt.Tile) {
			res0 := Tiles34Result(tt.Tile)
			res1 := Tiles34Backtrack(tt.Tile)
			t.Log(tt.Line[0], res0, res1)
			if fmt.Sprint(res0) != fmt.Sprint(res1) {
				t.Fatal("res0 != res1")
			}
		} else {
			t.Log(tt.Line[0], "no hu card")
		}
	}
}
