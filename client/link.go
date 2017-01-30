package client

func IsRequester(c *Config) {
	c.isRequester = true
}

func IsNotResponder(c *Config) {
	c.isResponder = false
}

type Link interface {
	Start()
}

// TODO: Provide some kind of config option for logger and logger level
type Config struct {
	isResponder bool
	isRequester bool
	broker string
	name string
	home string
	token string
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

	return &l
}

type link struct {
	conf Config
	cl httpClient
}

func (l *link) Start() {

}