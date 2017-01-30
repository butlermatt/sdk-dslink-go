package client

import (
	"net/http"
	"net/url"
	"time"
	"fmt"
	"io/ioutil"
	"crypto/rand"
	"strings"
	"encoding/json"
)

import (
	"github.com/butlermatt/dslink/crypto"
	"github.com/gorilla/websocket"
)

const pingTime = 45 * time.Second

type dsResp struct {
	Id 	  string `json:"id"`
	PublicKey string `json:"publicKey"`
	WsUri 	  string `json:"wsUri"`
	HttpUri   string `json:"httpUri"`
	Version   string `json:"version"`
	TempKey   string `json:"tempKey"`
	Salt 	  string `json:"salt"`
	SaltS 	  string `json:"saltS"`
	SaltL 	  string `json:"saltL"`
	Path 	  string `json:"path"`
	Format 	  string `json:"format"`
}

type httpClient struct {
	dsId	 string
	msgId	 uint32
	reqId	 uint32
	keyMaker crypto.ECDH
	htClient *http.Client
	rawUrl   *url.URL
	home	 string
	token    string
	tHash	 string
	wsClient *websocket.Conn
	cPriv 	 crypto.PrivateKey
	in       chan string
	out	 chan string
	ping     *time.Timer
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
		q.Add("token", c.token + c.tHash)
	}
	u.RawQuery = q.Encode()

	values := fmt.Sprintf("{\"publicKey\": \"%s\", \"isRequester\": false, \"isResponder\": true," +
		"\"linkData\": {}, \"version\": \"1.1.2\", \"formats\": [\"json\"], \"enableWebSocketCompression\": true}",
		c.cPriv.PublicKey.Base64())
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
	return dr, nil
}

func (c *httpClient) connectWs(config *dsResp) (*websocket.Conn, error) {
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
		q.Add("token", c.token + c.tHash)
	}
	u.RawQuery = q.Encode()
	u.Scheme = "ws"

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to Websocket at: %s\nError: %s", u.String(), err)
	}

	return conn, nil
}

func (c *httpClient) handleConnections() {
	go func() {
		for {
			mt, p, err := c.wsClient.ReadMessage()
			if err != nil {
				//TODO: Better logging/handling here
				fmt.Printf("Read error! %v\n", err)
				return
			}
			if mt != websocket.TextMessage {
				fmt.Println("Read error. Data is binary!?")
				return
			}
			c.in<-string(p)
		}
	}()

	c.wsClient.WriteMessage(websocket.TextMessage, []byte("{}"))
	for {
		select {
		case s := <-c.in:
			//TODO: Handle a received message
			fmt.Printf("Received message: %s\n", s)
		case o := <-c.out:
			c.wsClient.WriteMessage(websocket.TextMessage, []byte(o))
			if !c.ping.Stop() {
				<-c.ping.C
			}
			c.ping.Reset(pingTime)
		case <- c.ping.C:
			c.msgId++
			m := fmt.Sprintf("{\"msg\": %d}", c.msgId)
			c.wsClient.WriteMessage(websocket.TextMessage, []byte(m))
			c.ping.Reset(pingTime)
		}
	}
}

// Dial will attempt to connect a link with the specified prefix to the specified address.
// Returns an error if connection handshake fails. Otherwise returns the connected httpClient.
func Dial(addr, prefix, home, token string) (*httpClient, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	c := &httpClient{
		keyMaker: crypto.NewECDH(),
		htClient: &http.Client{Timeout: time.Second * 60},
		rawUrl: u,
		home: home,
	}

	// TODO: The keys should be managed outside of the httpClient and
	// passed in as needed
	c.cPriv, err = crypto.LoadKey("")
	if err != nil {
		c.cPriv, err = c.keyMaker.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("Unable to generate key: %v", err)
		}
		_ = crypto.SaveKey(c.cPriv,"")
	}
	c.dsId = c.cPriv.DsId(prefix)

	if len(token) >= 16 { // TODO: Why 16??
		c.token = token[:16]
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
	c.in = make(chan string)
	c.out = make(chan string)

	go c.handleConnections()

	return c, nil
}