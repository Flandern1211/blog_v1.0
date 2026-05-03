package errors

// 错误码定义
const (
	// 成功
	CodeSuccess = 0

	// 通用错误 4xx
	CodeBadRequest       = 400
	CodeUnauthorized     = 401
	CodeForbidden        = 403
	CodeNotFound         = 404
	CodeMethodNotAllowed = 405
	CodeRequestTimeout   = 408
	CodeConflict         = 409

	CodeInvalidParam = 400
	CodeMissingParam = 400

	// 服务器错误 5xx
	CodeInternalError      = 500
	CodeNotImplemented     = 501
	CodeServiceUnavailable = 503

	// 业务错误 1xxx
	CodeUserNotFound       = 1003
	CodeUserAlreadyExists  = 6103
	CodeInvalidCredentials = 1002
	CodeUserDisabled       = 1004
	CodeInvalidToken       = 1203
	CodeTokenExpired       = 1202

	// 其他业务错误
	CodeRequestError     = 9001
	CodeDbOpError        = 9004
	CodeRedisOpError     = 9005
	CodeUserAuthError    = 9006
	CodeOldPasswordErr   = 1010
	CodeTokenNotExist    = 1201
	CodeTokenTypeErr     = 1204
	CodeTokenCreateErr   = 1205
	CodePermissionErr    = 1206
	CodeForceOffline     = 1207
	CodeFileUploadErr    = 9100
	CodeFileReceiveErr   = 9101
	CodeTagHasArt        = 4003
	CodeCateHasArt       = 3003
	CodeResourceNotExist = 6002
	CodeMenuNotExist     = 6006
	CodeSendEmailErr     = 6101
	CodeCodeWrong        = 6102
	CodeEmailExist       = 6104
	CodeNoLogin          = 6105

	CodeForceOfflineSelf    = 1208
	CodeNoAdminAccess       = 1209
	CodeResourceUsedByRole  = 6003
	CodeResourceHasChildren = 6004
	CodeMenuUsedByRole      = 6007
	CodeMenuHasChildren     = 6008
)

// 错误码对应的文本消息
var codeMessages = map[int]string{
	CodeSuccess:             "成功",
	CodeBadRequest:          "请求参数错误",
	CodeUnauthorized:        "未授权",
	CodeForbidden:           "禁止访问",
	CodeNotFound:            "资源不存在",
	CodeMethodNotAllowed:    "方法不允许",
	CodeRequestTimeout:      "请求超时",
	CodeConflict:            "资源冲突",
	CodeInternalError:       "服务器内部错误",
	CodeNotImplemented:      "功能未实现",
	CodeServiceUnavailable:  "服务不可用",
	CodeUserNotFound:        "该用户不存在",
	CodeUserAlreadyExists:   "该用户名已存在",
	CodeInvalidCredentials:  "密码错误",
	CodeUserDisabled:        "该账号已被禁用",
	CodeInvalidToken:        "TOKEN 不正确，请重新登陆",
	CodeTokenExpired:        "TOKEN 已过期，请重新登陆",
	CodeRequestError:        "请求参数格式错误",
	CodeDbOpError:           "数据库操作异常",
	CodeRedisOpError:        "Redis 操作异常",
	CodeUserAuthError:       "用户认证异常",
	CodeOldPasswordErr:      "旧密码不正确",
	CodeTokenNotExist:       "TOKEN 不存在，请重新登陆",
	CodeTokenTypeErr:        "TOKEN 格式错误，请重新登陆",
	CodeTokenCreateErr:      "TOKEN 生成失败",
	CodePermissionErr:       "权限不足",
	CodeForceOffline:        "您已被强制下线",
	CodeFileUploadErr:       "文件上传失败",
	CodeFileReceiveErr:      "文件接收失败",
	CodeTagHasArt:           "删除失败，标签下存在文章",
	CodeCateHasArt:          "删除失败，分类下存在文章",
	CodeResourceNotExist:    "该资源不存在",
	CodeMenuNotExist:        "该菜单不存在",
	CodeSendEmailErr:        "发送邮件失败",
	CodeCodeWrong:           "验证码错误或已过期",
	CodeEmailExist:          "该邮箱已经注册",
	CodeNoLogin:             "用户未登录",
	CodeForceOfflineSelf:    "不能强制下线自己",
	CodeNoAdminAccess:       "仅管理员可登录后台",
	CodeResourceUsedByRole:  "该资源正在被角色使用，无法删除",
	CodeResourceHasChildren: "该资源下存在子资源，无法删除",
	CodeMenuUsedByRole:      "该菜单正在被角色使用，无法删除",
	CodeMenuHasChildren:     "该菜单下存在子菜单，无法删除",
}

// GetMessage 获取错误码对应的文本消息
func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
