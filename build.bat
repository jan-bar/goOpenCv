@echo on

set openCvLib=C:\opencv\build\install\x64\mingw\bin
echo "%PATH%" | find /i "%openCvLib%" >nul 2>nul || set "PATH=%openCvLib%;%PATH%"

set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

go build -ldflags "-s -w" -trimpath
