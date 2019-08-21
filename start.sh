#!/bin/sh
nohup ./cmd/api >> logs/api.log &
nohup ./cmd/matching >> logs/matching.log &
nohup ./cmd/cancel >> logs/cancel.log &
nohup ./cmd/treat >> logs/treat.log &
nohup ./cmd/workers >> logs/workers.log &
