package legacyjson

type Response struct {
	RetCode int         `json:"retcode"`
	ErrMsg  string      `json:"errmsg"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(data interface{}) Response {
	return Response{
		RetCode: 0,
		ErrMsg:  "",
		Data:    data,
	}
}

func Error(message string) Response {
	return Response{
		RetCode: -1,
		ErrMsg:  message,
	}
}

func Info(message string) Response {
	return Response{
		RetCode: -2,
		ErrMsg:  message,
	}
}

func Jump(message string) Response {
	return Response{
		RetCode: -3,
		ErrMsg:  message,
	}
}

func NotFound(message string) Response {
	return Response{
		RetCode: -4,
		ErrMsg:  message,
	}
}
