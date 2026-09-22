// Package errcode 定义后端统一的业务错误码。
//
// 与 HTTP 状态码解耦，前端根据 code 判断业务结果，与 HTTP 状态码无关。
package errcode

// 业务错误码定义。
const (
	// CodeOK 成功
	CodeOK = 200
	// CodeBadRequest 参数错误
	CodeBadRequest = 400
	// CodeUnauthorized 未登录
	CodeUnauthorized = 401
	// CodeForbidden 无权限
	CodeForbidden = 403
	// CodeNotFound 资源不存在
	CodeNotFound = 404
	// CodeConflict 资源冲突（如重复）
	CodeConflict = 409
	// CodeTooManyRequests 限流
	CodeTooManyRequests = 429
	// CodeInternal 服务器内部错误
	CodeInternal = 500
	// CodeBadGateway 上游错误（AI 模型等）
	CodeBadGateway = 502
	// CodeUnavailable 服务暂不可用
	CodeUnavailable = 503
	// CodeGatewayTimeout 上游超时
	CodeGatewayTimeout = 504
)

// CodeMessage 业务错误码到默认消息的映射。
var CodeMessage = map[int]string{
	CodeOK:                "OK",
	CodeBadRequest:        "参数错误",
	CodeUnauthorized:      "未登录或登录已过期",
	CodeForbidden:         "没有访问权限",
	CodeNotFound:          "资源不存在",
	CodeConflict:          "资源冲突",
	CodeTooManyRequests:   "操作过于频繁",
	CodeInternal:          "服务器内部错误",
	CodeBadGateway:        "上游服务异常",
	CodeUnavailable:       "服务暂不可用",
	CodeGatewayTimeout:    "上游服务超时",
}

// Message 返回业务错误码对应的默认文案。
func Message(code int) string {
	if msg, ok := CodeMessage[code]; ok {
		return msg
	}
	return "未知错误"
}
