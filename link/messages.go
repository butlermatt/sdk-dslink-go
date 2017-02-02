package link

type message struct {
	Msg  int32      `json:"msg"`
	Ack  int32      `json:"ack"`
	Reqs []Request  `json:"requests,omitempty"`
	Resp []Response `json:"responses,omitempty"`
	Salt string     `json:"salt,omitempty"`
}

type Request struct {
	Rid    int32  `json:"rid"`
	Method string `json:"method"`
	Path   string `json:"path"`
}

type Response struct {
	Rid     int32         `json:"rid"`
	Stream  string        `json:"stream"`
	Updates []interface{} `json:"updates"`
	Error   msgErr        `json:"error"`
}

func NewResp(rid int32) *Response {
	return &Response{Rid: rid}
}

type msgErr struct {
	Type   string `json:"type"`
	Msg    string `json:"msg"`
	Phase  string `json:"phase"`
	Path   string `json:"path"`
	Detail string `json:"detail"`
}
