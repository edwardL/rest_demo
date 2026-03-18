package validator

// Option 选项函数类型：接收Server指针，用于配置参数
type Option func(c *validatorConf)

// Init 使用可选配置初始化校验器全局配置。
func Init(opts ...Option) {
	for _, opt := range opts {
		opt(conf)
	}
}

// WithConditionalFunction 设置条件验证函数
func WithConditionalFunction(fn string, f ConditionalFunction) Option {
	return func(c *validatorConf) {
		c.conditionalFunctionMap[fn] = f
	}
}

// WithConditionalFunctions 设置条件验证函数
func WithConditionalFunctions(fl map[string]ConditionalFunction) Option {
	return func(c *validatorConf) {
		for fn, f := range fl {
			c.conditionalFunctionMap[fn] = f
		}
	}
}

// WithCompareFunction 设置比较函数
func WithCompareFunction(fn string, f compareFunction) Option {
	return func(c *validatorConf) {
		c.compareFunctionMap[fn] = f
	}
}

// WithCompareFunctions 设置比较函数
func WithCompareFunctions(fl map[string]compareFunction) Option {
	return func(c *validatorConf) {
		for fn, f := range fl {
			c.compareFunctionMap[fn] = f
		}
	}
}

// WithErrCode 设置错误码对应的错误信息
func WithErrCode(code string, msg string) Option {
	return func(c *validatorConf) {
		c.validatorErrCodeMsg[code] = msg
	}
}

// WithErrCodesMsg 批量设置错误码对应的错误信息
func WithErrCodesMsg(ec map[string]string) Option {
	return func(c *validatorConf) {
		for k, v := range ec {
			c.validatorErrCodeMsg[k] = v
		}
	}
}

// WithErrCodeMsg 设置错误码对应的错误信息
func WithErrCodeMsg(code string, msg string) Option {
	return func(c *validatorConf) {
		c.validatorErrCodeMsg[code] = msg
	}
}

// WithDefErrCodeMsg 设置默认状态码和信息
func WithDefErrCodeMsg(code, msg string) Option {
	return func(c *validatorConf) {
		c.defErrCode = code
		c.defErrMsg = msg
	}
}
