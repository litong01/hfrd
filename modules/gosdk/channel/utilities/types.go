package utilities

type ConfigUpdate struct {
	Payload Payload
}

type Payload struct {
	Header Header
	Data   Data
}

type Header struct {
	Channel_header Channel_header
}

type Channel_header struct {
	Channel_id string
	Type       string
}

type Data struct {
	Config_update map[string]interface{}
}

type UpdateOptions struct {
	ChannelID              string
	LedgerClientOrg        string
	MspID                  string
	OrdererOrgUpdate       bool // The config type : Orderer_Related or Peer_related
	OrdererAddressesUpdate bool
	PeerOrgUpdate          bool
	OrdererOrgName         string
	OrdererName            string
	BatchTimeout           string
	MaxMessageCount        float64
	PreferredMaxBytes       float64
	OrdererAddressesAction string
	OrdererAddresses       []string
	AnchorPeers            []string
}
