Testsuite Jenkins Implementation
================================

This implementation currently uses three jenkins jobs to help testers to setup environments, provide env details and run tests

Jenkins Jobs to accomplish various requests
===========================================

# network

This job takes 3 parameters to allow testers to create, delete and query a fabric
network. These 3 parameters are:

       method: indicate if the request is to create, query or remove the network, possible values are
           POST,DELETE,GET. The values are a direct map to Restful API POST, DELETE and GET request.
       serviceid: the desired service identification. If missing when create a network, then a UUID
           will be generated. For query and delete, this parameter is required.
       env: the desired environment, currently it can be bluemix staging env or bluemix production
           environment. The possible values for this are bxstaging, bxproduction, more environment
           can be supported once they become available.
           
This job produces possibly 3 files. They are:

       jobStatus.json    job status and links
       network.json      network key file
       service.json      service id and service key names

# connection

This job takes 2 parameters to allow testers to get connection profiles. These 2 parameters are:

       serviceid: indicate the service identification of the network in question. This should be the
          id received in the network requests.
       env: the desired environment, currently it can be bluemix staging env or bluemix production
           environment. The possible values for this are bxstaging, bxproduction, more environment
           can be supported once they become available.

This job produces 3 files. They are:

       jobStatus.json    job status and links
       package.tar       all the network key file, connection profiles etc
       
       
# test

This job runs the test that a tester actually uploads. It takes a json file as jenkins multiline
parameter as http query parameter. See below for detailes of this parameter.
       
This job will produce one file to indicate the results of the test run.


The file formats that this implementation requires or produces
==============================================================

# jobStatus.json

This file always gets produced by each job run. It should look like the following:

        {
          "status":"SUCCESS",
          "artifacts": {
            "network": "the url to the network.json file",
            "service": "the service id and service key"
          }
        }

Different jobs may return different items in the artifacts elements

# testconfig parameter when send request to run a test:

       testconfig: the parameter name with the following content:
       {
         "url": "http://gryphpxe1.pok.stglabs.ibm.com/workloads/hfrd_ibpth/soltest.tar.gz",
         "hash": "d69f4502b964bf90ec95e4256682b707",
         "startcmd": "test-entrypoint.sh",
         "sslcerts": true
       }
       
The url element indicates where the test code package will be downloaded. hash element
indicate what hash to use to verify the package and the startcmd indicate what script to
run to start the test, that has to be relative to the $WORKDIR which is always mounted
against the running container /opt/fabrictest, it is also the home directory of the user
which runs the test in the runner container. The hash part has not been implemented yet.
