#!/bin/sh
go build -o ./cmd/api api/api.go
go build -o ./cmd/matching trade/matching.go
go build -o ./cmd/cancel order/cancel.go
go build -o ./cmd/treat trade/treat.go
go build -o ./cmd/workers workers/workers.go
