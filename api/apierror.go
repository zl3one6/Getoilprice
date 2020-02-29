package api

type APIException struct {
	Code    int    `json:"-"`
	ErrCode int    `json:"errcode"`
	Msg     string `json:"msg"`
	Request string `json:"request"`
}

func (e *APIException) Error() string {
	return e.Msg
}

func NewAPIException(code int, errcode int, msg string) *APIException {
	return &APIException{
		Code:    code,
		ErrCode: errcode,
		Msg:     msg,
	}
}
