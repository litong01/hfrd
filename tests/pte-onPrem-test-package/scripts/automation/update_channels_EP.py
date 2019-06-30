import argparse
import json
import sys

parser = argparse.ArgumentParser(description="Python script updates the channels_EP.json file")
parser._action_groups.pop() # removing optional argument title because I want to show required args before optional

#setting up required arguments
requiredArgs = parser.add_argument_group('required arguments')
requiredArgs.add_argument("-nc", "--numberOfChannels", type=int, metavar='', required=True, help="Number of channels to be created")
requiredArgs.add_argument("-pc", "--peersPerChannel", type=int, metavar='', default=[1], nargs='+', help="Number of peers per channel")

#setting up optional arguments
optionalArgs = parser.add_argument_group('optional arguments')
optionalArgs.add_argument("-sc", "--specialChannels", type=int, metavar='', nargs='+', help="Special Channels")
optionalArgs.add_argument("-sp", "--specialPeersPerChannel", type=int, metavar='', nargs='+', help="Special Peers per channel")

args = parser.parse_args()

class UpdateChannels:
    def __init__(self):
        self.number_channels = args.numberOfChannels
        self.peers_per_channel = args.peersPerChannel
        self.special_channels, self.special_peer_members = self.verify_args()
        self.channel_structure = self.generate_channels()
        self.output = "../../conf/channels_EP.json"
        self.writeToOutput()

    def verify_args(self):
        # this function ensures we are passing special channels cases with the peer members for that case together
        if args.specialChannels:
            if args.specialPeersPerChannel:
                return args.specialChannels, args.specialPeersPerChannel
            else:
                print("You must pass in the --specialPeersPerChannel or -sp argument")
                parser.parse_args(['-h'])
        else:
            return None, None

    def generate_channels(self):



        structure = {}
        structure["channels"] = []

        for number in range(1, self.number_channels + 1):
            temp = {}
            temp["name"] = "channel{}".format(number)
            temp["members"] = []
            temp["batchSize"] = {}
            temp["batchSize"]["messageCount"] = 100
            temp["batchSize"]["absoluteMaxBytes"] = 103809024
            temp["batchSize"]["preferredMaxBytes"] = 103809024
            temp["batchTimeout"] = "10s"
            temp["channelRestrictionMaxCount"] = "150"

            # for special channel configuration
            if self.special_channels:
                if number in self.special_channels:
                    for special_peer in self.special_peer_members:
                        temp["members"].append("PeerOrg{}".format(special_peer))

                    # Add the channel structure to the list
                    structure["channels"].append(temp)
                    continue

            # setting up the peerOrg members
            for peer in self.peers_per_channel:
                temp["members"].append("PeerOrg{}".format(peer))

            # Add the channel structure to the list
            structure["channels"].append(temp)

            # deleting temp for next iteration
            del temp

        return structure

    def writeToOutput(self):
        # this function writes to config-net-${networkID}.json file
        with open(self.output, "w") as f:
            json.dump(self.channel_structure, f, indent=4, sort_keys=True)


def main():
    channels = UpdateChannels()

if __name__ == "__main__":
    main()