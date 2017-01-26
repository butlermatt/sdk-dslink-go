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

type client struct {
	dsId	 string
	msgId	 uint32
	reqId	 uint32
	keyMaker crypto.ECDH
	htClient *http.Client
	rawUrl   *url.URL
	wsClient *websocket.Conn
	cPriv 	 crypto.PrivateKey
	in       chan string
	out	 chan string
}

// Close will force the Websocket on the client to be closed.
func (c *client) Close() {
	if c.wsClient != nil {
		_ = c.wsClient.Close()
	}
}

func (c *client) getWsConfig() (*dsResp, error) {
	u, _ := url.Parse(c.rawUrl.String())
	q := u.Query()
	q.Add("dsId", c.dsId)
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

func (c *client) connectWs(config *dsResp) (*websocket.Conn, error) {
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
	u.RawQuery = q.Encode()
	u.Scheme = "ws"

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to Websocket at: %s\nError: %s", u.String(), err)
	}

	return conn, nil
}

func (c *client) handleConnections() {
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

	tick := time.Tick(30 * time.Second)
	c.wsClient.WriteMessage(websocket.TextMessage, []byte("{}"))
	for {
		select {
		case s := <-c.in:
			//TODO: Handle a received message
			fmt.Printf("Received message: %s\n", s)
		case o := <-c.out:
			c.wsClient.WriteMessage(websocket.TextMessage, []byte(o))
		case <- tick:
			c.msgId++
			m := fmt.Sprintf("{\"msg\": %d}", c.msgId)
			c.wsClient.WriteMessage(websocket.TextMessage, []byte(m))
		}
	}
}

// Dial will attempt to connect a link with the specified prefix to the specified address.
// Returns an error if connection handshake fails. Otherwise returns the connected client.
func Dial(addr, prefix string) (*client, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	c := &client{
		keyMaker: crypto.NewECDH(),
		htClient: &http.Client{Timeout: time.Second * 60},
		rawUrl: u,
	}

	// TODO: The keys should be managed outside of the client and
	// passed in as needed
	c.cPriv, err = c.keyMaker.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("Unable to generate key: %v", err)
	}
	c.dsId = c.cPriv.DsId(prefix)

	ret, err := c.getWsConfig()
	if err != nil {
		return nil, err
	}

	conn, err := c.connectWs(ret)
	if err != nil {
		return nil, err
	}

	c.wsClient = conn
	c.in = make(chan string)
	c.out = make(chan string)

	go c.handleConnections()

	return c, nil
}