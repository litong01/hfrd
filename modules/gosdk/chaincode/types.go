package chaincode

// Chaincode configuration yaml file-- chaincodes section
type chaincodes struct {
	Chaincode chaincode      `mapstructure:"chaincode`
}

type chaincode struct {
	Install     install      `mapstructure:"install"`
	Instantiate instantiate  `mapstructure:"instantiate"`
	Invoke      invoke       `mapstructure:"invoke"`
}

type install struct {
	NamePrefix	string	`mapstructure:"namePrefix"`
	Count		int 	`mapstructure: "count"`
	Version 	string	`mapstructure:"version"`
	Path		string	`mapstructure:"path"`
}

type instantiate struct {
	NamePrefix 	string	`mapstructure:"namePrefix"`
	Count		int	`mapstructure: "count"`
	Version		string	`mapstructure:"version"`
	Path		string	`mapstructure:"path"`
	Channel		string	`mapstructure:"channel"`
}

type invoke struct {
	Name		string	 `mapstructure:"name"`
	Count		int	 `mapstructure: "count"`
	Channel		string	 `mapstructure:"channel"`
	Args 		[]string `mapstructure:"args"`
	Threads		int 	 `mapstructure:"threads"`
}
