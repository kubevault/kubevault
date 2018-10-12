package types

type AuthLoginResponse struct {
	Auth *Auth `json:"auth"`
}

type Auth struct {
	ClientToken string `json:"client_token"`
}
