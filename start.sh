#!/bin/sh
nohup ./bin/api >> logs/api.log &
nohup ./bin/matching >> logs/matching.log &
nohup ./bin/cancel >> logs/cancel.log &
nohup ./bin/treat >> logs/treat.log &
nohup ./bin/workers >> logs/workers.log &
