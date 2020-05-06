# goOpenCv

#### 介绍
学习和使用opencv  
主要使用opencv2.4.9版本,通过cgo调用相关dll中的方法  

#### 安装教程

1.  将.\OpenCvPath\lib\*.dll拷贝到环境变量Path的任意一个目录  
2.  需要编译则要将.\OpenCvPath的绝对路径加到GOPATH中,  
    注意OpenCvPath里面有很多我添加或修改了,和github原作者的略微不一样  
3.  可以去.\OpenCvPath\src\github.com\go-opencv\go-opencv\samples看例子  
4.  另外分享一个好用工具,按键精灵带的工具"Picker.7z",用来分析窗口句柄和屏幕坐标很好用  

#### 使用说明

1.  没啥特殊要求,可以参考.\LianLianKan里面的例子,build.bat是编译脚本  
2.  由于是图像匹配,因此需要精确的坐标,因此需要进入.\LianLianKan\runPet中执行runPet.exe来启动swf游戏  
2.  启动游戏后不要改变窗口大小,点击开始游戏后,再cmd下执行LianLianKan.exe即可自动玩游戏咯  
3.  可以查看lianliankan.gif看看我的效果吧  

![lianliankan.gif](.\lianliankan.gif)
