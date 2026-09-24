package apperror

// Error 表示可安全返回给客户端的业务错误。
type Error struct {
	Status  int
	Reason  string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// New 创建业务错误。
func New(status int, code string, message string) error {
	return &Error{Status: status, Reason: code, Message: message}
}

// WrapInternal 创建内部错误，避免向客户端泄露数据库细节。
func WrapInternal(err error) error {
	_ = err
	return &Error{Status: 500, Reason: "INTERNAL_ERROR", Message: "服务暂时不可用"}
}
