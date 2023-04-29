package config

import (
	"errors"
	"flag"
	"net"
	"net/url"
	"os"
	"strconv"

	"github.com/Karzoug/loyalty_program/pkg/e"
)

const (
	defaultRunAddress           = "localhost:8081"
	defaultAccrualSystemAddress = "http://localhost:8080"
	defaultDatabaseURI          = ""
	defaultSecretKey            = ""
	defaultDebug                = false
)

type config struct {
	runAddress                 string
	accrualSystemAddressString string
	accrualSystemAddressURL    url.URL
	databaseURI                string
	secretKey                  string
	debug                      bool
}

// Read reads config values from (in order of priority): environment values, flags, defaults values.
func Read() (*config, error) {
	var c config

	c.readFlags()
	if err := c.readEnvs(); err != nil {
		return nil, e.Wrap("read environment values", err)
	}
	if err := c.validate(); err != nil {
		return nil, e.Wrap("no valid config values", err)
	}

	return &c, nil
}

// RunAddress is a rest server address (host:port).
func (c config) RunAddress() string {
	return c.runAddress
}

// AccrualSystemAddress is an accrual system address URL.
func (c config) AccrualSystemAddress() url.URL {
	return c.accrualSystemAddressURL
}

// DatabaseURI is a database connection string.
func (c config) DatabaseURI() string {
	return c.databaseURI
}

// SecretKey is a key to create a JWT signature.
func (c config) SecretKey() string {
	return c.secretKey
}

// IsDebugMode indicates whether the service is running in debug mode.
func (c config) IsDebugMode() bool {
	return c.debug
}

func (c *config) readFlags() {
	if flag.Parsed() {
		return
	}
	flag.StringVar(&c.runAddress, "a", defaultRunAddress, "rest server host and port")
	flag.StringVar(&c.accrualSystemAddressString, "r", defaultAccrualSystemAddress, "accrual system address (incl.scheme)")
	flag.StringVar(&c.databaseURI, "d", defaultDatabaseURI, "database connection string")
	flag.StringVar(&c.secretKey, "k", defaultSecretKey, "key to create a JWT signature")
	flag.BoolVar(&c.debug, "debug", defaultDebug, "debug mode")

	flag.Parse()
}

func (c *config) readEnvs() error {
	if runAddressString, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		c.runAddress = runAddressString
	}
	if accrualSystemAddressString, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		c.accrualSystemAddressString = accrualSystemAddressString
	}
	if databaseURIString, ok := os.LookupEnv("DATABASE_URI"); ok {
		c.databaseURI = databaseURIString
	}
	if secretKeyString, ok := os.LookupEnv("SECRET_KEY"); ok {
		c.secretKey = secretKeyString
	}
	if debugString, ok := os.LookupEnv("DEBUG"); ok {
		debugBool, err := strconv.ParseBool(debugString)
		if err != nil {
			return e.Wrap("parse variable 'DEBUG' error", err)
		}
		c.debug = debugBool
	}

	return nil
}

func (c *config) validate() error {
	_, _, err := net.SplitHostPort(c.runAddress)
	if err != nil {
		return errors.New("rest server host and port have wrong format")
	}

	u, err := url.ParseRequestURI(c.accrualSystemAddressString)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("accrual system address has wrong format")
	}
	c.accrualSystemAddressURL = *u

	if c.databaseURI == "" {
		return errors.New("database connection string must be non empty")
	}

	if c.secretKey == "" {
		return errors.New("secret key must be non empty")
	}

	return nil
}
