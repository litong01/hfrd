package resource

// Plan struct holds metadata of a blockchain network,
// like environment(bxstaging, bxproduction, cm and more to come)
// Testers should be able to choose where the blockchain network
// to be created, like "ZBC01 in POK" for performance testers
type Plan struct {
	Env      string        `json:"env"`    // bxstaging, bxproduction, cm, etc...
	Location string        `json:"loc"`    // Only required for Enterprise Plan
	Name     string        `json:"name"`   // Starter or Enterprise plan, [sp, ep]
	Config   NetworkConfig `json:"config"` // Customized network configuration. Only required for cm environment
}

type NetworkConfig struct {
	NumOfOrgs  int    `json:"numOfOrgs"`
	NumOfPeers int    `json: "numOfPeers`
	LedgerType string `json: "ledgerType"`
}

type Test struct {
	Url      string `json:"url"`
	Hash     string `json:"hash"`
	Startcmd string `json:"startcmd"`
	Sslcerts bool   `json:"sslcerts"`
}

// a subset of couch.Network
type Service struct {
	NetworkId  string `json:"networkId"`
	ServiceId  string `json:"serviceId"`
	PlanName   string `json:"planName"`
	Env        string `json:"env"`
	CreatedAt  string `json:"createdAt"` // Human readable format
	JobName    string `json:"jobName"`
	JobId      string `json:"jobId"`
	Location   string `json:"location,omitempty"`   // location id of IBP Enterprise Plan
	NumOfOrgs  int    `json:"numOfOrgs,omitempty"`  // number of organizations-- IBP EP
	NumOfPeers int    `json:"numOfPeers,omitempty"` // number of peers per organization-- IBP EP
	LedgerType string `json:"ledgerType,omitempty"` // ledger type: couch/levelDB-- IBP EP
}
