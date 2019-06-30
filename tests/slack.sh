#!/usr/bin/env bash

function slack() {
	if [ "$1" == "green" ]; then
		color="#00FF00"
	elif [ "$1" == "red" ]; then
		color="#FF0000"
	elif [ "$1" == "blue" ]; then
		color="#0000FF"
	else
		echo "Invalid Slack Message Color"
		exit 1
	fi

	slack_message=$2
	slack_alert=$3

	if [[ -z ${slack_alert} ]]; then
		curl -s -X POST \
			-H "Content-type: application/json" \
			--data "{ \"channel\":\"#${SLACK_CHANNEL}\", \"username\": \"${SLACK_USERNAME}\", \"attachments\" : [ { \"text\" : \"${slack_message}\", \"color\": \"${color}\" } ]}" \
			${SLACK_URL} >/dev/null
	else
		curl -s -X POST \
			-H "Content-type: application/json" \
			--data "{ \"channel\":\"#${SLACK_CHANNEL}\", \"username\": \"${SLACK_USERNAME}\", \"text\": \"${slack_alert}\", \"attachments\" : [ { \"text\" : \"${slack_message}\", \"color\": \"${color}\" } ]}" \
			${SLACK_URL} >/dev/null
	fi

}
