#!/bin/bash

TIMES=$(( ( RANDOM % 30 )  + 1 ))
echo "Going to work $TIMES times."

for i in $(seq 1 $TIMES); do
    echo "Working on $i.."
    sleep 0.$RANDOM
done

echo "Done."
