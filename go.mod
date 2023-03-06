module goOpenCv

go 1.19

replace github.com/go-opencv/go-opencv => ./OpenCvPath/src/github.com/go-opencv/go-opencv

require (
	github.com/go-opencv/go-opencv v0.0.0-00010101000000-000000000000
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e
	github.com/vova616/screenshot v0.0.0-20220801010501-56c10359473c
	golang.org/x/sys v0.6.0
)

require github.com/BurntSushi/xgb v0.0.0-20210121224620-deaf085860bc // indirect
