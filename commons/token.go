package commons

import "fmt"

type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (t AccessToken) String() string {
	return fmt.Sprintf("access_token='%s', token_type='%s'", t.AccessToken, t.TokenType)
}
