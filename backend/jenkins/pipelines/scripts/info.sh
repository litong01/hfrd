#!/bin/bash
# cf script to show account information
# parameters are endpoint, apikey, org and space
bx login -a $endpoint --apikey $apikey -o $org -s $space
bx info

# All the following commands are the working commands
# bx login -a $endpoint --apikey $apikey -o $org -s $space
# bx info > /scripts/theinfo.txt
# bx cf cs $service $serviceplan tongliInstance
# bx cf service tongliInstance
# bx cf service noneexists
# bx cf create-service-key tongliInstance tongliInstanceKey
# bx cf service-key tongliInstance tongliInstanceKey
# bx cf delete-service-key tongliInstance tongliInstanceKey -f
# bx cf delete-service tongliInstance -f
# bx cf services
