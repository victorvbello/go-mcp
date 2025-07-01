package types

//Information about a validated access token, provided to request handlers.
type AuthInfo struct {
	//The access token.
	Token string
	//When the token expires (in seconds since epoch).
	ExpiresAt int
}
