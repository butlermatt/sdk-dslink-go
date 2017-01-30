package link

type message struct {
	Msg  int32      `json:"msg"`
	Ack  int32      `json:"ack"`
	Reqs []request  `json:"requests"`
	Resp []response `json:"responses"`
	Salt string     `json:"salt"`
}

type request struct {
	Rid    int32  `json:"rid"`
	Method string `json:"method"`
	Path   string `json:"path"`
}

type response struct {
	Rid     int32         `json:"rid"`
	Stream  string        `json:"stream"`
	Updates []interface{} `json:"updates"`
	Error   msgErr        `json:"error"`
}

type msgErr struct {
	Type   string `json:"type"`
	Msg    string `json:"msg"`
	Phase  string `json:"phase"`
	Path   string `json:"path"`
	Detail string `json:"detail"`
}
