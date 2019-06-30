#!/usr/bin/env bash


###############################################################################
#                                                                             #
#                              HFRD-TEST-WITH                               		  #
#                                                                             #
###############################################################################
WORKDIR=$(pwd)
. $WORKDIR/scripts/utils.sh

# Set all of the steps false as the default
NETWORK_CREATE=false
RETRIEVE_CREDS=false
CHANNEL_CREATE=false
PROFILE_CREATE=false
USERCERTS_CREATE=false
SCFILE_CREATE=false
INSTALL_TOOLS=false
WORKLOAD_DRIVE=false
NETWORK_DELETE=false
workload=samplecc

# Help function
Print_Help() {
	echo ""
	echo "Usage:"
	echo "./docker-entrypoint.sh [OPTIONS]"
	echo ""
	echo "Options:"
	echo "-a | --all  : Run all of the processes.You must supply the workload name(marbles/samplecc) that you want to run"
	echo "-n | --createNetwork  : Create a new network.Will return network.json and service.json"
	echo "-r | --retrieveCreds  : (Can't be used together with option '-n')To use existing network for measurements,we need to get the service credentials.You can use this option to retrieve credentials belonged to an existing network"
	echo "-c | --channels  : Create and join channels based on the channels configuration file"
	echo "-p | --profiles  : Fetch connection profiles"
	echo "-e | --enroll  : Enroll and upload user certs"
	echo "-s | --scfile  : Get PTE SCFile from connection files"
	echo "-i | --install  : Install SAR and NMON on network containers"
	echo "-w | --workload  : Set the PTE workload that you want to drive.You must supply the workload name(marbles/samplecc) that you want to run"
	echo "-d | --delete  : Delete network"
	exit 1
}

# Parse the input arguments
Parse_Arguments() {
	input=false
	while [ $# -gt 0 ]; do
		case $1 in
			--all | -a)
				NETWORK_CREATE=true
				CHANNEL_CREATE=true
				PROFILE_CREATE=true
				USERCERTS_CREATE=true
				SCFILE_CREATE=true
				INSTALL_TOOLS=true
				WORKLOAD_DRIVE=true
				input=true
				;;
			--createNetwork | -n)
				NETWORK_CREATE=true
				input=true
				;;
			--retrieveCreds | -r)
				RETRIEVE_CREDS=true
				input=true
				;;
			--channels | -c)
				CHANNEL_CREATE=true
				input=true
				;;
			--profiles | -p)
				PROFILE_CREATE=true
				input=true
				;;
			--enroll | -e)
				USERCERTS_CREATE=true
				input=true
				;;
			--scfile | -s)
				SCFILE_CREATE=true
				input=true
				;;
			--install | -i)
				INSTALL_TOOLS=true
				input=true
				;;
			--workload | -w)
				WORKLOAD_DRIVE=true
				input=true
				;;
			--delete | -d)
				NETWORK_DELETE=true
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

	# Conflict options check
	# Confilict options: '-n' and '-r'.
	if ${NETWORK_CREATE} && ${RETRIEVE_CREDS} ; then
		echo "Can't use option '-n' and '-r' at the same time"
		Print_Help
	fi
}

log() {
	printf "${PROG}  ${1}\n" | tee -a $HOME/results/logs/hfrd.log
}

