#!/bin/sh
go build -o bin/api api/api.go
go build -o bin/matching trade/matching.go
go build -o bin/cancel order/cancel.go
go build -o bin/treat trade/treat.go
go build -o bin/workers workers/workers.go
