# goOpenCv

#### 介绍
该例子为使用opencv匹配连连看上各个动物图片,得到数据,根据算法控制鼠标自动玩游戏  

#### 安装教程

准备gocv环境

可以下我编译的: [下载地址](https://github.com/jan-bar/go-opencv/releases/download/v0.0.1/opencv.7z)

解压后建立c盘链接: `mklink /j c:\opencv xxx\opencv`

或者按照gocv文档编译

[文档](https://gocv.io/getting-started/windows/)

下载gcc环境: [下载地址](https://sourceforge.net/projects/mingw-w64/files/),选择最新版`x86_64-posix-seh`

下载cmake: [下载地址](https://cmake.org/download/),可以选择zip免安装版本

然后根据脚本进行编译: [编译脚本](https://github.com/hybridgroup/gocv/blob/release/win_build_opencv.cmd)

`set http_proxy=127.0.0.1:1081 & set https_proxy=127.0.0.1:1081` 设置代理中途需要下载GitHub资源

`wget https://github.com/opencv/opencv/archive/4.7.0.zip -O opencv-4.7.0.zip` 解压到当前目录

`wget https://github.com/opencv/opencv_contrib/archive/4.7.0.zip -O opencv_contrib-4.7.0.zip` 解压到当前目录

最好执行`mklink /j c:\opencv xxx\opencv`,保证脚本使用都是`c:\opencv`路径,包括gcc和cmake工具

`set enable_shared=ON` 使用动态dll编译,记得cmake命令里面几个路径改为自己需要的

`set enable_shared=OFF` 使用静态编译,记得cmake命令里面几个路径改为自己需要的

编译完成后将`install`路径按照gocv要求弄好,做个压缩包存起来也可以

注意编译出来的可执行程序还依赖`libwinpthread-1.dll,libstdc++-6.dll,libgcc_s_seh-1.dll`这3个dll,一般安装window的git就有

不然还得将上面解压的gcc环境里的这3个dll路径添加到PATH环境变量中


#### 使用说明

1.  由于是图像匹配,因此需要精确的坐标,因此需要进入.\LianLianKan\runPet中执行runPet.exe来启动swf游戏
      安装: [flash](https://www.flash.cn/compatibility)
      [下载地址](https://www.flash.cn/cdm/latest/flashplayerax_install_cn_web.exe)
2.  启动游戏后不要改变窗口大小,点击开始游戏后,再cmd下执行LianLianKan.exe即可自动玩游戏咯  
      需要下载[msvcp120.dll](https://www.dll-files.com/msvcp120.dll.html)
      需要下载[msvcr120.dll](https://www.dll-files.com/msvcr120.dll.html)
3.  可以查看lianliankan.gif看看我的效果吧  
