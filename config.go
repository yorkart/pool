package pool

import "errors"

type Config struct {
	InitialCap  int
	MaxCap      int
	IdleTimeout int
}

func (c *Config) copy() (*Config, error) {
	if err := c.valid(); err != nil {
		return nil, err
	}
	return &Config{
		InitialCap:  c.InitialCap,
		MaxCap:      c.MaxCap,
		IdleTimeout: c.IdleTimeout,
	}, nil
}

func (c *Config) valid() error {
	if c.InitialCap < 0 || c.MaxCap <= 0 || c.InitialCap > c.MaxCap {
		return errors.New("invalid capacity settings")
	}

	if c.IdleTimeout < 0 {
		return errors.New("invalid idel timeout settings")
	}

	return nil
}
