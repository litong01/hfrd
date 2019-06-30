#!/usr/bin/env bash

TOP=$(pwd)

docker build -t pte-hfrd .

docker run -v ${TOP}/results:/home/testuser/results pte-hfrd ./docker-entrypoint.sh
