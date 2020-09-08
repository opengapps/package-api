package packageapi

import (
	"errors"

	"github.com/spf13/viper"
)

// Option serves as the client configuration
type Option func(*application) error

// WithConfig provides viper config to the client
func WithConfig(cfg *viper.Viper) Option {
	return func(c *application) error {
		if cfg == nil {
			return errors.New("config is nil")
		}
		c.cfg = cfg
		return nil
	}
}

// WithStorage provides Storage to the client
func WithStorage(storage Storage) Option {
	return func(c *application) error {
		if storage == nil {
			return errors.New("storage is nil")
		}
		c.storage = storage
		return nil
	}
}
