#!/bin/sh

echo "** QMD test runner.. lets go! **"

curl -v -X POST --data '{}' "http://qmd:pass@localhost:8080/scripts/test.sh"

curl -v -X POST --data '{"args":["1"]}' "http://qmd:pass@localhost:8080/scripts/wait.sh"
