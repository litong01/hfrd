#!/usr/bin/env bash
#
#   Install SAR and NMON Tools
#   Description: Installs SAR and NMON on network containers.
#   Dependencies: network.json
#   Note: network.json contains each org's service credentials in blockchain network
#   One parameter :
#		1)path of network.json
. utils.sh

MAX_RETRY=2
PROG="[hfrd-install-tools]"
log() {
	printf "${PROG}  ${1}\n" | tee -a $HOME/results/logs/install_tools.log
}

# Help function
Print_Help() {
	log ""
	log "Usage:"
	log "./install_tools.sh [OPTIONS]"
	log ""
	log "Options:"
	log "-n | --networkPath           :   The path of network.json which contains all of the service credentials in blockchain network,including msp_id,networkId,API key/secret,"
}

# Parse the input arguments
Parse_Arguments() {
	while [ $# -gt 0 ]; do
		case $1 in
			--network | -n)
				shift
				networkPath=$1
				;;
			--help | -h)
				Print_Help
				;;
		esac
		shift
	done

	# If no components are passed, print error
	if [ -z "$networkPath" ]; then
		log "ERROR: Not enough parameters supplied."
		Print_Help
		exit 1
	fi

}

# Install SAR and NMON on containers
installTools(){
    lpar=$1
    container_id=$2
    log "Installing SAR and NMON on container $container_id on LPAR $lpar"
    sshpass -p "pass4chain" ssh -o StrictHostKeyChecking=no root@$lpar "docker-exec-ssh $container_id PATH=$PATH:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin apt-get update"
    statuscode1=`echo $?`
    sshpass -p "pass4chain" ssh -o StrictHostKeyChecking=no root@$lpar "docker-exec-ssh $container_id PATH=$PATH:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin apt-get install -y --force-yes sysstat nmon"
    statuscode2=`echo $?`
    sshpass -p "pass4chain" ssh -o StrictHostKeyChecking=no root@$lpar "docker-exec-ssh $container_id sed -i "s/false/true/" /etc/default/sysstat"
    statuscode3=`echo $?`
    if [ "$statuscode1" -eq 0 ] && [ "$statuscode2" -eq 0 ] && [ "$statuscode3" -eq 0 ]; then
		return 0
    else
		return 1
    fi
}

# Verify Successful Installation
checkTools(){
    lpar=$1
    container_id=$2
    log "Verifying SAR and NMON installation on container $container_id on LPAR $lpar"
    sshpass -p "pass4chain" ssh -o StrictHostKeyChecking=no root@$lpar "docker-exec-ssh $container_id sar 1 1"
    statuscode1=`echo $?`
    sshpass -p "pass4chain" ssh -o StrictHostKeyChecking=no root@$lpar "docker-exec-ssh $container_id nmon -? | head -n 2"
    statuscode2=`echo $?`
    if [ "$statuscode1" -eq 0 ] && [ "$statuscode2" -eq 0 ]; then
		return 0
    else
		return 1
    fi
}

Parse_Arguments $@

# Sanity check
if [[ ! -f $networkPath ]]; then
	log "Missing hfrd configuration file, cannot continue."
	exit 1
fi

msp_id=$(jq -r keys[0] $networkPath)
networkId=$(jq -r .$msp_id.network_id $networkPath)

#	Step 1: Install SAR and NMON tools on each container in network
#	Step 2: Verify that tools have been installed on each container
lparIPs=($(cat $HOME/conf/hosts | awk '{print $1}'))
for lpar_ip in "${lparIPs[@]}"; do
    peerIds=($(sshpass -p "pass4chain" ssh -o StrictHostKeyChecking=no root@$lpar_ip "docker ps | grep $networkId-fabric-peer | awk '{print \$1}'"))
    ordererIds=($(sshpass -p "pass4chain" ssh -o StrictHostKeyChecking=no root@$lpar_ip "docker ps | grep $networkId-fabric-orderer | awk '{print \$1}'"))
    kafkaIds=($(sshpass -p "pass4chain" ssh -o StrictHostKeyChecking=no root@$lpar_ip "docker ps | grep $networkId-kafka | awk '{print \$1}'"))
    zookeeperIds=($(sshpass -p "pass4chain" ssh -o StrictHostKeyChecking=no root@$lpar_ip "docker ps | grep $networkId-zookeeper | awk '{print \$1}'"))
    logstashIds=($(sshpass -p "pass4chain" ssh -o StrictHostKeyChecking=no root@$lpar_ip "docker ps | grep logstash | awk '{print \$1}'"))
    containerIds=(${peerIds[@]} ${ordererIds[@]} ${kafkaIds[@]} ${zookeeperIds[@]} ${logstashIds[@]})
    for container in "${containerIds[@]}"; do
        runWithRetry "installTools $lpar_ip $container"
    done

    for container in "${containerIds[@]}"; do
        runWithRetry "checkTools $lpar_ip $container"
    done
done