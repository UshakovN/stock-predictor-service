package rabbitmq

import "fmt"

const conn = "amqp://%s:%s@%s:%d/"

type Config struct {
  User     string `yaml:"user" required:"true"`
  Password string `yaml:"password" required:"true"`
  Host     string `yaml:"host" required:"true"`
  Port     int    `yaml:"port" required:"true"`
  QueueKey string `yaml:"queue_key" required:"true"`
}

func (c *Config) ConnectString() string {
  return fmt.Sprintf(conn,
    c.User,
    c.Password,
    c.Host,
    c.Port,
  )
}
