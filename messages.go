package dslink

type Message struct {
	Msg  int32       `json:"msg" msgpack:"msg"`
	Ack  int32       `json:"ack,omitempty" msgpack:"ack,omitempty"`
	Reqs []*Request  `json:"requests,omitempty" msgpack:"requests,omitempty"`
	Resp []*Response `json:"responses,omitempty" msgpack:"responses,omitempty"`
	Salt string      `json:"salt,omitempty" msgpack:"salt,omitempty"`
}

type Request struct {
	Rid    int32     `json:"rid" msgpack:"rid"`
	Method string    `json:"method" msgpack:"method"`
	Path   string    `json:"path,omitempty" msgpack:"path,omitempty"`
	Paths  []SubPath `json:"paths,omitempty" msgpack:"paths,omitempty"`
	Sids   []int     `json:"sids,omitempty" msgpack:"sids,omitempty"`
}

type SubPath struct {
	Path string `json:"path" msgpack:"path"`
	Sid  int    `json:"sid" msgpack:"sid"`
	Qos  uint8  `json:"qos,omitempty" msgpack:"qos,omitempty"`
}

type Response struct {
	Rid     int32               `json:"rid" msgpack:"rid"`
	Stream  string              `json:"stream" msgpack:"stream"`
	Updates [][]interface{}     `json:"updates" msgpack:"updates"`
	Columns []map[string]string `json:"columns,omitempty" msgpack:"columns,omitempty"`
	Error   *MsgErr             `json:"error,omitempty" msgpack:"error,omitempty"`
}

func (r *Response) AddUpdate(name string, value interface{}) {
	var u []interface{}
	u = append(u, name, value)
	r.Updates = append(r.Updates, u)
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
