package account

// MobilePhonePayload 充数用的，没啥用途的
type MobilePhonePayload struct {
	Phone    string
	Password string
}

func NewMobilePhonePayload(phone, pwd string) *MobilePhonePayload {
	return &MobilePhonePayload{Phone: phone, Password: pwd}
}

func (tp *MobilePhonePayload) GetIdentType() IdentityType {
	return IdentityTypeMobilePhone
}
func (tp *MobilePhonePayload) GetNickname() string {
	return ""
}
func (tp *MobilePhonePayload) GetIdentifier() string {
	return tp.Phone
}
func (tp *MobilePhonePayload) GetAccount() string {
	return tp.Phone
}
func (tp *MobilePhonePayload) GetAvatar() string {
	return ""
}
func (tp *MobilePhonePayload) GetCredential() string {
	return tp.Password
}
