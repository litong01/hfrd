#!/usr/bin/env bash

python  scripts/zipgen.py -n /input/network.json -o /output 
python  scripts/uploadcert.py -d /output