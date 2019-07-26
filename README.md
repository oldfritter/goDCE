# Go版交易所

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