package config

import (
	"errors"
	"flag"
	"os"
	"strconv"

	"github.com/Karzoug/loyalty_program/pkg/e"
)

const (
	defaultRunAddress           = ":8080"
	defaultAccrualSystemAddress = ""
	defaultDatabaseURI          = ""
	defaultSecretKey            = ""
	defaultDebug                = false
	defaultBrokerURI            = "redis://localhost:6379/"
)

type config struct {
	runAddress           string
	accrualSystemAddress string
	brokerURI            string
	databaseURI          string
	secretKey            string
	debug                bool
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

// RunAddress is a rest server address and port.
func (c config) RunAddress() string {
	return c.runAddress
}

// AccrualSystemAddress is an accrual system address and port.
func (c config) AccrualSystemAddress() string {
	return c.accrualSystemAddress
}

// BrokerURI is a broker connection string.
func (c config) BrokerURI() string {
	return c.brokerURI
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
	flag.StringVar(&c.runAddress, "a", defaultRunAddress, "rest server address and port")
	flag.StringVar(&c.accrualSystemAddress, "r", defaultAccrualSystemAddress, "accrual system address and port")
	flag.StringVar(&c.brokerURI, "b", defaultBrokerURI, "message broker connection string")
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
		c.accrualSystemAddress = accrualSystemAddressString
	}
	if brokerURIString, ok := os.LookupEnv("BROKER_URI"); ok {
		c.brokerURI = brokerURIString
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
	if c.runAddress == "" {
		return errors.New("rest server address must be non empty")
	}

	if c.accrualSystemAddress == "" {
		return errors.New("accrual system address must be non empty")
	}

	if c.brokerURI == "" {
		return errors.New("message broker connection string must be non empty")
	}

	if c.databaseURI == "" {
		return errors.New("database connection string must be non empty")
	}

	if c.secretKey == "" {
		return errors.New("secret key must be non empty")
	}

	return nil
}
