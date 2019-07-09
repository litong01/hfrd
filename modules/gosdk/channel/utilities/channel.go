package utilities

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	"hfrd/modules/gosdk/channel/utilities/configtxlator"
	hfrdcommon "hfrd/modules/gosdk/common"

	"github.com/hyperledger/fabric/protos/common"

	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"strconv"
)

const (
	REPLACE_ORDERER_ADDRESS = "replace"
	ADD_ORDERER_ADDRESS     = "add"
	REMOVE_ORDERER_ADDRESS  = "remove"
)

// CreatChannelTxEnvelope is used to build channel creation envelope
func CreatChannelTxEnvelope(applicationCapability string, channelId string, consortiumName string, mspIds ...string) ([]byte, error) {
	configUpdate := &common.ConfigUpdate{}
	configUpdate.ChannelId = channelId
	readSet := &common.ConfigGroup{}
	readSet.Version = 0
	readSet.Values = map[string]*common.ConfigValue{}
	readSet.Values["Consortium"] = &common.ConfigValue{}
	readSet.Groups = map[string]*common.ConfigGroup{}
	readSet.Groups["Application"] = &common.ConfigGroup{}
	readSet.Groups["Application"].Groups = map[string]*common.ConfigGroup{}
	for _, mspId := range mspIds {
		readSet.Groups["Application"].Groups[mspId] = &common.ConfigGroup{}
	}
	configUpdate.ReadSet = readSet

	writeSet := &common.ConfigGroup{}
	writeSet.Values = map[string]*common.ConfigValue{}
	writeSet.Values["Consortium"] = &common.ConfigValue{}
	consortium := &common.Consortium{}
	consortium.Name = consortiumName
	consortiumBytes, err := proto.Marshal(consortium)
	if err != nil {
		return nil, err
	}
	writeSet.Values["Consortium"].Value = consortiumBytes
	writeSet.Groups = map[string]*common.ConfigGroup{}
	writeSet.Groups["Application"] = &common.ConfigGroup{}
	writeSet.Groups["Application"].Version = 1
	writeSet.Groups["Application"].Groups = map[string]*common.ConfigGroup{}
	for _, mspId := range mspIds {
		writeSet.Groups["Application"].Groups[mspId] = &common.ConfigGroup{}
	}
	writeSet.Groups["Application"].ModPolicy = "Admins"
	writeSet.Groups["Application"].Policies = map[string]*common.ConfigPolicy{}
	// Admins
	writeSet.Groups["Application"].Policies["Admins"] = &common.ConfigPolicy{}
	writeSet.Groups["Application"].Policies["Admins"].ModPolicy = "Admins"
	adminsPolicy, err := makeImplicitMetaPolicy("Admins", common.ImplicitMetaPolicy_MAJORITY)
	if err != nil {
		return nil, err
	}
	writeSet.Groups["Application"].Policies["Admins"].Policy = adminsPolicy
	// Readers
	readersPolicy, err := makeImplicitMetaPolicy("Readers", common.ImplicitMetaPolicy_ANY)
	if err != nil {
		return nil, err
	}
	writeSet.Groups["Application"].Policies["Readers"] = &common.ConfigPolicy{}
	writeSet.Groups["Application"].Policies["Readers"].ModPolicy = "Admins"
	writeSet.Groups["Application"].Policies["Readers"].Policy = readersPolicy
	// Writers
	writersPolicy, err := makeImplicitMetaPolicy("Writers", common.ImplicitMetaPolicy_ANY)
	if err != nil {
		return nil, err
	}
	writeSet.Groups["Application"].Policies["Writers"] = &common.ConfigPolicy{}
	writeSet.Groups["Application"].Policies["Writers"].ModPolicy = "Admins"
	writeSet.Groups["Application"].Policies["Writers"].Policy = writersPolicy

	capabilities := &common.Capabilities{
		Capabilities: make(map[string]*common.Capability),
	}

	// Set Application Capability to V1_2 to enable private data
	if !strings.Contains(applicationCapability, "1_1") {
		capabilities.Capabilities["V1_2"] = &common.Capability{}
	}

	capBytes, err := proto.Marshal(capabilities)
	if err != nil {
		return nil, err
	}
	writeSet.Groups["Application"].Values = make(map[string]*common.ConfigValue)
	writeSet.Groups["Application"].Values["Capabilities"] = &common.ConfigValue{}
	writeSet.Groups["Application"].Values["Capabilities"].ModPolicy = "Admins"
	writeSet.Groups["Application"].Values["Capabilities"].Value = capBytes

	writeSet.Version = 0
	configUpdate.WriteSet = writeSet

	var configUpdateBytes []byte
	if configUpdateBytes, err = proto.Marshal(configUpdate); err != nil {
		return nil, err
	}
	configUpdateEnv := &common.ConfigUpdateEnvelope{
		ConfigUpdate: configUpdateBytes,
	}

	//channel header
	chHeader := &common.ChannelHeader{
		Type:    int32(common.HeaderType_CONFIG_UPDATE),
		Version: int32(0),
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
			Nanos:   0,
		},
		ChannelId: channelId,
		Epoch:     uint64(0),
	}
	var chHeaderBytes []byte
	var payloadBytes []byte
	if chHeaderBytes, err = proto.Marshal(chHeader); err != nil {
		return nil, err
	}
	header := &common.Header{
		ChannelHeader: chHeaderBytes,
	}
	if dataBytes, err := proto.Marshal(configUpdateEnv); err != nil {
		return nil, err
	} else if payloadBytes, err = proto.Marshal(&common.Payload{
		Header: header,
		Data:   dataBytes,
	}); err != nil {
		return nil, err
	} else {
		return proto.Marshal(&common.Envelope{
			Payload: payloadBytes,
		})
	}
}

