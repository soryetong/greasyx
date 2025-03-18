package xerror

// 1开头系统校验类,2开头用户及用户行为校验类
const (
	OK                   = 200  // 通用-Success
	Error                = 400  // 通用-ERROR
	ServerError          = 500  // 系统错误
	ParameterIllegal     = 1002 // 参数不合法
	NoAuth               = 1003 // 权限不足
	NotData              = 1004 // 没有数据
	HasData              = 1005 // 数据已存在
	UnauthorizedToken    = 1006 // 非法的用户token
	NeedLogin            = 1007 // 请先登录
	RequestLimit         = 1008 // 请求频繁,请稍后再试
	CaptchaGenerateError = 1009 // 验证码生成错误
	CaptchaError         = 1010 // 验证码错误
	LoginFail            = 2000 // 登录失败
	LoginPasswordError   = 2001 // 密码错误
	LoginNoUser          = 2002 // 该用户不存在
	LoginBan             = 2003 // 你暂时不能进行登录操作
	ThirdLoginError      = 2004 // 第三方登录失败
	UserIsset            = 2005 // 用户已存在
)

func GetErrorMessage(code int64, message ...string) string {
	var codeMessage string
	codeMap := map[int64]string{
		OK:                   "Success",
		Error:                "请求错误",
		ServerError:          "系统错误",
		ParameterIllegal:     "参数不合法",
		NoAuth:               "权限不足",
		NotData:              "没有数据",
		HasData:              "数据已存在",
		UnauthorizedToken:    "非法的用户token",
		NeedLogin:            "请先登录",
		RequestLimit:         "请求频繁,请稍后再试",
		CaptchaGenerateError: "验证码生成错误",
		CaptchaError:         "验证码错误",
		LoginFail:            "登录失败",
		LoginPasswordError:   "密码错误",
		LoginNoUser:          "该用户不存在",
		LoginBan:             "你暂时不能进行登录操作",
		ThirdLoginError:      "第三方登录失败",
		UserIsset:            "用户已存在",
	}

	if len(message) > 0 {
		codeMessage = message[0]
	} else {
		if value, ok := codeMap[code]; ok {
			codeMessage = value
		} else {
			codeMessage = "系统错误!"
		}
	}

	return codeMessage
}
