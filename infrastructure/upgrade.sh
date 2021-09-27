#!/bin/bash -e
go build .
NAME=remotemoe_$(date +"%Y-%m-%d_%H:%M:%S")
ssh remotemoe mv /usr/local/bin/remotemoe /tmp/${NAME}
scp remotemoe remotemoe:/usr/local/bin/remotemoe
ssh remotemoe "systemctl restart remotemoe && systemctl status remotemoe"