// QueryChannelConfig is used to query latest channel config block
func QueryChannelConfigBlock(ledgerClient *ledger.Client) ([]byte, error) {
	ledgerInfo, err := ledgerClient.QueryInfo()
	if err != nil {
		hfrdcommon.Logger.Error(fmt.Sprintf("QueryInfo return error: %s", err))
		return nil, err
	}
	currentBlock, err := ledgerClient.QueryBlockByHash(ledgerInfo.BCI.CurrentBlockHash)

	if err != nil {
		hfrdcommon.Logger.Error(fmt.Sprintf("QueryBlockByHash return error: %s", err))
		return nil, err
	}
	if currentBlock.Metadata == nil {
		hfrdcommon.Logger.Error(fmt.Sprintf("QueryBlockByHash block data is nil"))
		return nil, err
	}

	b, err := proto.Marshal(currentBlock)
	if err != nil {
		return nil, err
	}

	lc, err := GetLastConfigIndexFromBlock(b)
	if err != nil {
		hfrdcommon.Logger.Error(fmt.Sprintf("GetLastConfigIndexFromBlock err: %s", err))
		return nil, err
	}

	lastBlock, err := ledgerClient.QueryBlock(lc)
	if err != nil {
		hfrdcommon.Logger.Error(fmt.Sprintf("Query Last Block err: %s", err))
	}
	b, err = proto.Marshal(lastBlock)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// CreateUpdateEnvelope is used to build channel update envelope
func CreateUpdateEnvelope(updateOptions *UpdateOptions, configBlockBytes []byte) ([]byte, *UpdateOptions, error) {
	// Decode channel block
	originalChannelGroup, err := configtxlator.DecodeProto("common.Block", configBlockBytes)
	if err != nil {
		return nil, updateOptions, errors.Errorf("Error in decode proto : %s", err)
	}
	oriBuffer := EncodeAndReplaceNull(originalChannelGroup)

	// Get the new channel group based on the update options
	var updatedChannelGroup = make(map[string]interface{})
	updatedChannelGroup, updateOptions, err = UpdateChannelGroup(originalChannelGroup, updateOptions)
	if err != nil {
		return nil, updateOptions, err
	}
	if !updateOptions.OrdererOrgUpdate && !updateOptions.OrdererAddressesUpdate && !updateOptions.PeerOrgUpdate {
		return nil, updateOptions, errors.Errorf("The new config are exactly the same as current config")
	}
	updatedBuffer := EncodeAndReplaceNull(updatedChannelGroup)

	// Encode and compute update proto bytes
	originalProtoBytes, err := configtxlator.EncodeProto("common.Config", oriBuffer.Bytes())
	if err != nil {
		return nil, updateOptions, errors.Errorf("error in encode json to proto message : %s", err)
	}
	updatedProtoBytes, err := configtxlator.EncodeProto("common.Config", updatedBuffer.Bytes())
	if err != nil {
		return nil, updateOptions, errors.Errorf("error in encode json to proto message: %s", err)
	}
	computeUpdateBytes, err := configtxlator.ComputeUpdt(originalProtoBytes, updatedProtoBytes, updateOptions.ChannelID)
	configUpdate, err := configtxlator.DecodeProto("common.ConfigUpdate", computeUpdateBytes)
	if err != nil {
		return nil, updateOptions, errors.Errorf("Error in decode proto : %s", err)
	}

	// Build config update envelope
	configUpdateEnvelope := &ConfigUpdate{}
	configUpdateEnvelope.Payload.Data.Config_update = configUpdate
	configUpdateEnvelope.Payload.Header.Channel_header.Type = "2"
	configUpdateEnvelope.Payload.Header.Channel_header.Channel_id = updateOptions.ChannelID

	// Replace and lowercase some strings
	var cueBuffer bytes.Buffer
	encoderUpdated := json.NewEncoder(&cueBuffer)
	encoderUpdated.SetIndent("", "\t")
	encoderUpdated.Encode(configUpdateEnvelope)
	cuestring := string(cueBuffer.Bytes())
	r := regexp.MustCompile("\"value\": null,\n")
	cuestring = r.ReplaceAllString(cuestring, "")
	r = regexp.MustCompile("\"policy\": null,\n")
	cuestring = r.ReplaceAllString(cuestring, "")
	replaceKeys := []string{"Payload", "Header", "Channel_header", "Data", "Config_update", "Type", "Channel_id"}
	cuestring = lowercaseStringWithKeys(cuestring, replaceKeys)

	cueProtoBytes, err := configtxlator.EncodeProto("common.Envelope", []byte(cuestring))
	if cueProtoBytes == nil {
		return nil, updateOptions, errors.Errorf("Error in update channel : no update envelope provided")
	}

	return cueProtoBytes, updateOptions, nil

}

func UpdateChannelGroup(originalGroup map[string]interface{}, updateOptions *UpdateOptions) (map[string]interface{}, *UpdateOptions, error) {
	var updatedChannelGroup = make(map[string]interface{})
	updatedChannelGroup = originalGroup

	batchTimeout, maxMessageCount, preferredMaxBytes, anchorPeerConfig, ordererCurrAddresses := GetConfigFromChannelGroup(updatedChannelGroup, updateOptions.MspID)
	if maxMessageCount != updateOptions.MaxMessageCount && updateOptions.MaxMessageCount != 0 {
		updatedChannelGroup["channel_group"].(map[string]interface{})["groups"].(map[string]interface{})["Orderer"].(map[string]interface{})["values"].(map[string]interface{})["BatchSize"].(map[string]interface{})["value"].(map[string]interface{})["max_message_count"] = updateOptions.MaxMessageCount
		updateOptions.OrdererOrgUpdate = true
	}
	if preferredMaxBytes != updateOptions.PreferredMaxBytes && updateOptions.PreferredMaxBytes != 0 {
		updatedChannelGroup["channel_group"].(map[string]interface{})["groups"].(map[string]interface{})["Orderer"].(map[string]interface{})["values"].(map[string]interface{})["BatchSize"].(map[string]interface{})["value"].(map[string]interface{})["preferred_max_bytes"] = updateOptions.PreferredMaxBytes
		updateOptions.OrdererOrgUpdate = true
	}

	if batchTimeout != updateOptions.BatchTimeout && updateOptions.BatchTimeout != "" {
		updatedChannelGroup["channel_group"].(map[string]interface{})["groups"].(map[string]interface{})["Orderer"].(map[string]interface{})["values"].(map[string]interface{})["BatchTimeout"].(map[string]interface{})["value"].(map[string]interface{})["timeout"] = updateOptions.BatchTimeout
		updateOptions.OrdererOrgUpdate = true
	}

	if len(updateOptions.OrdererAddresses) != 0 {
		ordererCurrBytes, err := json.Marshal(ordererCurrAddresses)
		if err != nil {
			return nil, nil, err
		}
		var ordererAddresses []string
		switch updateOptions.OrdererAddressesAction {
		case REPLACE_ORDERER_ADDRESS:
			ordererAddresses = updateOptions.OrdererAddresses
		case ADD_ORDERER_ADDRESS:
			for _, ordererAddress := range updateOptions.OrdererAddresses {
				if contains(ordererCurrAddresses, ordererAddress) == -1 {
					ordererCurrAddresses = append(ordererCurrAddresses, ordererAddress)
				}
			}
			ordererAddresses = ordererCurrAddresses
		case REMOVE_ORDERER_ADDRESS:
			for _, ordererAddress := range updateOptions.OrdererAddresses {
				for {
					if index := contains(ordererCurrAddresses, ordererAddress); index != -1 {
						ordererCurrAddresses = append(ordererCurrAddresses[:index], ordererCurrAddresses[index+1:]...)
					} else {
						break
					}
				}
			}
			ordererAddresses = ordererCurrAddresses
		}
		addressesValue := make(map[string][]string)
		addressesValue["addresses"] = updateOptions.OrdererAddresses
		ordererNewBytes, err := json.Marshal(ordererAddresses)
		if err != nil {
			return nil, nil, err
		}
		if string(ordererCurrBytes) != string(ordererNewBytes) {
			updatedChannelGroup["channel_group"].(map[string]interface{})["values"].(map[string]interface{})["OrdererAddresses"].(map[string]interface{})["value"].(map[string]interface{})["addresses"] = ordererAddresses
			updateOptions.OrdererAddressesUpdate = true
		}
	}

	if len(updateOptions.AnchorPeers) != 0 {
		var anchorPeers []map[string]interface{}
		for _, anchorPeer := range updateOptions.AnchorPeers {
			anchorConfig := make(map[string]interface{})
			if len(strings.Split(anchorPeer, ":")) != 2 {
				return nil, updateOptions, errors.Errorf("Please double check your anchor peers configurations.They must include both ip and port")
			}
			anchorConfig["host"] = strings.Split(anchorPeer, ":")[0]
			port, err := strconv.ParseFloat(strings.Split(anchorPeer, ":")[1], 32)
			if err != nil {
				return nil, nil, err
			}
			anchorConfig["port"] = port
			anchorPeers = append(anchorPeers, anchorConfig)
		}
		newAnchorConfig := make(map[string]interface{})
		newAnchorConfig["mod_policy"] = "Admins"
		anchorValue := make(map[string][]map[string]interface{})
		anchorValue["anchor_peers"] = anchorPeers
		newAnchorConfig["value"] = anchorValue
		if anchorPeerConfig == nil {
			updatedChannelGroup["channel_group"].(map[string]interface{})["groups"].(map[string]interface{})["Application"].(map[string]interface{})["groups"].(map[string]interface{})[updateOptions.MspID].(map[string]interface{})["values"].(map[string]interface{})["AnchorPeers"] = newAnchorConfig
			updateOptions.PeerOrgUpdate = true
		} else {
			anchorCurrBytes, err := json.Marshal(anchorPeerConfig["value"].(map[string]interface{})["anchor_peers"])
			if err != nil {
				return nil, nil, err
			}
			anchorNewBytes, err := json.Marshal(anchorPeers)
			if err != nil {
				return nil, nil, err
			}
			if string(anchorCurrBytes) != string(anchorNewBytes) {
				updatedChannelGroup["channel_group"].(map[string]interface{})["groups"].(map[string]interface{})["Application"].(map[string]interface{})["groups"].(map[string]interface{})[updateOptions.MspID].(map[string]interface{})["values"].(map[string]interface{})["AnchorPeers"].(map[string]interface{})["value"].(map[string]interface{})["anchor_peers"] = anchorPeers
				updateOptions.PeerOrgUpdate = true
			}
		}
	}

	configOut := make(map[string]interface{})
	configOut["channel_group"] = updatedChannelGroup["channel_group"]
	return configOut, updateOptions, nil
}

// GetConfigFromChannelGroup will get batchTimeout/batchSize/AnchorPeers from channelGroup
func GetConfigFromChannelGroup(channelGroup map[string]interface{}, mspID string) (batchTimeout string, maxMessageCount float64, preferredMaxBytes float64, anchorPeers map[string]interface{}, orderers []string) {

	maxMessageCount = channelGroup["channel_group"].(map[string]interface{})["groups"].(map[string]interface{})["Orderer"].(map[string]interface{})["values"].(map[string]interface{})["BatchSize"].(map[string]interface{})["value"].(map[string]interface{})["max_message_count"].(float64)

	preferredMaxBytes = channelGroup["channel_group"].(map[string]interface{})["groups"].(map[string]interface{})["Orderer"].(map[string]interface{})["values"].(map[string]interface{})["BatchSize"].(map[string]interface{})["value"].(map[string]interface{})["preferred_max_bytes"].(float64)

	batchTimeout = channelGroup["channel_group"].(map[string]interface{})["groups"].(map[string]interface{})["Orderer"].(map[string]interface{})["values"].(map[string]interface{})["BatchTimeout"].(map[string]interface{})["value"].(map[string]interface{})["timeout"].(string)

	mspValues := channelGroup["channel_group"].(map[string]interface{})["groups"].(map[string]interface{})["Application"].(map[string]interface{})["groups"].(map[string]interface{})[mspID].(map[string]interface{})["values"].(map[string]interface{})
	var orgAnchors map[string]interface{}
	if _, ok := mspValues["AnchorPeers"]; ok {
		orgAnchors = mspValues["AnchorPeers"].(map[string]interface{})
	}
	ordererAddressesValue := channelGroup["channel_group"].(map[string]interface{})["values"].(map[string]interface{})["OrdererAddresses"].(map[string]interface{})["value"].(map[string]interface{})
	var ordererAddresses []string
	if _, ok := ordererAddressesValue["addresses"]; ok {
		for _, address := range ordererAddressesValue["addresses"].([]interface{}) {
			ordererAddresses = append(ordererAddresses, address.(string))
		}

	}
	return batchTimeout, maxMessageCount, preferredMaxBytes, orgAnchors, ordererAddresses
}

// PrintMSPidFromChannelGroup print MSPid list in given channel group config
func PrintMSPidFromChannelGroup(channelGroup map[string]interface{}) {

	orgMspFromChannelConfig := channelGroup["channel_group"].(map[string]interface{})["groups"].(map[string]interface{})["Application"].(map[string]interface{})["groups"].(map[string]interface{})
	var orgMSPId []string
	for key := range orgMspFromChannelConfig {
		orgMSPId = append(orgMSPId, key)
	}
	hfrdcommon.Logger.Info(fmt.Sprintf("org MSPID in channel config: %s", orgMSPId))
}

// GetLastConfigIndexFromBlock retrieves the index of the last config block as encoded in the block metadata
func GetLastConfigIndexFromBlock(blockBytes []byte) (uint64, error) {
	block := &common.Block{}
	proto.Unmarshal(blockBytes, block)
	md, err := GetMetadataFromBlock(block, common.BlockMetadataIndex_LAST_CONFIG)
	if err != nil {
		return 0, err
	}
	lc := &common.LastConfig{}
	err = proto.Unmarshal(md.Value, lc)
	if err != nil {
		return 0, err
	}
	return lc.Index, nil
}

// GetMetadataFromBlock retrieves metadata at the specified index.
func GetMetadataFromBlock(block *common.Block, index common.BlockMetadataIndex) (*common.Metadata, error) {
	md := &common.Metadata{}
	err := proto.Unmarshal(block.Metadata.Metadata[index], md)
	if err != nil {
		return nil, err
	}
	return md, nil
}

func makeImplicitMetaPolicy(subPolicyName string, rule common.ImplicitMetaPolicy_Rule) (*common.Policy, error) {
	valueBytes, err := proto.Marshal(&common.ImplicitMetaPolicy{
		Rule:      rule,
		SubPolicy: subPolicyName,
	})
	if err != nil {
		return nil, err
	}
	return &common.Policy{
		Type:  int32(common.Policy_IMPLICIT_META),
		Value: valueBytes,
	}, nil
}

func UpdateChannel(sdk *fabsdk.FabricSDK, configuUpdateEnvelope []byte, updateOptions *UpdateOptions) error {
	var mspClient *mspclient.Client
	var signingIdentities []msp.SigningIdentity
	var err error
	if updateOptions.OrdererOrgUpdate || updateOptions.OrdererAddressesUpdate {
		mspClient, err = mspclient.New(sdk.Context(), mspclient.WithOrg(updateOptions.OrdererOrgName))
		if err != nil {
			hfrdcommon.Logger.Error(fmt.Sprintf("Error creating msp client: %s", err))
			return err
		}
		adminIdentity, err := mspClient.GetSigningIdentity(hfrdcommon.ADMIN)
		if err != nil {
			hfrdcommon.Logger.Error(fmt.Sprintf("failed to get admin signing identity: %s", err))
			return err
		}
		signingIdentities = append(signingIdentities, adminIdentity)
	}
	if updateOptions.PeerOrgUpdate {
		mspClient, err = mspclient.New(sdk.Context(), mspclient.WithOrg(updateOptions.LedgerClientOrg))
		if err != nil {
			hfrdcommon.Logger.Error(fmt.Sprintf("Error creating msp client: %s", err))
			return err
		}
		adminIdentity, err := mspClient.GetSigningIdentity(hfrdcommon.ADMIN)
		if err != nil {
			hfrdcommon.Logger.Error(fmt.Sprintf("failed to get admin signing identity: %s", err))
			return err
		}
		signingIdentities = append(signingIdentities, adminIdentity)
	}

	orgClientContext := sdk.Context(fabsdk.WithUser(hfrdcommon.ADMIN), fabsdk.WithOrg(updateOptions.LedgerClientOrg))
	orgClient, err := resmgmt.New(orgClientContext)
	if err != nil {
		hfrdcommon.Logger.Error(fmt.Sprintf("Error creating resource manager client: %s", err))
		return err
	}

	req := resmgmt.SaveChannelRequest{ChannelID: updateOptions.ChannelID, ChannelConfig: bytes.NewReader(configuUpdateEnvelope),
		SigningIdentities: signingIdentities}

	res, err := orgClient.SaveChannel(req, resmgmt.WithOrdererEndpoint(updateOptions.OrdererName))
	if err != nil || res.TransactionID == "" {
		if strings.Contains(err.Error(), "error validating ReadSet") {
			hfrdcommon.Logger.Error(fmt.Sprintf("The channel:%s might already exist, skip the error: %s", updateOptions.ChannelID, err))
			return err
		} else {
			hfrdcommon.Logger.Error(fmt.Sprintf("failed to create channel with name %s: %s", updateOptions.ChannelID, err))
			return err
		}
	}
	return nil
}

func WaitUntilUpdateSucc(ledgerClient *ledger.Client, updateOptions *UpdateOptions) error {
	var waitCount int
	for {
		waitCount++
		configBlockBytes, err := QueryChannelConfigBlock(ledgerClient)
		if err != nil {
			return err
		}

		currentChannelGroup, err := configtxlator.DecodeProto("common.Block", configBlockBytes)
		if err != nil {
			return errors.Errorf("Error in decode proto : %s", err)
		}

		batchTimeout, maxMessageCount, preferredMaxBytes, anchorPeerConfig, ordererAddresses := GetConfigFromChannelGroup(currentChannelGroup, updateOptions.MspID)
		if updateOptions.OrdererOrgUpdate {
			if updateOptions.BatchTimeout != "" && updateOptions.BatchTimeout != batchTimeout {
				continue
			}
			if updateOptions.MaxMessageCount != 0 && updateOptions.MaxMessageCount != maxMessageCount {
				continue
			}
			if updateOptions.PreferredMaxBytes != 0 && updateOptions.PreferredMaxBytes != preferredMaxBytes {
				continue
			}
		}
		// Check if orderer addresses are updated successfully.Comment out for now.
		//if updateOptions.OrdererAddressesUpdate {
		//	fmt.Printf("Current orderer addresses: %s \n", ordererAddresses)
		//}
		if updateOptions.PeerOrgUpdate {
			if anchorPeerConfig == nil {
				continue
			}
		}
		hfrdcommon.Logger.Info(fmt.Sprintf("Successfully updated channel: %s", updateOptions.ChannelID))
		hfrdcommon.Logger.Info(fmt.Sprintf("    Batchtimeout: %s", batchTimeout))
		hfrdcommon.Logger.Info(fmt.Sprintf("    BatchSize.MaxMessageCount: %f", maxMessageCount))
		hfrdcommon.Logger.Info(fmt.Sprintf("    BatchSize.PreferredMaxBytes: %f", preferredMaxBytes))
		hfrdcommon.Logger.Info(fmt.Sprintf("    OrdererAddresses: %s", ordererAddresses))
		hfrdcommon.Logger.Info(fmt.Sprintf("    Org anchor peers: %#v", anchorPeerConfig))
		break
	}
	return nil
}

func EncodeAndReplaceNull(mapJson map[string]interface{}) *bytes.Buffer {
	var updatedBuffer bytes.Buffer
	encoderUpdated := json.NewEncoder(&updatedBuffer)
	encoderUpdated.SetIndent("", "\t")
	encoderUpdated.Encode(mapJson)

	r := regexp.MustCompile("\"signing_identity\": null,\n")
	out := r.ReplaceAllString(string(updatedBuffer.Bytes()), "")
	r = regexp.MustCompile("\"fabricNodeOus\": null,\n")
	out = r.ReplaceAllString(out, "")
	r = regexp.MustCompile("\"value\": null,\n")
	out = r.ReplaceAllString(out, "")
	r = regexp.MustCompile("\"policy\": null,\n")
	out = r.ReplaceAllString(out, "")
	r = regexp.MustCompile("\"crypto_config\": null,\n")
	out = r.ReplaceAllString(out, "")

	return bytes.NewBuffer([]byte(out))
}

func lowercaseStringWithKeys(original string, keys []string) string {
	if len(keys) == 0 {
		return original
	}
	for _, key := range keys {
		keyLowercase := strings.ToLower(key)
		original = strings.Replace(original, key, keyLowercase, -1)
	}
	return original
}

// CreateNewOrgEnvelope used to build add new org channel update envelope
func CreateNewOrgEnvelope(origChannelGroup map[string]interface{}, modifiedChannelGroup map[string]interface{}, channelID string) ([]byte, error) {
	//Encode channel group to buffer
	oriBuffer := EncodeAndReplaceNull(origChannelGroup)
	updatedBuffer := EncodeAndReplaceNull(modifiedChannelGroup)

	//Encode and compute update proto bytes
	originalProtoBytes, err := configtxlator.EncodeProto("common.Config", oriBuffer.Bytes())
	if err != nil {
		return nil, errors.Errorf("error in encode json to proto message : %s \n", err)
	}
	//for debug
	//_ = originalProtoBytes
	updatedProtoBytes, err := configtxlator.EncodeProto("common.Config", updatedBuffer.Bytes())
	if err != nil {
		return nil, errors.Errorf("error in encode json to proto message: %s \n", err)
	}

	computeUpdateBytes, err := configtxlator.ComputeUpdt(originalProtoBytes, updatedProtoBytes, channelID)
	if err != nil {
		return nil, errors.Errorf("Error in configtxlator.ComputeUpdt : %s \n", err)
	}

	configUpdate, err := configtxlator.DecodeProto("common.ConfigUpdate", computeUpdateBytes)
	if err != nil {
		return nil, errors.Errorf("Error in decode proto : %s \n", err)
	}

	// Build config update envelope
	configUpdateEnvelope := &ConfigUpdate{}
	configUpdateEnvelope.Payload.Data.Config_update = configUpdate
	configUpdateEnvelope.Payload.Header.Channel_header.Type = "2"
	configUpdateEnvelope.Payload.Header.Channel_header.Channel_id = channelID

	// Replace and lowercase some strings
	var cueBuffer bytes.Buffer
	encoderUpdated := json.NewEncoder(&cueBuffer)
	encoderUpdated.SetIndent("", "\t")
	encoderUpdated.Encode(configUpdateEnvelope)
	cuestring := string(cueBuffer.Bytes())
	r := regexp.MustCompile("\"value\": null,\n")
	cuestring = r.ReplaceAllString(cuestring, "")
	r = regexp.MustCompile("\"policy\": null,\n")
	cuestring = r.ReplaceAllString(cuestring, "")
	r = regexp.MustCompile("\"signing_identity\": null,\n")
	cuestring = r.ReplaceAllString(cuestring, "")
	replaceKeys := []string{"Payload", "Header", "Channel_header", "Data", "Config_update", "Type", "Channel_id"}
	cuestring = lowercaseStringWithKeys(cuestring, replaceKeys)
	//return oriBuffer.Bytes(),err
	cueProtoBytes, err := configtxlator.EncodeProto("common.Envelope", []byte(cuestring))
	if cueProtoBytes == nil {
		return nil, errors.Errorf("Error in update channel : no update envelope provided")
	}
	return cueProtoBytes, nil
}

//CollectSign collect SigningIdentity for provided members
func CollectSign(sdk *fabsdk.FabricSDK, requiredMember []string) ([]msp.SigningIdentity, error) {
	var mspClient *mspclient.Client
	var signingIdentities []msp.SigningIdentity
	var err error
	for _, peerOrg := range requiredMember {
		mspClient, err = mspclient.New(sdk.Context(), mspclient.WithOrg(peerOrg))
		if err != nil {
			hfrdcommon.Logger.Error(fmt.Sprintf("Error creating msp client: %s", err))
			return nil, err
		}
		adminIdentity, err := mspClient.GetSigningIdentity(hfrdcommon.ADMIN)
		if err != nil {
			hfrdcommon.Logger.Error(fmt.Sprintf("failed to get admin signing identity: %s", err))
			return nil, err
		}
		signingIdentities = append(signingIdentities, adminIdentity)
	}
	return signingIdentities, err
}

//AddOrgToChannel submit save channel request with provided channelID, transaction envelope and required signature to orderer name
func AddOrgToChannel(sdk *fabsdk.FabricSDK, channelID string, addOrgEnvelopByte []byte, signIdentities []msp.SigningIdentity, ledgerClientOrg string, ordererName string) error {
	orgClientContext := sdk.Context(fabsdk.WithUser(hfrdcommon.ADMIN), fabsdk.WithOrg(ledgerClientOrg))
	orgClient, err := resmgmt.New(orgClientContext)
	if err != nil {
		hfrdcommon.Logger.Error(fmt.Sprintf("Error creating resource manager client: %s", err))
		return err
	}

	req := resmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfig: bytes.NewReader(addOrgEnvelopByte),
		SigningIdentities: signIdentities}

	res, err := orgClient.SaveChannel(req, resmgmt.WithOrdererEndpoint(ordererName))
	if err != nil || res.TransactionID == "" {
		hfrdcommon.Logger.Error(fmt.Sprintf("failed to add new org to channel with name %s: %s", channelID, err))
		return err
	}
	hfrdcommon.Logger.Info(fmt.Sprintf("Successfully add new org to channel: %s, transactionID is %s", channelID, res.TransactionID))
	return nil
}

func contains(s []string, e string) int {
	for i, a := range s {
		if a == e {
			return i
		}
	}
	return -1
}
