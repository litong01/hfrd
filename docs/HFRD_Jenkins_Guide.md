HFRD Jenkins Guide
 ===============
The document aims to show how to set up a new jenkins backend server and how to deploy HFRD jenkins jobs.

* [Prerequisites](#Prerequisites)
* [Set up a new Jenkins server](#Set-up-a-new-Jenkins-server)
* [Deploy HFRD Jenkins jobs](#Deploy-HFRD-Jenkins-jobs)
* [Manage default config files](#Manage-default-config-files)

## Prerequisites
To set up a new jenkins server the following prerequisites must be installed firstly:

    * Docker
    * Jenkins Job Builder(https://docs.openstack.org/infra/jenkins-job-builder/)

Jenkins Job Builder uses jenkins job configuration file(YAML/JSON) to configure jenkins. To install jenkins job builder:

    sudo apt-get install python-virtualenv
    virtualenv hyp
    source hyp/bin/activate
    pip install 'jenkins-job-builder==2.0.0'
    jenkins-jobs --version

Note: Not do this in root account
## Set up a new Jenkins server

Step 1: Start a new jenkins server:

    docker run -p 9090:8080 -p 50000:50000 -v /home/ibmadmin/jenkins:/var/jenkins_home -d --name jenkins jenkins/jenkins:lts


Step 2: Access jenkins dashboard: `http://jenkins_server_url:9090`
Then install the suggested jenkins plugins.
Note: You can find the initial jenkins password in
`/var/jenkins_home/secrets/initialAdminPassword`

## Deploy HFRD Jenkins jobs
Totally there are two ways to deploy a new jenkins job:

    * Use Jenkins Job Builder(recommend)
    * Manually add a new job in jenkins dashboard
We strongly recommend use the Jenkins Job Builder to deploy new jenkins jobs.

##### Step 1: Configure the jenkins.ini to your own jenkins server

Backup the jenkins.ini.example to jenkins.ini

`cp hfrd/jjb/jenkins.ini.example hfrd/jjb/jenkins.ini`

After copying the jenkins.ini.example, modify jenkins.ini with your Jenkins username, API token and Jenkins URL

    [job_builder]
    ignore_cache=True
    keep_descriptions=False
    include_path=.:scripts:~/git/
    recursive=True

    [jenkins]
    user=admin <Provide your Jenkins  username>
    password= <Refer below steps to get API token>
    url=http://jenkins_server_url:9090
    #This is deprecated, use job_builder section instead
    ignore_cache=True

#### Retrieve API token

Login to the Jenkins dashboard, go to your user page by clicking on your username. Click Configure and then click Show API Token.

#### Step 2: Update(Deploy) Jenkins jobs to jenkins server

    cd hfrd
    jenkins-jobs --conf jjb/jenkins.ini update jjb/hfrd-jenkins-jobs.yaml
If everything is okay , you can find these new jenkins jobs in jenkins dashboard.


## Manage default config files
Sometimes you need some default configurations or some confidential configurations that you don't want to expose them in your test script,then you can add some default config files in jenkins server.Jenkins jobs will read from theses confg files to finish their jobs.

#### Step 1: Install required jenkins plugin
To add config files , you need to install a new jenkins plugin named 'Config File Provider Plugin'.You can search and install the plugin in `http://jenkins_server_url:9090/pluginManager`

#### Step 2: Add new config files
You can find the configFiles in `http://jenkins_server_url:9090/configFiles`.If you want add a new one, click `Add a new Config` in left side.
Totally we have five config files named 'bxproduction' 'bxproduction-ep' 'bxstaging' 'bxstaging-ep' 'cm'.
For example, you can specify the following arguments in  'cm'

    ---
    apikey: <The API Key>
    cmUrl: <The url of cluster manager>
    heliosUrl: <The url of helios>
    heliosSecret: <The helios secret>
    loc: <The location that you want to create networks in>
    plan: <The name of IBP plan >
    hsm: <Need or Don't need the hsm>
    numOfOrgs: <The number of orgs you want to create in network>
    numOfPeers: <The number of peers per org you want to create>
    ledgerType: <The ledger type>


After all of these are done,you are ready to go!