package link

//import "log"

func IsRequester(c *Config) {
	c.isRequester = true
}

func IsNotResponder(c *Config) {
	c.isResponder = false
}

func AutoInit(c *Config) {
	c.autoInit = true
}

//var Log *log.Logger

//func Logger(l *log.Logger) func(c *Config) {
//	return func(c *Config) {
//		Log = l
//	}
//}

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
	autoInit bool
	broker string
	name string
	home string
	token string
	rootPath string
	//logFile string
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

	if l.conf.autoInit {
		l.Init()
	}

	return &l
}

type link struct {
	conf Config
	cl *httpClient
}

func (l *link) Init() {
	if l.conf.name[len(l.conf.name) - 1] != '-' {
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