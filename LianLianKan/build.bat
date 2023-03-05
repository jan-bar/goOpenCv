@echo on

set openCvLib=%cd%\..\OpenCvPath\lib

REM 可将opencv*.dll拷贝到环境变量path中
REM 或者按照下面方式将dll目录添加到path中
echo "%PATH%" | find "%openCvLib%" >nul
if %errorlevel% neq 0 (
  set "PATH=%openCvLib%;%PATH%"
)

set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

go build -ldflags "-s -w" -o LianLianKan.exe lianliankan.go
.\LianLianKan.exe
