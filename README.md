# Go版 数字货币交易所

```
                       _______    ______   ________ 
                      |       \  /      \ |        \
    ______    ______  | $$$$$$$\|  $$$$$$\| $$$$$$$$
   /      \  /      \ | $$  | $$| $$   \$$| $$__    
  |  $$$$$$\|  $$$$$$\| $$  | $$| $$      | $$  \   
  | $$  | $$| $$  | $$| $$  | $$| $$   __ | $$$$$   
  | $$__| $$| $$__/ $$| $$__/ $$| $$__/  \| $$_____ 
   \$$    $$ \$$    $$| $$    $$ \$$    $$| $$     \
   _\$$$$$$$  \$$$$$$  \$$$$$$$   \$$$$$$  \$$$$$$$$
  |  \__| $$                                        
   \$$    $$                                        
    \$$$$$$                                         
                                                                       
```
```
此项目本是为Rfinex交易所准备的升级版本，因为领导决定全面转向RUST，所以下马了。
为了使这个项目代码不至于废弃，遂移出本企业相关的内容，仅保留交易核心功能开源出来。
此外，我和同事会继续更新此代码。
```

# Install Golang
```
wget https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz

sudo tar zxvf go1.11.2.linux-amd64.tar.gz -C /usr/local

vim ~/.bashrc
添加
export GOROOT=/usr/local/go
export GOPATH=/home/ubuntu/go # 在自己的主目录添加go目录
export PATH=$PATH:$GOPATH:/usr/local/go/bin

source ~/.bashrc

```

# 获取代码
```
go get github.com/oldfitter/goDCE
```
# 安装依赖
```
cd ~/go/src/github.com/oldfritter/goDCE 
 ./install.sh # 绝大多数情况下，你是需要科学上网才能完成依赖安装
```
# 导入前端代码
将前端代码导入public/assets目录下即可
# 编译
```
./build.sh
```
# 启动
```
./start.sh
```
# 停止
```
./stop.sh
```
# 重启
```
./restart.sh
```
