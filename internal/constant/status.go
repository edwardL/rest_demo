package constant

const StatusEnable = 1
const StatusDisable = 2

type Status uint8

func (f Status) Enabled() bool {
	return f == StatusEnable
}

func (f Status) Disabled() bool {
	return f == StatusDisable
}

const CtxTokenKey = "token_claims"