cleanEnvironment(){
	rm -rf $HOME/results/*
}
createRequiredDirs(){
	if [[ ! -f $HOME/results/tmp ]];then
		mkdir -p $HOME/results/tmp
	fi
	if [[ ! -f $HOME/results/logs ]];then
		mkdir -p $HOME/results/logs
	fi
}

Parse_Arguments $@

createRequiredDirs

# Environments
PROG="[hfrd]"
source $WORKDIR/hfrd_test.cfg

if [[ $name == 'ep' ]]; then
	channelConfig=${WORKDIR}/conf/channels_EP.json
else
	channelConfig=${WORKDIR}/conf/channels_SP.json
fi

cd $WORKDIR/scripts
if $NETWORK_CREATE ; then
	log "---------------------------------"
	log "---------------------------------"
	log "Create ${env}-${name} Network 	  "
	log "---------------------------------"
	log "---------------------------------"
	cleanEnvironment
	createRequiredDirs
	./network_generator.sh -c $WORKDIR/hfrd_test.cfg
	if [ $? -ne 0 ]; then
		exit 2
	fi
fi

cd $WORKDIR/scripts
if $RETRIEVE_CREDS ; then
	log "-----------------------------------------------------"
	log "-----------------------------------------------------"
	log " Retrieve service credentials for service: $serviceid "
	log "-----------------------------------------------------"
	log "-----------------------------------------------------"
	cleanEnvironment
	createRequiredDirs
	./retrieve_creds.sh -s $serviceid -c $WORKDIR/hfrd_test.cfg
	if [ $? -ne 0 ]; then
		exit 2
	fi
fi

cd $WORKDIR
if $CHANNEL_CREATE ; then
	log "---------------------------------"
	log "---------------------------------"
	log "Create And Join channels 	      "
	log "---------------------------------"
	log "---------------------------------"
	./scripts/channels_generator.sh -c $channelConfig -n $HOME/results/creds/network.json
	if [ $? -ne 0 ]; then
		exit 2
	fi
fi

cd $WORKDIR/scripts
if $PROFILE_CREATE ; then
	log "---------------------------------"
	log "---------------------------------"
	log "Request connection profiles 	  "
	log "---------------------------------"
	log "---------------------------------"
	./connectionProfile_generator.sh -c $WORKDIR/hfrd_test.cfg -s $HOME/results/creds/service.json
	if [ $? -ne 0 ]; then
		exit 2
	fi
fi

cd $WORKDIR/scripts
if $USERCERTS_CREATE ; then
	log "---------------------------------"
	log "---------------------------------"
	log "Enroll user and upload certs  	  "
	log "---------------------------------"
	log "---------------------------------"
	ls $HOME/results/creds/ConnectionProfile_*
	if [ $? -ne 0 ]; then
		exit 1
	fi
	for profile in $( ls $HOME/results/creds/ConnectionProfile_* )
	do
		./user_enroller.sh -c $channelConfig -n $HOME/results/creds/network.json -p $profile
		if [ $? -ne 0 ]; then
			exit 2
		fi
	done
fi

cd $WORKDIR/scripts
if $SCFILE_CREATE ; then
	log "---------------------------------"
	log "---------------------------------"
	log "Generate Caliper Network file 			  "
	log "---------------------------------"
	log "---------------------------------"
	# Gather MSPIDs
	cp -r $HOME/results/creds/ $HOME/caliper/network/fabric/ibpep
	# Use python script to generate Caliper networkfiles. Results will be stored in ${HOME}/results/SCFiles/  with name 'config-net-${networkId}.json'
	cd $HOME/caliper/network/fabric/ibpep
	python ${WORKDIR}/scripts/ibp_fabric_config_gen.py 
	if [ $? -ne 0 ]; then
		exit 2
	fi
	cp ibpep_fabric.json $HOME/caliper/benchmark/simple 

fi

cd $WORKDIR/scripts
if $INSTALL_TOOLS ; then
	log "---------------------------------"
	log "---------------------------------"
	log "Install SAR and NMON   	      "
	log "---------------------------------"
	log "---------------------------------"
	./install_tools.sh -n $HOME/results/creds/network.json
fi

cd $WORKDIR/scripts
if $WORKLOAD_DRIVE ; then
	log "---------------------------------"
	log "---------------------------------"
	log "Drive Caliper workload: $workload	"
	log "---------------------------------"
	log "---------------------------------"
	
	if [[ $runMode == 'local' ]]; then
		./workload_driver.sh -w $workload -n $HOME/results/creds/network.json
		if [ $? -ne 0 ]; then
			exit 2
		fi
	else
		log "only support local run model "
	fi

fi

cd $WORKDIR/scripts
if $NETWORK_DELETE ; then
	log "---------------------------------"
	log "---------------------------------"
	log "Delete network by serviceid      "
	log "---------------------------------"
	log "---------------------------------"
	./network_delete.sh -c $WORKDIR/hfrd_test.cfg -s $HOME/results/creds/service.json
	if [ $? -ne 0 ]; then
		exit 2
	fi
fi