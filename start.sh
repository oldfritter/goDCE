#!/bin/sh
nohup ./bin/api >> logs/api.log &
nohup ./bin/matching >> logs/matching.log &
nohup ./bin/cancel >> logs/cancel.log &
nohup ./bin/trade >> logs/trade.log &
nohup ./bin/workers >> logs/workers.log &
