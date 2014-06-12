#!/bin/bash
##--
## Example test script for QMD
##--

# CLI args
username=$1
force_err=$2

if [ $# -lt 1 ]; then
  echo "Usage: $0 <username> [force_err]"
  echo "-"
  echo "username : Your username!"
  echo "force_err : Force the script to fail (for testing). Any value will make it err."
  exit 1
fi

echo "Exec'ing $0 $@ in $QMD_TMP"

if [ -n "$force_err" ]; then
  echo "oops.. forced error....."
  exit 1
fi

echo "Hello $username!"
echo "<some return value.. for $username>\n\n\n" > $QMD_OUT
cat "$QMD_TMP/test.txt" >> $QMD_OUT
exit 0
