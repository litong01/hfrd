package jwt

type IbmJwt struct {
	IbmHeader
	IbmPayload
	RawString string
}
type IbmHeader struct {
	Kid string `json:"kid"`	// key type
	Alg string `json:"alg"`	// algorithm for the key
}

type IbmPayload struct {
	IamId string `json:"iam_id"`
	Id string `json:"id"`
	Realmid string `json:"realmid"`
	Identifier string `json:"identifier"`
	GivenName string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Name string `json:"name"`
	Email string `json:"email"`
	Sub string `json:"sub"`
	Account account `json:"account"`
	Iat int64 `json:"iat"`
	Exp int64 `json:"exp"`
	Iss string `json:"iss"`
	GrantType string `json:"grant_type"`
	Scope string `json:"scope"`
	ClientId string `json:"client_id"`
	Acr int64 `json"acr"`
	Amr []string `json:"amr"`
}

type account struct {
	Bss string `json:"bss"`
}
