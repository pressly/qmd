#!/bin/bash

echo "\$QMD_OUTPUT = " $QMD_OUT

echo "Running uname -a"
uname -a >> $QMD_OUT

echo "Running date"
date >> $QMD_OUT
