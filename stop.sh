#!/bin/sh

cat pids/matching.pid  | xargs kill -INT

cat pids/cancel.pid  | xargs kill -INT

cat pids/treat.pid  | xargs kill -INT

cat pids/workers.pid  | xargs kill -INT

cat pids/api.pid  | xargs kill -INT
