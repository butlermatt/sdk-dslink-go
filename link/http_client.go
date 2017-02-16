package link

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

import (
	"bytes"
	"github.com/butlermatt/dslink"
	"github.com/butlermatt/dslink/crypto"
	"github.com/gorilla/websocket"
	"gopkg.in/vmihailenco/msgpack.v2"
)

const pingTime = 30 * time.Second
const maxMsgId = 0x7FFFFFFF

type msgFormat int

const (
	fmtJson msgFormat = iota
	fmtMsgP
)

type encode interface {
	Encode(interface{}) error
}

type decode interface {
	Decode(interface{}) error
}

type dsResp struct {
	Id        string `json:"id"`
	PublicKey string `json:"publicKey"`
	WsUri     string `json:"wsUri"`
	HttpUri   string `json:"httpUri"`
	Version   string `json:"version"`
	TempKey   string `json:"tempKey"`
	Salt      string `json:"salt"`
	SaltS     string `json:"saltS"`
	SaltL     string `json:"saltL"`
	Path      string `json:"path"`
	Format    string `json:"format"`
}

type httpClient struct {
	enc       encode
	dec       decode
	encBuf    bytes.Buffer
	decBuf    bytes.Buffer
	dsId      string
	msgId     int32
	reqId     uint32
	keyMaker  crypto.ECDH
	htClient  *http.Client
	rawUrl    *url.URL
	home      string
	token     string
	tHash     string
	wsClient  *websocket.Conn
	cPriv     crypto.PrivateKey
	in        chan []byte
	out       chan *dslink.Message
	ping      *time.Timer
	msgs      chan *dslink.Message
	format    msgFormat
	responder bool
	requester bool
}

// Close will force the Websocket on the httpClient to be closed.
func (c *httpClient) Close() {
	if c.wsClient != nil {
		_ = c.wsClient.Close()
	}
}

func (c *httpClient) getWsConfig() (*dsResp, error) {
	u, _ := url.Parse(c.rawUrl.String())
	q := u.Query()
	q.Add("dsId", c.dsId)
	if c.home != "" {
		q.Add("home", c.home)
	}
	if c.tHash != "" {
		q.Add("token", c.token+c.tHash)
	}
	u.RawQuery = q.Encode()

	// TODO: Put this in a struct!
	values := fmt.Sprintf("{\"publicKey\": \"%s\", \"isRequester\": %t, \"isResponder\": %t,"+
		"\"linkData\": {}, \"version\": \"1.1.2\", \"formats\": [\"msgpack\",\"json\"], \"enableWebSocketCompression\": true}",
		c.cPriv.PublicKey.Base64(), c.requester, c.responder)
	res, err := c.htClient.Post(u.String(), "application/json", strings.NewReader(values))
	if err != nil {
		return nil, fmt.Errorf("Error connecting to address: \"%s\"\nError: %s", c.rawUrl, err)
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read response: %s", err)
	}

	dr := &dsResp{}
	if err = json.Unmarshal(b, dr); err != nil {
		return nil, fmt.Errorf("Unable to decode response: %s\nError: %s", b, err)
	}
	log.Printf("Received configuration: %+v\n", *dr)
	return dr, nil
}

func (c *httpClient) connectWs(config *dsResp) (*websocket.Conn, error) {
	switch config.Format {
	case "json":
		c.format = fmtJson
	case "msgpack":
		c.format = fmtMsgP
	default:
		return nil, fmt.Errorf("Unknown message format: %s", config.Format)
	}

	sPub, err := c.keyMaker.UnmarshalPublic(config.TempKey)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse server key: %s\nError: %s", config.TempKey, err)
	}

	shared := c.keyMaker.GenerateSharedSecret(c.cPriv, sPub)
	auth := c.keyMaker.HashSalt(config.Salt, shared)

	u, err := c.rawUrl.Parse(config.WsUri)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse Websocket URL: %s\nError: %s", config.WsUri, err)
	}

	q := u.Query()
	q.Add("auth", auth)
	q.Add("format", config.Format)
	q.Add("dsId", c.dsId)
	if c.home != "" {
		q.Add("home", c.home)
	}
	if c.tHash != "" {
		q.Add("token", c.token+c.tHash)
	}
	u.RawQuery = q.Encode()
	u.Scheme = "ws"

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to Websocket at: %s\nError: %s", u.String(), err)
	}

	return conn, nil
}

