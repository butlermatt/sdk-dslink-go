package dslink

type Message struct {
	Msg  int32      `json:"msg" msgpack:"msg"`
	Ack  int32      `json:"ack,omitempty" msgpack:"ack,omitempty"`
	Reqs []Request  `json:"requests,omitempty" msgpack:"requests,omitempty"`
	Resp []Response `json:"responses,omitempty" msgpack:"responses,omitempty"`
	Salt string     `json:"salt,omitempty" msgpack:"salt,omitempty"`
}

type Request struct {
	Rid    int32  `json:"rid" msgpack:"rid"`
	Method string `json:"method" msgpack:"method"`
	Path   string `json:"path" msgpack:"path"`
}

type Response struct {
	Rid     int32         `json:"rid" msgpack:"rid"`
	Stream  string        `json:"stream" msgpack:"stream"`
	Updates []interface{} `json:"updates" msgpack:"updates"`
	Error   MsgErr        `json:"error" msgpack:"error"`
}

func NewResp(rid int32) *Response {
	return &Response{Rid: rid}
}

type MsgErr struct {
	Type   string `json:"type" msgpack:"type"`
	Msg    string `json:"msg" msgpack:"msg"`
	Phase  string `json:"phase" msgpack:"phase"`
	Path   string `json:"path" msgpack:"path"`
	Detail string `json:"detail" msgpack:"detail"`
}
