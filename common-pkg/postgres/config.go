package postgres

import "fmt"

const conn = "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s"

type Config struct {
	Host     string `yaml:"host" required:"true"`
	Port     int    `yaml:"port" required:"true"`
	User     string `yaml:"user" required:"true"`
	Password string `yaml:"password" required:"true"`
	DBName   string `yaml:"db_name" required:"true"`
	SSLMode  string `yaml:"ssl_mode" required:"true"`
}

func (c *Config) ConnectString() string {
	return fmt.Sprintf(conn,
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.DBName,
		c.SSLMode,
	)
}
