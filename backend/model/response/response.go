package response

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

const (
	SUCCESS      = 200
	ERROR        = 500
	UNAUTHORIZED = 401
)

func Result(code int, data interface{}, msg string) Response {
	return Response{
		Code: code,
		Data: data,
		Msg:  msg,
	}
}

func Success(data interface{}) Response {
	return Result(SUCCESS, data, "操作成功")
}

func SuccessWithMsg(data interface{}, msg string) Response {
	return Result(SUCCESS, data, msg)
}

func Fail(data interface{}) Response {
	return Result(ERROR, data, "操作失败")
}

func FailWithMsg(data interface{}, msg string) Response {
	return Result(ERROR, data, msg)
}

func Unauthorized(msg string) Response {
	return Result(UNAUTHORIZED, nil, msg)
}
