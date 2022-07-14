package gobrightbox

import (
	"net/url"
	"time"
)

// Server represents a Cloud Server
// https://api.gb1.brightbox.com/1.0/#server
type Server struct {
	Id        string
	Name      string
	Status    string
	Locked    bool
	Hostname  string
	Fqdn      string
	CreatedAt *time.Time `json:"created_at"`
	// DeletedAt is nil if the server has not yet been deleted
	DeletedAt         *time.Time `json:"deleted_at"`
	StartedAt         *time.Time `json:"started_at"`
	UserData          string     `json:"user_data"`
	CompatibilityMode bool       `json:"compatibility_mode"`
	DiskEncrypted     bool       `json:"disk_encrypted"`
	ServerConsole
	Account      Account
	Image        Image
	ServerType   ServerType `json:"server_type"`
	Zone         Zone
	Snapshots    []Image
	CloudIPs     []CloudIP `json:"cloud_ips"`
	Interfaces   []ServerInterface
	ServerGroups []ServerGroup `json:"server_groups"`
}

// ServerConsole is embedded into Server and contains the fields used in reponse
// to an ActivateConsoleForServer request.
type ServerConsole struct {
	ConsoleToken        string     `json:"console_token"`
	ConsoleUrl          string     `json:"console_url"`
	ConsoleTokenExpires *time.Time `json:"console_token_expires"`
}

// ServerOptions is used in conjunction with CreateServer and UpdateServer to
// create and update servers.
type ServerOptions struct {
	Id                string   `json:"-"`
	Image             string   `json:"image,omitempty"`
	Name              *string  `json:"name,omitempty"`
	ServerType        string   `json:"server_type,omitempty"`
	Zone              string   `json:"zone,omitempty"`
	UserData          *string  `json:"user_data,omitempty"`
	ServerGroups      []string `json:"server_groups,omitempty"`
	CompatibilityMode *bool    `json:"compatibility_mode,omitempty"`
	DiskEncrypted     *bool    `json:"disk_encrypted,omitempty"`
}

// ServerInterface represent a server's network interface(s)
type ServerInterface struct {
	Id          string
	MacAddress  string `json:"mac_address"`
	IPv4Address string `json:"ipv4_address"`
	IPv6Address string `json:"ipv6_address"`
}

// Servers retrieves a list of all servers
func (c *Client) Servers() ([]Server, error) {
	var servers []Server
	_, err := c.MakeApiRequest("GET", "/1.0/servers", nil, &servers)
	if err != nil {
		return nil, err
	}
	return servers, err
}

// Server retrieves a detailed view of one server
func (c *Client) Server(identifier string) (*Server, error) {
	server := new(Server)
	_, err := c.MakeApiRequest("GET", "/1.0/servers/"+identifier, nil, server)
	if err != nil {
		return nil, err
	}
	return server, err
}

// CreateServer creates a new server.
//
// It takes a ServerOptions struct which requires, at minimum, a valid Image
// identifier. Not all attributes can be specified at create time (such as Id,
// which is allocated for you)
func (c *Client) CreateServer(newServer *ServerOptions) (*Server, error) {
	server := new(Server)
	_, err := c.MakeApiRequest("POST", "/1.0/servers", newServer, &server)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// UpdateServer updates an existing server's attributes. Not all attributes can
// be changed after creation time (such as Image, ServerType and Zone).
//
// Specify the server you want to update using the ServerOptions Id field
func (c *Client) UpdateServer(updateServer *ServerOptions) (*Server, error) {
	server := new(Server)
	_, err := c.MakeApiRequest("PUT", "/1.0/servers/"+updateServer.Id, updateServer, &server)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// DestroyServer issues a request to destroy the server
func (c *Client) DestroyServer(identifier string) error {
	_, err := c.MakeApiRequest("DELETE", "/1.0/servers/"+identifier, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// StopServer issues a request to stop ("power off") an existing server
func (c *Client) StopServer(identifier string) error {
	_, err := c.MakeApiRequest("POST", "/1.0/servers/"+identifier+"/stop", nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// StartServer issues a request to start ("power on") an existing server
func (c *Client) StartServer(identifier string) error {
	_, err := c.MakeApiRequest("POST", "/1.0/servers/"+identifier+"/start", nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// RebootServer issues a request to reboot ("ctrl+alt+delete") an existing
// server
func (c *Client) RebootServer(identifier string) error {
	_, err := c.MakeApiRequest("POST", "/1.0/servers/"+identifier+"/reboot", nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// ResetServer issues a request to reset ("power cycle") an existing server
func (c *Client) ResetServer(identifier string) error {
	_, err := c.MakeApiRequest("POST", "/1.0/servers/"+identifier+"/reset", nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// ShutdownServer issues a request to shut down ("tap the power button") an
// existing server
func (c *Client) ShutdownServer(identifier string) error {
	_, err := c.MakeApiRequest("POST", "/1.0/servers/"+identifier+"/shutdown", nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// LockServer locks an existing server, preventing it's destruction without
// first unlocking. Deprecated, use LockResource instead.
func (c *Client) LockServer(identifier string) error {
	return c.LockResource(Server{Id: identifier})
}

// UnlockServer unlocks a previously locked existing server, allowing
// destruction again. Deprecated, use UnLockResource instead.
func (c *Client) UnlockServer(identifier string) error {
	return c.UnLockResource(Server{Id: identifier})
}

// SnapshotServer issues a request to snapshot the disk of an existing
// server. The snapshot is allocated an Image Id which is returned within an
// instance of Image.
func (c *Client) SnapshotServer(identifier string) (*Image, error) {
	res, err := c.MakeApiRequest("POST", "/1.0/servers/"+identifier+"/snapshot", nil, nil)
	if err != nil {
		return nil, err
	}
	imageID := getLinkRel(res.Header.Get("Link"), "img", "snapshot")
	if imageID != nil {
		img := new(Image)
		img.Id = *imageID
		return img, nil
	}
	return nil, nil
}

// ActivateConsoleForServer issues a request to enable the graphical console for
// an existing server. The temporarily allocated ConsoleUrl, ConsoleToken and
// ConsoleTokenExpires data are returned within an instance of Server.
func (c *Client) ActivateConsoleForServer(identifier string) (*Server, error) {
	server := new(Server)
	_, err := c.MakeApiRequest("POST", "/1.0/servers/"+identifier+"/activate_console", nil, server)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// FullConsoleUrl returns the console url for the server with the token in the
// query string.  Server needs a ConsoleUrl and ConsoleToken, retrieved using
// ActivateConsoleForServer
func (s *Server) FullConsoleUrl() string {
	if s.ConsoleUrl == "" || s.ConsoleToken == "" {
		return s.ConsoleUrl
	}
	u, err := url.Parse(s.ConsoleUrl)
	if u == nil || err != nil {
		return s.ConsoleUrl
	}
	values := u.Query()
	if values.Get("password") != "" {
		return s.ConsoleUrl
	}
	values.Set("password", s.ConsoleToken)
	u.RawQuery = values.Encode()
	return u.String()
}
