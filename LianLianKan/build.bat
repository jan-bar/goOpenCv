@echo on

REM 获取该模块实际目录,里面有可用的dll文件
for /F %%i in ('go list -f "{{.Dir}}" "github.com/jan-bar/go-opencv/opencv"') do (
  set openCvLib=%%i\lib
)

REM 可将opencv*.dll拷贝拷贝到PATH目录,或者设置PATH目录
echo "%PATH%" | find /i "%openCvLib%" >nul 2>nul || set "PATH=%openCvLib%;%PATH%"

REM 没运行swf游戏,则启动游戏
tasklist /FI "IMAGENAME eq flashplayer.exe" /FO CSV /NH | find /i "flashplayer.exe" >nul 2>nul || start runPet\flashplayer.exe runPet\pet.swf

set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

go build -ldflags "-s -w" -o LianLianKan.exe lianliankan.go
.\LianLianKan.exe
