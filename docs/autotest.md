# HFRD autotest job
In **HFRD**, you can schedule some test jobs which will run automatically based on you defined schedule. For example, a job can run every night at 1:00am. The job will stand up a fabric network according to your specification and run serious of the tests that you provide, HFRD then will save these test results produced by the tests, eventually clean up fabric network from the k8s cluster and everything produced during the test except the test results.

### Prepare to run a HFRD auto test
To run the automated test, you will need at least one k8s cluster, a network spec file and a test plan. In some cases, you may wish to run fabric network and your tests in different k8s clusters, then you will need to have two k8s clusters. Kubeconfig file from these clusers will need to be provided. The following is a chart how to organize these files:

```
    org2p4
        initdir
            chaincode.tgz
            kubeconfigclient.zip
            kubeconfigserver.zip
            networkspec.yml
            testplan1.yml
            testplan2.yml    
```

You can name the top directory anything you like, but it will be very useful if you name it something meaningful. Like the above, it was named as "org2p4", means the fabric network contains 2 orgs and 4 peers, every test run defined under this directory will be against that kind of fabric network. Under this directory, there has to be a directory named initdir which contains chaincode, network spec, kube configuration files and test plans. The names of each artifacts are very important, if you name these things differently, HFRD won't be able to use them, thus an error will occur. Chaincode has to be packaged as a tgz file, the content of the chaincode can be any go chaincode, but the file must be named chaincode.tgz in tgz format. kubeconfigclient.zip and kubeconfigserver.zip are kube configuration zip files, kubeconfigclient.zip is the kube configuration file of the k8s cluster for running hfrd tests. kubeconfigserver.zip is the kube configuration file of the k8s cluster for standing up fabric network. They can be the same file if you run both hfrd test and fabric network on the same k8s cluster, even in that case, you will still need to have two files, but they are basically the copy of each other. Networkspec.yml file is a regular HFRD network spec file which defines how the fabric network should look like. test plan yaml files are regular HFRD test plan files, they have to be named in a pattern like testplanxxx.yml. For example, testplan1.yml, testplan2.yml, are valid test plan files. the numeric name after word testplan is very important, the number decide the sequence of these tests.

Once you setup these files in this format, you will need to place these files using the structure described above on hfrd server's contentRepo directory. Normally that directory is ~/hfrd/contentRepo if you setup hfrd server using the hfrd setup program.

### Schedule a test job to run on a regular bases
Once you've done the prepare step described above, you can log in to hfrd jenkins server to schedule a job. It will be the easiest to simply create a new pipeline job by copying a job comes with HFRD named autotest. Let's you create a new one by copying the autotest job, named MyAutoTest, then you can configure MyAutoTest job uisng any jenkins provided scheduling capabilities.

When you change the schedule of your newly created job, you must provide a default value for the two parameters, the first parameter is called "uid", the default value should be "org2p4" if you meant to use the artifacts in the preparation step above. Of course, you can use string there, but that string should match a directory under ~/hfrd/contentRepo, and in that directory you should have a directory named initdir which contains necessary files described in the preparation step. The second parameter has the default value already which you normally should not change.

### Inspect the test run results
After you schedule the job, you can either wait for the job to run at specified time or event or you can simply kick it off yourself. If everything goes as you expected, the job will be started, fabric network will be stood up, tests will run and results will be recorded.

You can find the results of each job on the apache server which gets setup by HFRD setup script. Normally it will be on the same machine at port 9696. Simply hit that port with your uid in the url, you should see the results, here is an example url

```
   http://server:9696/org2p4
```
You should see the initdir and the testresults directory. Directory testresults contains all the test results, no matter how many times that your test job run, all the test results will be in this directory. In the testresults directory, you will be able to find many sub directories, each represent a test defined in your testplanxxx.yml file.
