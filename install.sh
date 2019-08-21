#!/bin/sh

cp -r config/amqp.yml.example config/amqp.yml
cp -r config/database.yml.example config/database.yml
cp -r config/env.yml.example config/env.yml
cp -r config/interfaces.yml.example config/interfaces.yml
cp -r config/redis.yml.example config/redis.yml
cp -r config/workers.yml.example config/workers.yml

mkdir cmd
mkdir pids
mkdir logs
mkdir -p public/assets

rm -rf vendor
rm -rf Godeps
go get -u github.com/dafiti/echo-middleware
go get -u github.com/garyburd/redigo/redis
go get -u github.com/go-sql-driver/mysql
go get -u github.com/gomodule/redigo/redis
go get -u github.com/jinzhu/gorm
go get -u github.com/kylelemons/go-gypsy
go get -u github.com/labstack/echo
go get -u github.com/oldfritter/matching
go get -u github.com/qor/i18n
go get -u github.com/shopspring/decimal
go get -u github.com/streadway/amqp

go get -u golang.org/x/crypto
go get -u golang.org/x/sys/unix

go get -u gopkg.in/yaml.v2
# godep save
