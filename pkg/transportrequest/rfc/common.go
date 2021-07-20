package rfc

// Connection Everything we need for connecting to the ABAP system
type Connection struct {
	// The endpoint in for form <protocol>://<host>:<port>, no path
	Endpoint string
	// The ABAP client, like e.g. "001"
	Client string
	// The ABAP instance, like e.g. "DEV",  "QA"
	Instance string
	User     string
	Password string
}
