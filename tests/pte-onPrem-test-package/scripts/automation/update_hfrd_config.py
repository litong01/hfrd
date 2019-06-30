import requests
import argparse
import sys
import re

parser = argparse.ArgumentParser(description="Python script updates the hfrd_test.cfg file")
parser._action_groups.pop() # removing optional argument title because I want to show required args before optional

#setting up required arguments
requiredArgs = parser.add_argument_group('required arguments')
requiredArgs.add_argument("-a", "--apiuser", metavar='', required=True, help="API user or tester name")
requiredArgs.add_argument("-e", "--env", metavar='', required=True, choices=["bxstaging", "bxproduction", "cm"], help="Test environment")
requiredArgs.add_argument("-p", "--plan", metavar='', required=True, choices=["sp", "ep"], help="Plan type")

#setting up optional arguments
optionalArgs = parser.add_argument_group('optional arguments')
optionalArgs.add_argument("-w", "--workload", metavar='', default="marbles", choices=["samplecc", "marbles"], help="Workload type")
optionalArgs.add_argument("-s", "--serviceId", metavar='', default="{serviceid}", help="Service ID only needed when we want to get the service credentials")
optionalArgs.add_argument("-ct", "--configType", metavar='', default="single", choices=["single", "multiple"], help="Type of network configuration file \n "
                                                                                                                    "multiple not supported yet")
optionalArgs.add_argument('-no', '--numOfOrgs', type=int, metavar='', default=1, help='Number of Orgs')
optionalArgs.add_argument('-np', '--numOfPeers', type=int, metavar='', default=1, help='Number of Peers per org')
optionalArgs.add_argument("-l", "--ledgerType", metavar='', default="levelDB", choices=["levelDB", "couch"],
                          help="Ledger type")

args = parser.parse_args()

AVAILABLE_BXSTAGE_EP_URL = "https://ibmblockchain-dev-v2.stage1.ng.bluemix.net/api/v1/network-locations/available"
AVAILABLE_BXPRODUCTION_EP_URL = "https://ibmblockchain-v2.ng.bluemix.net/api/v1/network-locations/available"

