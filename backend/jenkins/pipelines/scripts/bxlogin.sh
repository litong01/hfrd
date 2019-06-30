#!/bin/bash
# The script to login to bx
# This file will require the following parameters
# endpoint, apikey, org, space, service
mkdir -p $WORKDIR/results
bx login -a $endpoint --apikey $apikey -o $org -s $space
