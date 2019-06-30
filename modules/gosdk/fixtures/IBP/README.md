Test Modules for IBP  Test 
 ===============

        
## Introduction

##Support-envs
## How to run pte test package
1. edit hfrd_test.cfg  to specify the parameters for IBP fabric network creation. 
2. create IBP fabric network 
    > example command: createIBPNetwork.sh -n  
    > results:
        > create a IBP network and save  service.json and network.json into $HOME/results/creds
3. cd scripts
4. generated network zip file.
   > example command: python zipgen.py -n $HOME/results/creds/network.json
   > results: 
        > generated a zip file in current dir , file name format : CP_{networkid}.zip
5. upload admin certs into IBP network 
   > example command:  python uploadcert.py  -d  zipoutput
   > results: 
        > admin user certs unploaded into IBP network
6. restart all IBP network peers   
   > example command:  python restartPeers.py  -d  zipoutput
   > results: 
       > all peers in IBP network restarted 