class UpdateConfiguration:
    def __init__(self, tester, config_type, environment, plan, service_id, workload):
        self.config_file = self.get_config_file(config_type)
        self.config_type = config_type
        self.api_user = tester
        self.env = environment
        self.plan = plan
        self.num_of_orgs = args.numOfOrgs
        self.num_of_peers = args.numOfPeers
        self.validate_args()
        self.loc = self.set_location()
        self.service_id = service_id
        self.workload = workload
        self.ledger_type = args.ledgerType

    def get_config_file(self, config_type):
        config_files_dict = {"single": "../../hfrd_test.cfg", "multiple": "../../conf/networks.yaml"}
        return config_files_dict[config_type.lower()]

    def generate_config_file(self):
        # this function generates the configuration file contents
        if args.configType == "single":
            file_content = """\
            # API SERVER DETAILS FOR RUNNING HFRD STARTER PLAN NETWORK TEST
            apiuser={}
            apiserverhost=http://hfrdrestsrv.rtp.raleigh.ibm.com:8080
            apiversion=v1
            apiserverbaseuri="$apiserverhost/$apiversion/$apiuser"
            # choose one from below array indicating:
            # starter, enterprise plan in staging or prod or cm
            # [ bxstaging, bxproduction, cm ]
            env={}
            # Currently 'bxstaging' supports sp and ep. 'bxproduction' only suppports sp. 'cm' only supports ep
            name={}
            # Currently 'cm' can be verified on 'POKM148_ZBC4'. 'bxsaging-ep' can be verified on 'ASH-CI'.
            loc={}
            # Currently 'bxstaging-ep' only supports 1 org and 0~3 peers per org.
            numOfOrgs={}
            numOfPeers={}
            # Specify Vcpus and memory
            vCpusPeer=2
            vCpusOrderer=1
            vCpusKafka=1
            memoryPeer=500
            memoryOrderer=500
            memoryKafka=500
            # ledgerType  : [levelDB couch]
            ledgerType={}
            # HFRD TEST SECTION
            # TEST Running Mode : local or cloud. Default is local
            runMode=local
            # Test Package Server : Used to store test package when runMode=cloud
            packageServerUser=user
            packageServerSecret=secret
            packageServerHost=csl-dev.rtp.raleigh.ibm.com
            testPackageServer=http://$packageServerHost:8081
            packageDir=/home/ibmadmin/Documents/hfrd/
            # Must provide the serviceid if you want to reuse an exsiting network
            serviceid={}
            # Must provide the workload name if you want to run measurements.Currently support 'samplecc' and 'marbles'
            workload={}
            RESULTS_DIR=$(pwd)/results/""".format(self.api_user, self.env, self.plan, self.loc, self.num_of_orgs,\
                                                                              self.num_of_peers, self.ledger_type, self.service_id, self.workload)
            file_content = re.sub(r'(^[ \t]+|[ \t]+(?=:))', '', file_content, flags=re.M) # removing tabs in the beginning
            return file_content

        else:
            sys.exit("{} network config file update automation not support yet".format(args.configType))


    def validate_args(self):
        # validates the parameters being passed''
        if self.env == "cm" and self.plan == "sp":
            sys.exit("{} environment only supports ep plan".format(self.env))

        #ledger_Input = input("Ledger type: ").lower()
        #if ledger_Input == "leveldb":
        #    self.ledger_type = "levelDB"
        #elif ledger_Input == "couch":
        #    self.ledger_type = ledger_Input
        #else:
        #    sys.exit("Ledger type supported (levelDB or couch)")

        #self.num_of_orgs = int(input("Number of Orgs: "))
        #self.num_of_peers = int(input("Number of Peers per Org: "))
        if (self.env == "bxstaging" or self.env == "bxproduction") and self.plan == "ep":
            if self.num_of_orgs > 1:
                sys.exit("{} {} plan only supports 1 num_of_org".format(self.env, self.plan))

            if self.num_of_peers > 3:
                sys.exit("{} {} plan only supports 0-3 peers per org".format(self.env, self.plan))




    def print_configuration(self):
        print("#"*30)
        print("YOUR CONFIGURATION")
        print("ApiUser: {}".format(self.api_user))
        print("Env: {}".format(self.env))
        print("Plan: {}".format(self.plan))
        print("Location: {}".format(self.loc))
        print("Number of Orgs: {}".format(self.num_of_orgs))
        print("Number of Peers per Org: {}".format(self.num_of_peers))
        print("Workload: {}".format(self.workload))
        print("Ledger type: {}".format(self.ledger_type))
        print("Service ID: {}".format(self.service_id))
        print("#"*30)

    def get_available_locations(self, url):
        # This function gets the available location to create a network for bxstaging and bxproduction ep plans
        contents = None
        try:
            r = requests.get(url)
            if r.status_code == 200:
                #html = r.text
                contents = r.json()

        except Exception as ex:
            print(str(ex))

        finally:
            return list(contents.keys())

    def set_location(self):
        # This function sets the location of where you want to create the plan
        location = None
        if self.env == "cm" and self.plan == "ep":
            location = "POKM148_ZBC4"
        elif self.env == "bxstaging" and self.plan == "ep":
            avaliable_locations = self.get_available_locations(AVAILABLE_BXSTAGE_EP_URL)
            print("Available locations in BXSTAGING: {}".format(avaliable_locations))
            if "ASH-PERFORMANCE" in avaliable_locations: # if ZBC09 cluster is available pick set as location else select at random
                location = "ASH-PERFORMANCE" # pick the first available location
            else:
                location = avaliable_locations[0]  # pick the first available location
            print("Setting location to {}".format(location))
        elif self.env == "bxproduction" and self.plan == "ep":
            avaliable_locations = self.get_available_locations(AVAILABLE_BXPRODUCTION_EP_URL)
            print("Available locations in BXPRODUCTION: {}".format(avaliable_locations))
            location = avaliable_locations[0]  # pick the first available location
            print("Setting location to {}".format(location))
        return location

    def update_config_file(self):
        contents = self.generate_config_file()
        try:
            with open(self.config_file, mode='w') as f:
                f.write(contents)

        except IOError:
            sys.exit(self.config_file + " file not found")



def main():
    hfrd = UpdateConfiguration(tester=args.apiuser, config_type=args.configType, environment=args.env, plan=args.plan, \
                               service_id=args.serviceId, workload=args.workload)
    hfrd.update_config_file()
    #hfrd.print_configuration()

if __name__ == "__main__":
    main()