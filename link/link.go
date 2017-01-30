package link

import (
	"os"
	"fmt"
)

import (
	lg "log"
	"io/ioutil"
)

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
		log = l
	}
}

type Link interface {
	// Init will initialize the link and setup the various configurations required for the link. This includes
	// loading the dslink.json and nodes.json files, if available, to populate the nodes.
	Init()
	// Start will start the link to establish connection to the broker.
	Start()
	Stop()
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
	logFile     string
	log	    bool
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
		if log == nil {
			log = lg.New(os.Stdout, "", lg.Lshortfile)
		}
	} else if log == nil {
		log = lg.New(ioutil.Discard, "", lg.Lshortfile)
	}

	if l.conf.autoInit {
		l.Init()
	}

	return &l
}

type link struct {
	conf Config
	cl   *httpClient
}

type dsJson struct {
	Config map[string]map[string]string `json:"configs"`
}

func (l *link) Init() {
	if l.conf.name[len(l.conf.name)-1] != '-' {
		l.conf.name += "-"
	}
	// TODO:
	// Load dslink.json
	// load nodes.json
}

func (l *link) Start() {
	// TODO
	var err error
	l.cl, err = dial(l.conf.broker, l.conf.name, l.conf.home, l.conf.token)
	if err != nil {
		panic(err)
	}
}

func (l *link) Stop() {
	l.cl.Close()
}

func (l *link) loadDsJson() {
	if l.conf.rootPath != "" {
		err := os.Chdir(l.conf.rootPath)
		if err != nil {
			fmt.Printf("Unable to load dslink.json, cannot find root path: %s\n", l.conf.rootPath)
			return
		}


	}
}