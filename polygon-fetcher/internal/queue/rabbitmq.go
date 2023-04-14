package queue

import (
  "context"
  "fmt"
  "main/internal/domain"
  "main/internal/queue/rabbitmq"
  "main/pkg/utils"

  ampq "github.com/rabbitmq/amqp091-go"
  log "github.com/sirupsen/logrus"
)

const (
  queueDurable    = false
  queueAutoDelete = false
  queueExclusive  = false
  queueNoWait     = false

  publishExchange  = ""
  publishMandatory = false
  publishImmediate = false
)

type MediaServiceQueue interface {
  PublishMessage(message *domain.PutMessage) error
}

type mediaServiceQueue struct {
  ctx context.Context
  mq  rabbitmq.Client
  key string
}

func NewMediaServiceQueue(ctx context.Context, config *rabbitmq.Config) (MediaServiceQueue, error) {
  mq, err := rabbitmq.NewClient(config)
  if err != nil {
    return nil, fmt.Errorf("cannot create new queue client: %v", err)
  }
  log.Infof("init queue '%s' client for media service", config.QueueKey)
  var args ampq.Table // dummy args

  if _, err = mq.QueueDeclare(
    config.QueueKey,
    queueDurable,
    queueAutoDelete,
    queueExclusive,
    queueNoWait,
    args,
  ); err != nil {
    return nil, fmt.Errorf("cannot declare new queue: %v", err)
  }

  return &mediaServiceQueue{
    ctx: ctx,
    mq:  mq,
    key: config.QueueKey,
  }, nil
}

func (msq *mediaServiceQueue) PublishMessage(message *domain.PutMessage) error {
  if message == nil {
    return fmt.Errorf("message is a nil")
  }
  publishing, err := formPublishingFromMessage(message)
  if err != nil {
    return fmt.Errorf("cannot form publishing: %v", err)
  }

  if err = msq.mq.PublishWithContext(
    msq.ctx,
    publishExchange,
    msq.key,
    publishMandatory,
    publishImmediate,
    *publishing,
  ); err != nil {
    return fmt.Errorf("cannot publish message: %v", err)
  }
  log.Infof("message with name '%s' for section '%s' published to '%s' queue",
    message.MetaInfo.Name, message.MetaInfo.Section, msq.key)

  return nil
}

func formPublishingFromMessage(message *domain.PutMessage) (*ampq.Publishing, error) {
  headers, err := message.MetaInfo.MarshalMap()
  if err != nil {
    return nil, fmt.Errorf("cannot marshal message meta info as map: %v", err)
  }
  return &ampq.Publishing{
    Headers:   headers,
    Body:      message.Content,
    Timestamp: utils.NotTimeUTC(),
  }, nil
}
