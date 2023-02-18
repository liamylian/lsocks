#!/bin/bash

export SOCKS_PORT=1080
export LOG_LEVEL=info
export LOG_FILE=main.log
export TRAFFICS_FILE=traffic.log
export CREDENTIALS=admin/admin,root/root


./worker &