import json
import os
import argparse

HOME = os.environ['HOME']+'/results/'

parser = argparse.ArgumentParser(description="Python script generates the SCFiles using MSPIDs")
parser.add_argument("-m", "--mspids", nargs="+", required=True, help="1 or more MSPIDs")
parser.add_argument("-n", "--networkId", metavar='', required=True, help="Network ID")

args = parser.parse_args()


class SCFileCreator:
    def __init__(self):
        self.MSPIDs = args.mspids
        self.peerInfo = {}
        self.SCFileObject = {}
        self.networkID = args.networkId
        self.writeToOutput(self.networkID)

    # Get information for each peerOrgs
    def getPeerInfo(self):
        # This function gets the peer information for all the peers and returns data in a dictionary format
        for mspid in self.MSPIDs:
            # read in connection profiles for each org
            with open(os.path.join(HOME, "creds/ConnectionProfile_{}.json".format(mspid)), "r") as f:
                orgCP = json.load(f)
            # read in admin cert for each org
            with open(os.path.join(HOME, "creds/{}admin/msp/signcerts/cert.pem".format(mspid)), "r") as f:
                orgCT = "".join(f.readlines())
            # read in priv key for each org
            with open(os.path.join(HOME, "creds/{}admin/msp/keystore/priv.pem".format(mspid)), "r") as f:
                orgPK = "".join(f.readlines())
            temp = {}
            temp["orgCP"] = orgCP
            temp["orgCT"] = orgCT
            temp["orgPK"] = orgPK

            self.peerInfo[mspid] = dict(temp)

        return self.peerInfo

    def generateSCFile(self):
        # This function builds the SCFile

        self.getPeerInfo()  # Calling the gatherPeerOrg function

        self.SCFileObject["test-network"] = {}
        print(self.MSPIDs)
        # Added GOPATH as per Tanya"s request
        self.SCFileObject["test-network"]["gopath"] = "GOPATH"

        for mspid in self.MSPIDs:
            # Need to make copy of all inner dict to a new address location without sharing the same reference as the first one
            self.SCFileObject["test-network"]["orderer"] = {}
            self.SCFileObject["test-network"][mspid] = {}
            self.SCFileObject["test-network"][mspid]["ca"] = {}
            self.SCFileObject["test-network"][mspid]["name"] = mspid
            self.SCFileObject["test-network"][mspid]["mspid"] = mspid
            self.SCFileObject["test-network"][mspid]["username"] = "admin"
            self.SCFileObject["test-network"][mspid]["privateKeyPEM"] = ""
            self.SCFileObject["test-network"][mspid]["signedCertPEM"] = ""
            self.SCFileObject["test-network"][mspid]["adminPath"] = ""

            # Storing certificate and private key
            self.SCFileObject["test-network"][mspid]["admin_cert"] = self.peerInfo[mspid]["orgCT"]
            self.SCFileObject["test-network"][mspid]["priv"] = self.peerInfo[mspid]["orgPK"]

            # getting all fabric_ca in peer org
            fabricCaPeerList = [fabric_ca for fabric_ca in
                                self.peerInfo[mspid]["orgCP"]["certificateAuthorities"].keys()]

            # storing the first fabric_ca since the data is the same for each peer org
            self.SCFileObject["test-network"][mspid]["ca"]["name"] = fabricCaPeerList[0]
            self.SCFileObject["test-network"][mspid]["ca"]["url"] = \
            self.peerInfo[mspid]["orgCP"]["certificateAuthorities"][fabricCaPeerList[0]]["url"]
            self.SCFileObject["test-network"][mspid]["secret"] = \
            self.peerInfo[mspid]["orgCP"]["certificateAuthorities"][fabricCaPeerList[0]]["registrar"][0]["enrollSecret"]

            # getting the right peer orgs
            for peer in self.peerInfo[mspid]["orgCP"]["organizations"][mspid]["peers"]:
                # building peer dict
                self.SCFileObject["test-network"][mspid][peer] = {}
                self.SCFileObject["test-network"][mspid][peer]["server-hostname"] = None
                self.SCFileObject["test-network"][mspid][peer]["tls_cacerts"] = ""
                self.SCFileObject["test-network"][mspid][peer]["requests"] = \
                self.peerInfo[mspid]["orgCP"]["peers"][peer]["url"]
                self.SCFileObject["test-network"][mspid][peer]["events"] = self.peerInfo[mspid]["orgCP"]["peers"][peer][
                    "eventUrl"]

            # getting data for each orderer
            for fabricOrderer in self.peerInfo[mspid]["orgCP"]["orderers"]:
                self.SCFileObject["test-network"]["tls_cert"] = \
                self.peerInfo[mspid]["orgCP"]["orderers"][fabricOrderer]["tlsCACerts"]["pem"]

                # building orderer dict
                self.SCFileObject["test-network"]["orderer"][fabricOrderer] = {}
                self.SCFileObject["test-network"]["orderer"][fabricOrderer]["name"] = "OrdererOrg"
                self.SCFileObject["test-network"]["orderer"][fabricOrderer]["mspid"] = "OrdererOrg"
                self.SCFileObject["test-network"]["orderer"][fabricOrderer]["mspPath"] = ""
                self.SCFileObject["test-network"]["orderer"][fabricOrderer]["adminPath"] = ""
                self.SCFileObject["test-network"]["orderer"][fabricOrderer]["comName"] = ""
                self.SCFileObject["test-network"]["orderer"][fabricOrderer]["server-hostname"] = None
                self.SCFileObject["test-network"]["orderer"][fabricOrderer]["tls_cacerts"] = ""
                self.SCFileObject["test-network"]["orderer"][fabricOrderer]["url"] = \
                self.peerInfo[mspid]["orgCP"]["orderers"][fabricOrderer]["url"]

                # setting the ordererID for each mspid
                self.SCFileObject["test-network"][mspid]["ordererID"] = fabricOrderer

        return self.SCFileObject

    def writeToOutput(self, outputFile):
        # this function writes to config-net-${networkID}.json file
        with open(os.path.join(HOME, "SCFiles/config-net-{}.json".format(outputFile)), "w") as f:
            json.dump(self.generateSCFile(), f, indent=4, sort_keys=True)


if __name__ == "__main__":
    scFileCreator = SCFileCreator()