package dslink

import (
	"fmt"
	"strings"
	"time"
)

type Message struct {
	Msg  int32       `json:"msg" msgpack:"msg"`
	Ack  int32       `json:"ack,omitempty" msgpack:"ack,omitempty"`
	Reqs []*Request  `json:"requests,omitempty" msgpack:"requests,omitempty"`
	Resp []*Response `json:"responses,omitempty" msgpack:"responses,omitempty"`
	Salt string      `json:"salt,omitempty" msgpack:"salt,omitempty"`
}

func (m *Message) String() string {
	s := fmt.Sprintf(`{"msg": %d, "ack": %d`, m.Msg, m.Ack)
	if len(m.Reqs) > 0 {
		s = fmt.Sprintf(`%s, "requests": [`, s)
		var l []string
		for _, req := range m.Reqs {
			l = append(l, req.String())
		}
		s = fmt.Sprintf(`%s%s]`, s, strings.Join(l, ", "))
	}

	if len(m.Resp) > 0 {
		s = fmt.Sprintf(`%s, "responses": [`, s)
		var l []string
		for _, resp := range m.Resp {
			l = append(l, resp.String())
		}
		s = fmt.Sprintf(`%s%s]`, s, strings.Join(l, ", "))
	}

	if m.Salt != "" {
		s = fmt.Sprintf(`%s, "salt": "%s"`, s, m.Salt)
	}

	s += "}"
	return s
}

type Request struct {
	Rid    int32      `json:"rid" msgpack:"rid"`
	Method MethodType `json:"method" msgpack:"method"`
	Path   string     `json:"path,omitempty" msgpack:"path,omitempty"`
	Paths  []*SubPath `json:"paths,omitempty" msgpack:"paths,omitempty"`
	Sids   []int32    `json:"sids,omitempty" msgpack:"sids,omitempty"`
}

type MethodType string;
const (
	MethodList    MethodType = "list"
	MethodSub     MethodType = "subscribe"
	MethodUnsub   MethodType = "unsubscribe"
	MethodClose   MethodType = "close"
	MethodSet     MethodType = "set"
	MethodRemove  MethodType = "remove"
	MethodInvoke  MethodType = "invoke"
)

func (r *Request) String() string {
	s := fmt.Sprintf(`{"rid": %d, "method": "%s"`, r.Rid, r.Method)
	if r.Path != "" {
		s = fmt.Sprintf(`%s, "path": "%s"`, s, r.Path)
	}
	if len(r.Paths) > 0 {
		var l []string
		s += `, "paths": [`
		for _, p := range r.Paths {
			l = append(l, p.String())
		}
		s = fmt.Sprintf(`%s%s]`, s, strings.Join(l, ", "))
	}
	if len(r.Sids) > 0 {
		s = fmt.Sprintf(`%s, "sids": %v`, s, r.Sids)
	}
	s += "}"
	return s
}

type SubPath struct {
	Path string `json:"path" msgpack:"path"`
	Sid  int32  `json:"sid" msgpack:"sid"`
	Qos  uint8  `json:"qos,omitempty" msgpack:"qos,omitempty"`
}

func (sp *SubPath) String() string {
	return fmt.Sprintf(`{"path": "%s", "sid": %d, "qos": %d}`, sp.Path, sp.Sid, sp.Qos)
}

type StreamState string;
const (
	StreamInit   StreamState = "initialize"
	StreamOpen   StreamState = "open"
	StreamClosed StreamState = "closed"
)

type Response struct {
	Rid     int32               `json:"rid" msgpack:"rid"`
	Stream  StreamState         `json:"stream" msgpack:"stream"`
	Updates []interface{}       `json:"updates" msgpack:"updates"`
	Columns []map[string]string `json:"columns,omitempty" msgpack:"columns,omitempty"`
	Error   *MsgErr             `json:"error,omitempty" msgpack:"error,omitempty"`
}

func (r *Response) String() string {
	s := fmt.Sprintf(`{"rid": %d, "stream": "%s", "updates": %v`, r.Rid, r.Stream, r.Updates)
	if len(r.Columns) > 0 {
		s = fmt.Sprintf(`%s, "columns": %v`, s, r.Columns)
	}
	if r.Error != nil {
		s = fmt.Sprintf(`%s, "error": %v`, s, r.Error)
	}

	s += "}"
	return s
}

func (r *Response) AddUpdate(name interface{}, value interface{}) {
	switch t := value.(type) {
	case *ValueUpdate:
		m := make(map[string]interface{})
		m[`ts`] = t.ts.Format(time.RFC3339Nano)
		m[`sid`] = name
		m[`value`] = t.Value()
		r.Updates = append(r.Updates, m)
	default:
		var u []interface{}
		u = append(u, name, value)
		r.Updates = append(r.Updates, u)
	}
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

func (e *MsgErr) String() string {
	return fmt.Sprintf(`{"type": %q, "msg": %q, "phase": %q, "path": %q, "detail": %q`,
		e.Type, e.Msg, e.Phase, e.Path, e.Detail)
}
