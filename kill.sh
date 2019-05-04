#!/bin/bash

ps ax | grep 'go' | grep -v 'vim' | grep -e 'authserver\|resource\|client' | cut -d' ' -f1 | while read pid
do
    echo "kill $pid"
    kill $pid 2>/dev/null
done
