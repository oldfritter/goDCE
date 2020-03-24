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

## Dependencies
* MySql
* RabbitMQ
* Redis

# Install Golang
```
wget https://dl.google.com/go/go1.13.linux-amd64.tar.gz

sudo tar zxvf go1.13.linux-amd64.tar.gz -C /usr/local

vim ~/.bashrc
添加
export GOROOT=/usr/local/go
export PATH=$PATH:$GOPATH:/usr/local/go/bin

source ~/.bashrc

```

# 获取代码
```
git clone https://github.com/oldfritter/goDCE
```
# 单独启动api
```
cd goDCE
go run api/api.go // 此时将自动安装依赖
```
# 导入前端代码
将前端代码导入`goDCE/public/assets`目录下即可
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
