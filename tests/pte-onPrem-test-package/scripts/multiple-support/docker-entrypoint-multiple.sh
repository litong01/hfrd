#!/usr/bin/env bash


###############################################################################
#                                                                             #
#                              HFRD-TEST                              		  #
#                                                                             #
###############################################################################
WORKDIR=$(pwd)
. $WORKDIR/scripts/utils.sh

# Set all of the steps false as the default
NETWORK_CREATE=false

# Help function
Print_Help() {
	echo ""
	echo "Usage:"
	echo "./docker-entrypoint.sh [OPTIONS]"
	echo ""
	echo "Options:"
	echo "-n | --createNetwork  : Create a new network.Will return network.json and service.json"
	exit 1
}

# Parse the input arguments
Parse_Arguments() {
	input=false
	while [ $# -gt 0 ]; do
		case $1 in
			--createNetwork | -n)
				NETWORK_CREATE=true
				input=true
				;;
			--help | -h)
				input=true
				Print_Help
				;;
		esac
		shift
	done

	if ! $input; then
		echo "Must provide at least one argument"
		Print_Help
	fi

}

log() {
	printf "${PROG}  ${1}\n" | tee -a ${dir_networks}logs/hfrd.log
}

Parse_Arguments $@

# Environments
PROG="[hfrd]"
# Load the configurations
create_variables $WORKDIR/conf/networks.yaml

if [[ ! -f ${dir_networks}logs ]];then
	mkdir -p ${dir_networks}logs
fi

if [[ $name == 'ep' ]]; then
	channelConfig=${WORKDIR}/conf/channels_EP.json
else
	channelConfig=${WORKDIR}/conf/channels_SP.json
fi

cd $WORKDIR/scripts
if $NETWORK_CREATE ; then
	log "---------------------------------"
	log "---------------------------------"
	log "Create multiple Networks 	  	  "
	log "---------------------------------"
	log "---------------------------------"
	./multiple-support/network_generator_multiple.sh -c $WORKDIR/conf/networks.yaml
fi