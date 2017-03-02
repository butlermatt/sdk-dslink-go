package conn

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"github.com/butlermatt/dslink"
	"github.com/butlermatt/dslink/log"
	"github.com/butlermatt/dslink/nodes"
)

const dslinkJson = "dslink.json"

// Optional configuration functions which can be passed to NewLink

// IsRequester is an option for NewLink. It specifies that the link
// should also include requester functionality. By default requester
// is disabled.
func IsRequester(c *config) {
	c.isRequester = true
}

// IsNotResponder is an option for NewLink. It will disable the
// Responder functionality of the link. By default responder
// is always enabled.
func IsNotResponder(c *config) {
	c.isResponder = false
}

// AutoInit is an option for NewLink. This will indicate that
// the link should automatically try to call the Init method
// during the NewLink call. By default you must manually call
// the Init method on the Link.
func AutoInit(c *config) {
	c.autoInit = true
}

// LogLevel is an option for NewLink. It accepts a Level from
// log. This indicates what level logging should be enabled.
// By default LogLevel is set to DisabledLevel
func LogLevel(l log.Level) func(c *config) {
	return func(c *config) {
		c.logLevel = l
	}
}

// OnConnected is an option for NewLink. It accepts a callback
// with the signature of ConnectedCB. If supplied this will be
// called once the Link has successfully connected to an upstream
// broker.
func OnConnected(oc ConnectedCB) func(c *config) {
	return func(c *config) {
		c.oc = oc
	}
}

type config struct {
	isResponder bool
	isRequester bool
	autoInit    bool
	broker      string
	name        string
	home        string
	token       string
	rootPath    string
	keyPath	    string
	logFile     string
	logLevel    log.Level
	oc          ConnectedCB
}

// NewLink will create a new Link. The prefix is a require string which
// identifies this link with the upstream broker. The prefix should end
// with a hyphen (-) character. You may specify a variable number of
// configuration methods to help configure the link
func NewLink(prefix string, options ...func(*config)) *Link {
	var l Link

	// Set default options
	l.conf.isResponder = true
	l.conf.logLevel = log.DisabledLevel
	// Handle Options passed
	for _, option := range options {
		option(&l.conf)
	}

	l.conf.name = prefix

	// Handle Flags
	parseFlags(&l.conf)

	if l.conf.logLevel == log.DisabledLevel {
		log.SetLevel(log.DisabledLevel)
		log.SetOutput(nil)

	} else {
		log.SetLevel(l.conf.logLevel)
	}

	if l.conf.logFile != "" {
		//TODO: Deal with log files right.
	}

	if l.conf.autoInit {
		l.Init()
	}

	return &l
}

type ConnectedCB func(*Link)

type Link struct {
	conf  config
	cl    *httpClient
	pr    *nodes.Provider
	msgs  chan *dslink.Message
	resp  chan *dslink.Response
	reqs  chan *dslink.Request
	salt  string
	reqer *nodes.Requester
	init  bool
}

type dsJson struct {
	Config map[string]map[string]string `json:"configs"`
}

func (l *Link) Init() {
	if l.conf.name[len(l.conf.name)-1] != '-' {
		l.conf.name += "-"
	}

	if l.conf.isResponder {
		l.resp = make(chan *dslink.Response)
		l.pr = nodes.NewProvider(l.resp)
	}

	if l.conf.isRequester {
		l.reqs = make(chan *dslink.Request)
		l.reqer = nodes.NewRequester(l.reqs)
	}

	// TODO:
	// Load dslink.json
	l.loadDsJson()
	// load nodes.json
}

func (l *Link) Start() {
	var err error
	l.msgs = make(chan *dslink.Message)
	l.cl, err = dial(&l.conf, l.msgs)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case im := <-l.msgs:
			go l.handleMessage(im)
		case oresp := <-l.resp:
			m := &dslink.Message{}
			if oresp != nil {
				m.Resp = append(m.Resp, oresp)
				l.cl.out <- m
			}
		case oreq := <-l.reqs:
			m := &dslink.Message{}
			if oreq != nil {
				m.Reqs = append(m.Reqs, oreq)
				l.cl.out <- m
			}
		}
	}
}

func (l *Link) Stop() {
	// TODO: This isn't stopping the link though
	l.cl.Close()
}

func (l *Link) GetProvider() *nodes.Provider {
	return l.pr
}

func (l *Link) GetRequester() *nodes.Requester {
	return l.reqer
}

func (l *Link) handleMessage(m *dslink.Message) {
	var ackM *dslink.Message

	if len(m.Reqs) == 0 && len(m.Resp) == 0 && m.Salt == "" {
		// Ignore message.
		return
	}

	if l.reqer != nil {
		for _, resp := range m.Resp {
			l.reqer.HandleResponse(resp)
		}
	} else if len(m.Resp) > 0 {
		log.Debug.Println("Received responses when no requester active.")
	}

	ackM = &dslink.Message{Ack: m.Msg}
	if m.Salt != "" {
		l.salt = m.Salt
		// TODO: Handle actually using the reconnect salt (if the websocket ever drops)
		if !l.init {
			l.init = true
			if l.conf.oc != nil {
				go l.conf.oc(l)
			}
		}
	}

	for _, req := range m.Reqs {
		res := l.pr.HandleRequest(req)
		if res != nil {
			ackM.Resp = append(ackM.Resp, res)
		}
	}

	if ackM != nil {
		l.cl.out<- ackM
	}
}

func (l *Link) loadDsJson() {
	if l.conf.rootPath != "" {
		err := os.Chdir(l.conf.rootPath)
		if err != nil {
			log.Warn.Printf("Unable to load %s, cannot find root path: %s\n", dslinkJson, l.conf.rootPath)
			return
		}
	}
	d, err := ioutil.ReadFile(dslinkJson)
	if err != nil && os.IsNotExist(err) {
		log.Warn.Printf("Unable to find file: %q", dslinkJson)
		return
	} else if err != nil {
		log.Error.Printf("Unexpected error: %v", err)
		return
	}

	ds := &dsJson{}
	err = json.Unmarshal(d, ds)
	if err != nil {
		log.Error.Printf("Unable to Unmarshal data: %s\nError:%v\n", d, err)
		return
	}

	key := ds.Config["key"]
	if key!= nil {
		keyPath, ok := key["value"]
		if ok {
			l.conf.keyPath = keyPath
		}
	}


}
