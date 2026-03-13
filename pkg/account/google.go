package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const GoogleAuthTokenUrl = "https://oauth2.googleapis.com/tokeninfo"

type GoogleTokenPayload struct {
	Iss           string `json:"iss"`
	Azp           string `json:"azp"`
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Iat           string `json:"iat"`
	Exp           string `json:"exp"`
	Alg           string `json:"alg"`
	Kid           string `json:"kid"`
	Typ           string `json:"typ"`
}

func (tp *GoogleTokenPayload) GetIdentType() IdentityType {
	return IdentityTypeGoogle
}
func (tp *GoogleTokenPayload) GetNickname() string {
	//if tp.Name != "" {
	//	return "username"
	//}
	return tp.Name
}
func (tp *GoogleTokenPayload) GetIdentifier() string {
	return tp.Sub
}
func (tp *GoogleTokenPayload) GetAccount() string {
	return tp.Email
}
func (tp *GoogleTokenPayload) GetAvatar() string {
	//if tp.Picture != "" {
	//	return "https://img.chatie.live/app/room_logo.png"
	//}
	return tp.Picture
}
func (tp *GoogleTokenPayload) GetCredential() string {
	return ""
}

/*
*

	{
	  "iss": "https://accounts.google.com",
	  "azp": "500262600030-8ltad8rl3ened17uqtdt8rt6n7uc7s05.apps.googleusercontent.com",
	  "aud": "500262600030-lv3gmkdt99p8l2617nm0etm76nofqlms.apps.googleusercontent.com",
	  "sub": "104197496379117818659",
	  "email": "sww19890618@gmail.com",
	  "email_verified": "true",
	  "name": "申文文",
	  "picture": "https://lh3.googleusercontent.com/a/ACg8ocL6fzlKRMy7Zq3ieDzdoY7dH6R2yYg3e-MjSo2PI3tXVhUJRQ=s96-c",
	  "given_name": "文文",
	  "family_name": "申",
	  "iat": "1728876202",
	  "exp": "1728879802",
	  "alg": "RS256",
	  "kid": "a50f6e70ef4b548a5fd9142eecd1fb8f54dce9ee",
	  "typ": "JWT"
	}

{"iss":"https://accounts.google.com","azp":"500262600030-8ltad8rl3ened17uqtdt8rt6n7uc7s05.apps.googleusercontent.com","aud":"500262600030-lv3gmkdt99p8l2617nm0etm76nofqlms.apps.googleusercontent.com","sub":"104197496379117818659","email":"sww19890618@gmail.com","email_verified":"true","name":"申文文","picture":"https://lh3.googleusercontent.com/a/ACg8ocL6fzlKRMy7Zq3ieDzdoY7dH6R2yYg3e-MjSo2PI3tXVhUJRQ=s96-c","given_name":"文文","family_name":"申","iat":"1728958789","exp":"1728962389","alg":"RS256","kid":"a50f6e70ef4b548a5fd9142eecd1fb8f54dce9ee","typ":"JWT"}
*/
func GoogleAuthToken(ctx context.Context, token string) (_ AuthPayload, err error) {
	var res GoogleTokenPayload
	url := GoogleAuthTokenUrl + "?id_token=" + token
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get google auth token fail")
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return nil, err
	}
	if res.EmailVerified != "true" {
		return nil, fmt.Errorf("google auth token verify fail")
	}
	return &res, nil
}
