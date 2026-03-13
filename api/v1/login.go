package v1

type LoginReq struct {
	Email     string `json:"email" binding:"required_without=Username" example:""`
	Username  string `json:"username" binding:"required_without=Email" example:"admin"` //登陆用户名
	Password  string `json:"password" binding:"required,min=6,max=30" example:"123456"` //登陆密码
	CaptchaId string `json:"captcha_id" binding:"omitempty"`
	Captcha   string `json:"captcha" binding:"omitempty"`
	Remember  bool   `json:"remember" binding:"omitempty"`
}

type LoginRes struct {
	UserId             int64  `json:"uid"`                  //用户id
	Token              string `json:"token"`                //token
	TokenExpire        int64  `json:"token_expire"`         //token过期时间
	RefreshToken       string `json:"refresh_token"`        //刷新token
	RefreshTokenExpire int64  `json:"refresh_token_expire"` //刷新token过期时间
}
