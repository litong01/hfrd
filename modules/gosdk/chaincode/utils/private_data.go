package utils

import (
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

type CollectionConfig struct {
	Name              string `json:"name"`
	Policy            string `json:"policy"`
	RequiredPeerCount int32  `json:"requiredPeerCount"`
	MaxPeerCount      int32  `json:"maxPeerCount"`
	BlockToLive       uint64 `json:"blockToLive"`
	MemberOnlyRead    bool   `json:"memberOnlyRead"`
}

func NewCollectionConfig(colName, policy string, reqPeerCount, maxPeerCount int32, blockToLive uint64, memberOnlyRead bool) (*cb.CollectionConfig, error) {
	p, err := cauthdsl.FromString(policy)
	if err != nil {
		return nil, err
	}
	cpc := &cb.CollectionPolicyConfig{
		Payload: &cb.CollectionPolicyConfig_SignaturePolicy{
			SignaturePolicy: p,
		},
	}
	return &cb.CollectionConfig{
		Payload: &cb.CollectionConfig_StaticCollectionConfig{
			StaticCollectionConfig: &cb.StaticCollectionConfig{
				Name:              colName,
				MemberOrgsPolicy:  cpc,
				RequiredPeerCount: reqPeerCount,
				MaximumPeerCount:  maxPeerCount,
				BlockToLive:       blockToLive,
			},
		},
	}, nil
}
