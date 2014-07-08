#!/bin/bash

MIN=$1
MAX=$2
PID=$$

echo "Running under PID #$PID"

echo "Making file in tmp dir at $QMD_TMP"
'Test #$PID' >> $QMD_TMP/$PID

echo "Making file in store dir at $QMD_STORE"
'Test #$PID' >> $QMD_STORE/$PID

while [ $MAX -eq $MAX ]
do
    NUM=`shuf -i $MIN-$MAX -n 1`
    echo "Randomly selected $NUM"

    if [ $NUM -eq 7 ]
    then
        # Success
        echo "Job well done!"
        exit 0

    elif [ $NUM -eq 4 ]
    then
        # Error
        echo "Error! Error!"
        exit 1

    elif [ $NUM -eq $MAX ]
    then
        # Crash the script
        echo "Abandon ship! Abandon ship! Everyone for themselves!"
        kill -SIGHUP $$

    else
        # Sleep
        echo "zzzzzzzzzzzzzzzzz"
        sleep $NUM
        echo "zzzzzz..I'M AWAKE"
    fi
done
