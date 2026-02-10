package playcamp

import (
	"log"
	"net/http"
	"time"
)

// Option configures a Client or Server.
type Option func(*config)

type config struct {
	environment Environment
	baseURL     string
	timeout     time.Duration
	isTest      bool
	maxRetries  int
	httpClient  *http.Client
	debug       *DebugOptions
}

// DebugOptions configures debug logging for HTTP requests and responses.
type DebugOptions struct {
	Enabled         bool
	Logger          func(format string, args ...any)
	LogRequestBody  bool
	LogResponseBody bool
	LogHeaders      bool
}

func defaultConfig() *config {
	return &config{
		environment: EnvironmentLive,
		timeout:     30 * time.Second,
		isTest:      false,
		maxRetries:  3,
	}
}

func defaultDebugLogger() func(string, ...any) {
	return func(format string, args ...any) {
		log.Printf("[playcamp] "+format, args...)
	}
}

// WithEnvironment sets the target environment.
func WithEnvironment(env Environment) Option {
	return func(c *config) {
		c.environment = env
	}
}

// WithBaseURL overrides the environment-derived base URL.
func WithBaseURL(url string) Option {
	return func(c *config) {
		c.baseURL = url
	}
}

// WithTimeout sets the HTTP request timeout. Default is 30 seconds.
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		c.timeout = d
	}
}

// WithTestMode enables test mode, which adds a test query parameter to requests.
func WithTestMode(enabled bool) Option {
	return func(c *config) {
		c.isTest = enabled
	}
}

// WithMaxRetries sets the maximum number of retry attempts. Default is 3.
func WithMaxRetries(n int) Option {
	return func(c *config) {
		c.maxRetries = n
	}
}

// WithHTTPClient sets a custom http.Client for making requests.
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		c.httpClient = client
	}
}

// WithDebug enables debug logging with the given options.
func WithDebug(opts DebugOptions) Option {
	return func(c *config) {
		if opts.Logger == nil {
			opts.Logger = defaultDebugLogger()
		}
		c.debug = &opts
	}
}
