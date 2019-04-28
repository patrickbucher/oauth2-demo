#!/bin/sh

ps a | grep 'go' | grep -e 'authserver\|resource\|client' | cut -d' ' -f1 | while read pid
do
    echo "kill $pid"
    kill $pid
done
