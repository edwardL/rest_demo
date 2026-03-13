package account

import (
	"context"
	"errors"
)

var ErrUnsupported = errors.New("unsupported account")

type IdentityType uint8

const IdentityTypeMobilePhone IdentityType = 1 //手机
const IdentityTypeGoogle IdentityType = 3      //google账号
const IdentityTypeFacebook IdentityType = 2    //fb账号
const IdentityTypeApple IdentityType = 6       //apple账号
const IdentityTypeSnapchat IdentityType = 8    //snapchat账号

type AuthPayload interface {
	GetIdentType() IdentityType //必须有
	GetIdentifier() string      //必须有
	GetAccount() string         //可能有
	GetNickname() string        //可能有
	GetAvatar() string          //可能有
	GetCredential() string      //mobilephone 登录必须有,密码
}

type Auth interface {
	AuthToken(ctx context.Context, token string) (_ AuthPayload, err error)
}

type AuthFunc func(ctx context.Context, token string) (_ AuthPayload, err error)

func (fn AuthFunc) AuthToken(ctx context.Context, token string) (_ AuthPayload, err error) {
	return fn(ctx, token)
}

type Config struct {
	Apple AppleConfig
}

type ThirdAccount struct {
	accounts map[IdentityType]Auth
}

func NewAccountAuth(conf Config) *ThirdAccount {
	thirdAccounts := map[IdentityType]Auth{
		IdentityTypeApple:    NewThirdAccountApple(conf.Apple),
		IdentityTypeGoogle:   AuthFunc(GoogleAuthToken),
		IdentityTypeSnapchat: AuthFunc(SnapchatAuthToken),
		IdentityTypeFacebook: AuthFunc(FacebookAuthToken),
	}
	return &ThirdAccount{
		accounts: thirdAccounts,
	}
}

func (account *ThirdAccount) AuthToken(ctx context.Context, identType IdentityType, token string) (_ AuthPayload, err error) {
	thirdAccount, ok := account.accounts[identType]
	if !ok {
		return nil, ErrUnsupported
	}
	return thirdAccount.AuthToken(ctx, token)
}
