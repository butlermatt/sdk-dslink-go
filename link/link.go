package link

import (
	"encoding/json"
	"io/ioutil"
	"os"
	lg "log"
	"github.com/butlermatt/dslink"
	"github.com/butlermatt/dslink/nodes"
)

const dslinkJson = "dslink.json"

func IsRequester(c *Config) {
	c.isRequester = true
}

func IsNotResponder(c *Config) {
	c.isResponder = false
}

func AutoInit(c *Config) {
	c.autoInit = true
}

var log *lg.Logger

func Logger(l *lg.Logger) func(c *Config) {
	return func(c *Config) {
		dslink.Log = l
	}
}

func Provider(p dslink.Provider) func(c *Config) {
	return func(c *Config) {
		c.provider = p
	}
}

func OnConnected(oc ConnectedCB) func(c *Config) {
	return func(c *Config) {
		c.oc = oc
	}
}

type Link interface {
	// Init will initialize the link and setup the various configurations required for the link. This includes
	// loading the dslink.json and nodes.json files, if available, to populate the nodes.
	Init()
	// Start will start the link to establish connection to the broker.
	Start()
	Stop()
	GetProvider() dslink.Provider
}

// TODO: Provide some kind of config option for logger and logger level
type Config struct {
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
	log         bool
	provider    dslink.Provider
	oc          ConnectedCB
}

func NewLink(prefix string, options ...func(*Config)) Link {
	var l link

	// Handle Options passed
	l.conf.isResponder = true
	for _, option := range options {
		option(&l.conf)
	}

	l.conf.name = prefix

	// Handle Flags
	parseFlags(&l.conf)

	if l.conf.log {
		if dslink.Log == nil {
			dslink.Log = lg.New(os.Stdout, "[DSA] ",0 /*lg.Lshortfile*/)
		}
	} else if dslink.Log == nil {
		dslink.Log = lg.New(ioutil.Discard, "[DSA] ", lg.Lshortfile)
	}
	log = dslink.Log

	if l.conf.autoInit {
		l.Init()
	}

	return &l
}

type ConnectedCB func(Link)

type link struct {
	conf  Config
	cl    *httpClient
	pr    dslink.Provider
	msgs  chan *dslink.Message
	resp  chan *dslink.Response
	salt  string
}

type dsJson struct {
	Config map[string]map[string]string `json:"configs"`
}

func (l *link) Init() {
	l.resp = make(chan *dslink.Response)
	if l.conf.name[len(l.conf.name)-1] != '-' {
		l.conf.name += "-"
	}

	if l.conf.provider != nil {
		l.pr = l.conf.provider
		l.conf.provider = nil
	} else {
		l.pr = nodes.NewProvider(l.resp)
	}
	// TODO:
	// Load dslink.json
	l.loadDsJson()
	// load nodes.json
}

func (l *link) Start() {
	var err error
	l.msgs = make(chan *dslink.Message)
	l.cl, err = dial(&l.conf, l.msgs)
	if err != nil {
		panic(err)
	}

	if l.conf.oc != nil {
		go l.conf.oc(l)
	}

	for {
		select {
		case im := <-l.msgs:
			go l.handleMessage(im)
		case or := <-l.resp:
			m := &dslink.Message{}
			if or != nil {
				m.Resp = append(m.Resp, or)
				l.cl.out <- m
			}
		}
	}
}

func (l *link) Stop() {
	l.cl.Close()
}

func (l *link) GetProvider() dslink.Provider {
	return l.pr
}

func (l *link) handleMessage(m *dslink.Message) {
	var r *dslink.Message

	if len(m.Reqs) == 0 && len(m.Resp) == 0 && m.Salt == "" {
		// Ignore message.
		return
	}

	r = &dslink.Message{Ack: m.Msg}
	if m.Salt != "" {
		l.salt = m.Salt
	}

	for _, req := range m.Reqs {
		res := l.pr.HandleRequest(req)
		if res != nil {
			r.Resp = append(r.Resp, res)
		}
	}

	if r != nil {
		l.cl.out<- r
	}
}

func (l *link) loadDsJson() {
	if l.conf.rootPath != "" {
		err := os.Chdir(l.conf.rootPath)
		if err != nil {
			log.Printf("Unable to load %s, cannot find root path: %s\n", dslinkJson, l.conf.rootPath)
			return
		}
	}
	d, err := ioutil.ReadFile(dslinkJson)
	if err != nil {
		log.Printf("Unable to open file: %s\nError: %v", dslinkJson, err)
		return
	}

	ds := &dsJson{}
	err = json.Unmarshal(d, ds)
	if err != nil {
		log.Printf("Unable to Unmarshal data: %s\nError:%v\n", d, err)
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
