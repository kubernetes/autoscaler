package gsclient

import (
	"crypto/tls"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultMaxNumberOfRetries     = 5
	defaultDelayIntervalMilliSecs = 1000
	version                       = "3.11.1"
	defaultAPIURL                 = "https://api.gridscale.io"
	resourceActiveStatus          = "active"
	requestDoneStatus             = "done"
	requestFailStatus             = "failed"
	bodyType                      = "application/json"
)

// Config holds config for client.
type Config struct {
	apiURL             string
	userUUID           string
	apiToken           string
	userAgent          string
	httpHeaders        map[string]string
	sync               bool
	httpClient         *http.Client
	delayInterval      time.Duration
	maxNumberOfRetries int
}

var logger = logrus.Logger{
	Out:   os.Stderr,
	Level: logrus.InfoLevel,
	Formatter: &logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: false,
	},
}

// NewConfiguration creates a new config.
//
// - Parameters:
//   - apiURL string: base URL of API.
//   - uuid string: UUID of user.
//   - token string: API token.
//   - debugMode bool: true => run client in debug mode.
//   - sync bool: true => client is in synchronous mode. The client will block until Create/Update/Delete processes.
//     are completely finished. It is safer to set this parameter to `true`.
//   - delayIntervalMilliSecs int: delay (in milliseconds) between requests when checking request (or retry 503 error code).
//   - maxNumberOfRetries int: number of retries when server returns 503 error code.
func NewConfiguration(apiURL string, uuid string, token string, debugMode, sync bool,
	delayIntervalMilliSecs, maxNumberOfRetries int) *Config {
	if debugMode {
		logger.Level = logrus.DebugLevel
	}

	cfg := &Config{
		apiURL:    apiURL,
		userUUID:  uuid,
		apiToken:  token,
		userAgent: "gsclient-go/" + version + " (" + runtime.GOOS + ")",
		sync:      sync,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			},
		},
		delayInterval:      time.Duration(delayIntervalMilliSecs) * time.Millisecond,
		maxNumberOfRetries: maxNumberOfRetries,
	}
	return cfg
}

// DefaultConfiguration creates a default configuration.
func DefaultConfiguration(uuid string, token string) *Config {
	cfg := &Config{
		apiURL:    defaultAPIURL,
		userUUID:  uuid,
		apiToken:  token,
		userAgent: "gsclient-go/" + version + " (" + runtime.GOOS + ")",
		sync:      true,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			},
		},
		delayInterval:      time.Duration(defaultDelayIntervalMilliSecs) * time.Millisecond,
		maxNumberOfRetries: defaultMaxNumberOfRetries,
	}
	return cfg
}

// SetLogLevel manually sets log level.
// Read more: https://github.com/sirupsen/logrus#level-logging
func SetLogLevel(level logrus.Level) {
	logger.Level = level
}
