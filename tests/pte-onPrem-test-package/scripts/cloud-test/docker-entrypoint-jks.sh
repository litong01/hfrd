#!/usr/bin/env bash

printf "

███████████  ████████████      ████████       ████████
███████████  ███████████████   █████████     █████████
   █████        ████   █████     ████████   ████████
   █████        ███████████      ████ ████ ████ ████
   █████        ███████████      ████  ███████  ████
   █████        ████   █████     ████   █████   ████
███████████  ███████████████   ██████    ███    ██████
███████████  ████████████      ██████     █     ██████

"

PROG="[HFRD-PTE-Test]"

####################
# Helper Functions #
####################
log() {
	printf "${PROG}  ${1}\n" | tee -a run.log
}

# Prepare files
ls -ltr

cp -f ${HOME}/config-chan1-TLS.json ${HOME}/fabric-sdk-node/test/PTE/SCFiles/config-chan1-TLS.json
cp -r ${HOME}/userInputs ${HOME}/fabric-sdk-node/test/PTE/userInputs

source ${HOME}/.nvm/nvm.sh
cd ${HOME}/fabric-sdk-node/test/PTE
##################################
# STEP 5 - PTE install chaincode #
##################################
log "---------------------------------"
log "---------------------------------"
log "Installing PTE Chaincode."
log "---------------------------------"
log "---------------------------------"
./pte_driver.sh userInputs/runCases.install.txt | tee -a run.log
sleep 5


######################################
# STEP 6 - PTE instantiate chaincode #
######################################
log "---------------------------------"
log "---------------------------------"
log "Instantiating PTE Chaincode."
log "---------------------------------"
log "---------------------------------"
./pte_driver.sh userInputs/runCases.instantiate.txt | tee -a run.log
sleep 30


####################
# STEP 7 - PTE run #
####################
log "---------------------------------"
log "---------------------------------"
log "Running PTE Now!"
log "---------------------------------"
log "---------------------------------"
./pte_driver.sh userInputs/runCases.run.txt | tee -a run.log

cp run.log ${HOME}/results
mkdir -p ${HOME}/results/creds
cp -r ${HOME}/creds/* ${HOME}/results/creds

sleep 120