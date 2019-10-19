@echo off

REM cd runPet
REM go build -i -ldflags "-s -w -H windowsgui"
REM .\runPet
REM cd ..

set openCvPath=%cd%\..\OpenCvPath
set openCvLib=%openCvPath%\lib

echo %GOPATH% | find "%openCvPath%" >nul
if %errorlevel% neq 0 ( REM 已经设置则不重复设置
  set GOPATH=%openCvPath%;%GOPATH%
)

REM 可将opencv*.dll拷贝到环境变量path中
echo %path% | find "%openCvLib%" >nul
if %errorlevel% neq 0 (
  set path=%openCvLib%;%path%
)

set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

@echo on

go build -i -ldflags "-s -w"
.\LianLianKan.exe
