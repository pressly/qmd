#!/bin/bash

for i in {1..$RAND}; do
    sleep 0.$RAND
    echo "Working on $i.."
done

echo "Done."
