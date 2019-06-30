package common

// ConnectionProfile shows
type ConnectionProfile struct {
	Client        Client                  `mapstructure:"client"`
	Orderers      map[string]Orderer      `mapstructure:"orderers"`
	Organizations map[string]Organization `mapstructure:"organizations"`
	Peers         map[string]Peer         `mapstructure:"peers"`
	Channels      map[string]Channel      `mapstructure:"channels"`
}

// Client shows
type Client struct {
	Organization string       `string:"organization"`
	Cryptoconfig Cryptoconfig `string:"cryptoconfig"`
}

type Cryptoconfig struct {
	Path string `string:"Path"`
}

// Organization shows
type Organizations struct {
	Organizations map[string]Organization `mapstructure:"organizations"`
}

// Organization shows
type Organization struct {
	Mspid                  string   `mapstructure:"mspid"`
	CryptoPath             string   `yaml:"cryptoPath,omitempty"`
	Peers                  []string `yaml:"peers,omitempty"`
	CertificateAuthorities []string `yaml:"certificateAuthorities,omitempty"`
}

type Peer struct {
	URL string `string:"url"`
}

// Orderer shows
type Orderer struct {
	URL string `string:"url"`
}
type Channels struct {
	Channels map[string]Channel `mapstructure:"channels"`
}

type Channel struct {
	Peers map[string]PeerChannel `mapstructure:"peers"`
}

type PeerChannel struct {
	EndorsingPeer  bool `mapstructure:"endorsingPeer"`
	ChaincodeQuery bool `mapstructure:"chaincodeQuery"`
	LedgerQuery    bool `mapstructure:"ledgerQuery"`
	EventSource    bool `mapstructure:"eventSource"`
}
