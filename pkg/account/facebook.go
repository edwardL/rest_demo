package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type FacebookTokenPayload struct {
	Name    string `json:"name"`
	Picture struct {
		Data struct {
			Height       int    `json:"height"`
			IsSilhouette bool   `json:"is_silhouette"`
			Url          string `json:"url"`
			Width        int    `json:"width"`
		} `json:"data"`
	} `json:"picture"`
	Id string `json:"id"`
}

func (tp *FacebookTokenPayload) GetIdentType() IdentityType {
	return IdentityTypeFacebook
}
func (tp *FacebookTokenPayload) GetNickname() string {
	return tp.Name
}
func (tp *FacebookTokenPayload) GetIdentifier() string {
	return tp.Id
}
func (tp *FacebookTokenPayload) GetAccount() string {
	return ""
}
func (tp *FacebookTokenPayload) GetAvatar() string {
	return tp.Picture.Data.Url
}
func (tp *FacebookTokenPayload) GetCredential() string {
	return ""
}

/*
*

	{
	  "name": "Wenwen Shen",
	  "picture": {
	    "data": {
	      "height": 50,
	      "is_silhouette": true,
	      "url": "https://scontent-nrt1-1.xx.fbcdn.net/v/t1.30497-1/84628273_176159830277856_972693363922829312_n.jpg?stp=c379.0.1290.1290a_cp0_dst-jpg_s50x50&_nc_cat=1&ccb=1-7&_nc_sid=7565cd&_nc_ohc=TWA1H4Q6GogQ7kNvgFQ-uCU&_nc_ht=scontent-nrt1-1.xx&edm=AP4hL3IEAAAA&_nc_gid=AShsPZhqchP6CAcGqFFc5So&oh=00_AYBhQM8iX1DFfq8iSZnyxRjkZZr4MBNW4FXtz4dHaCr-Yw&oe=6733FE19",
	      "width": 50
	    }
	  },
	  "id": "122098746578120621"
	}

{"name":"Wenwen Shen","picture":{"data":{"height":50,"is_silhouette":true,"url":"https://scontent.fdxb2-1.fna.fbcdn.net/v/t1.30497-1/84628273_176159830277856_972693363922829312_n.jpg?stp=c379.0.1290.1290a_cp0_dst-jpg_s50x50\u0026_nc_cat=1\u0026ccb=1-7\u0026_nc_sid=7565cd\u0026_nc_ohc=TWA1H4Q6GogQ7kNvgGalcHt\u0026_nc_zt=24\u0026_nc_ht=scontent.fdxb2-1.fna\u0026edm=AP4hL3IEAAAA\u0026_nc_gid=AW7OQs3sj0K6fhJspOlUqFu\u0026oh=00_AYCWCDQ2wpObcWRtJOdnZd84BAQsOUs6L80qqZLuAPFFDA\u0026oe=67354F99","width":50}},"id":"122098746578120621"}
*/
func FacebookAuthToken(ctx context.Context, token string) (_ AuthPayload, err error) {
	res := FacebookTokenPayload{}
	url := fmt.Sprintf("https://graph.facebook.com/v3.3/me?access_token=%s&fields=name,picture&method=get&pretty=0&sdk=joey&suppress_http_code=1", token)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get facebook auth token fail")
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
	if res.Id == "" {
		return nil, fmt.Errorf("facebook auth token verify fail")
	}
	return &res, nil
}