func (c *httpClient) marshal(v interface{}) (int, []byte, error) {
	var f int
	var d []byte
	var err error
	switch c.format {
	case fmtJson:
		f = websocket.TextMessage
		d, err = json.Marshal(v)
	case fmtMsgP:
		f = websocket.BinaryMessage
		d, err = msgpack.Marshal(v)
	}

	return f, d, err
}

func (c *httpClient) unmarshal(data []byte, v interface{}) error {
	var err error
	switch c.format {
	case fmtJson:
		err = json.Unmarshal(data, v)
	case fmtMsgP:
		err = msgpack.Unmarshal(data, v)
	}
	return err
}

func (c *httpClient) handleConnections() {
	go func() {
		for {
			_, p, err := c.wsClient.ReadMessage()
			if err != nil {
				//TODO: Better logging/handling here
				log.Printf("Read error! %v\n", err)
				return
			}
			c.in <- p
		}
	}()

	c.wsClient.WriteMessage(websocket.TextMessage, []byte("{}"))
	for {
		select {
		case s := <-c.in:
			msg := &dslink.Message{Msg: -1, Ack: -1}
			err := c.unmarshal(s, msg)
			if err != nil {
				log.Printf("Error unmarshalling %s\nError: %v\n", s, err)
			}
			log.Printf("Recv: %v", msg)
			c.msgs <- msg
		case m := <-c.out:
			if c.msgId == maxMsgId {
				c.msgId = 0
			}
			c.msgId++
			m.Msg = c.msgId
			t, s, err := c.marshal(*m)
			if err != nil {
				log.Printf("Error marshalling %+v\nError: %+v\n", *m, err)
				continue
			}
			log.Printf("Sent: %v\n", m)
			c.wsClient.WriteMessage(t, s)
			if !c.ping.Stop() {
				<-c.ping.C
			}
			c.ping.Reset(pingTime)
		case <-c.ping.C:
			go func() {
				m := &dslink.Message{Msg: c.msgId}
				c.out <- m
			}()
			c.ping.Reset(pingTime)
		}
	}
}

// Dial will attempt to connect a link with the specified prefix to the specified address.
// Returns an error if connection handshake fails. Otherwise returns the connected httpClient.
func dial(conf *Config, msgs chan *dslink.Message) (*httpClient, error) {
	u, err := url.Parse(conf.broker)
	if err != nil {
		return nil, err
	}

	c := &httpClient{
		keyMaker:  crypto.NewECDH(),
		htClient:  &http.Client{Timeout: time.Second * 60},
		rawUrl:    u,
		home:      conf.home,
		msgs:      msgs,
		responder: conf.isResponder,
		requester: conf.isRequester,
	}

	// TODO: The keys should be managed outside of the httpClient and
	// passed in as needed
	c.cPriv, err = crypto.LoadKey(conf.keyPath)
	if err != nil {
		c.cPriv, err = c.keyMaker.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("Unable to generate key: %v", err)
		}
		_ = crypto.SaveKey(c.cPriv, conf.keyPath)
	}
	c.dsId = c.cPriv.DsId(conf.name)

	if len(conf.token) >= 16 { // TODO: Why 16??
		c.token = conf.token[:16]
		c.tHash = c.keyMaker.HashToken(c.dsId, c.token)
	}

	ret, err := c.getWsConfig()
	if err != nil {
		return nil, err
	}

	conn, err := c.connectWs(ret)
	if err != nil {
		return nil, err
	}

	c.wsClient = conn
	c.ping = time.NewTimer(pingTime)
	c.in = make(chan []byte)
	c.out = make(chan *dslink.Message)

	go c.handleConnections()

	return c, nil
}
