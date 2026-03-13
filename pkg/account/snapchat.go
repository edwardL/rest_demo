package account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SnapchatTokenPayload struct {
	Errors []interface{} `json:"errors"`
	Data   struct {
		Me struct {
			DisplayName string `json:"displayName"`
			Bitmoji     struct {
				Avatar string `json:"avatar"`
			} `json:"bitmoji"`
			ExternalId string `json:"externalId"`
		} `json:"me"`
	} `json:"data"`
}

func (tp *SnapchatTokenPayload) GetIdentType() IdentityType {
	return IdentityTypeSnapchat
}
func (tp *SnapchatTokenPayload) GetNickname() string {
	return tp.Data.Me.DisplayName
}
func (tp *SnapchatTokenPayload) GetIdentifier() string {
	return tp.Data.Me.ExternalId
}
func (tp *SnapchatTokenPayload) GetAccount() string {
	return ""
}
func (tp *SnapchatTokenPayload) GetAvatar() string {
	return tp.Data.Me.Bitmoji.Avatar
}
func (tp *SnapchatTokenPayload) GetCredential() string {
	return ""
}

// {"errors":[],"data":{"me":{"displayName":"茶铁","bitmoji":{"avatar":"https://sdk.bitmoji.com/render/panel/2e85858e-0458-4503-88d9-ce0fc1c72205-uYQ7jn8uvBiYRa76Gk5u8jeBswg6Wks5vf9aZpuxPBW_aeU1msQQoQ-v1.png?transparent=1\u0026palette=1"},"externalId":"CAESIEfyOE4YgLrTQ1QQ+jF+9cBUfIXpCtTVi2eUJrUKusH4"}}}
func SnapchatAuthToken(ctx context.Context, token string) (_ AuthPayload, err error) {
	var res SnapchatTokenPayload
	//testToken := []byte("{\"errors\":[],\"data\":{\"me\":{\"displayName\":\"茶铁\",\"bitmoji\":{\"avatar\":\"https://sdk.bitmoji.com/render/panel/2e85858e-0458-4503-88d9-ce0fc1c72205-uYQ7jn8uvBiYRa76Gk5u8jeBswg6Wks5vf9aZpuxPBW_aeU1msQQoQ-v1.png?transparent=1\\u0026palette=1\"},\"externalId\":\"CAESIEfyOE4YgLrTQ1QQ+jF+9cBUfIXpCtTVi2eUJrUKusH4\"}}}")
	//err = json.Unmarshal(testToken, &res)
	//return &res, err
	content := map[string]interface{}{
		"query": "{me{displayName bitmoji{avatar} externalId}}",
	}
	reqBody, _ := json.Marshal(content)
	req, err := http.NewRequestWithContext(ctx,
		"POST",
		"https://kit.snapchat.com/v1/me",
		bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get snapchat auth token fail")
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
	if len(res.Errors) != 0 || res.Data.Me.ExternalId == "" {
		return nil, fmt.Errorf("snapchat auth token verify fail")
	}
	return &res, nil
}
