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

  consumeName    = ""
  consumeAutoAck = false
  consumeExcl    = false
  consumeNoLocal = false
  consumeNoWait  = false

  ackNoMultiple = false
  ackRequeue    = true
)

type MessageHandler func(message *domain.PutMessage) error

type MediaServiceQueue interface {
  PublishMessage(message *domain.PutMessage) error
  ConsumeMessages(handler MessageHandler) error
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

func (msq *mediaServiceQueue) ConsumeMessages(handler MessageHandler) error {
  if handler == nil {
    return fmt.Errorf("message handler is a nil")
  }
  var args ampq.Table // dummy args

  consumerChan, err := msq.mq.Consume(
    msq.key,
    consumeName,
    consumeAutoAck,
    consumeExcl,
    consumeNoLocal,
    consumeNoWait,
    args,
  )
  if err != nil {
    return fmt.Errorf("cannot consume messages from '%s' queue", msq.key)
  }

  go func() {
    for delivery := range consumerChan {
      // form domain message
      message, err := formMessageFromDelivery(delivery)
      if err != nil {
        log.Errorf("malformed message in delivery: %v", err)
        continue
      }
      // format message description
      messageDesc := fmt.Sprintf("message with name '%s' for section '%s'",
        message.MetaInfo.Name, message.MetaInfo.Section)

      // handle queue messages with retries
      err = utils.DoWithRetry(func() error {
        if err := handler(message); err != nil {
          return fmt.Errorf("cannot handle %s. error: %v", messageDesc, err)
        }
        return nil
      })
      if err != nil {
        log.Errorf("handle failed for %s. error: %v", messageDesc, err)
        // send delivery negative acknowledgement to consumer and do requeue
        if err = delivery.Nack(ackNoMultiple, ackRequeue); err != nil {
          log.Errorf("cannot nack consumer about not handled delivery: %v", err)
        }
        continue
      }
      // send delivery acknowledgement to consumer
      if err = delivery.Ack(ackNoMultiple); err != nil {
        log.Errorf("cannot ack consumer about handled delivery: %v", err)
      }
      log.Infof("successfully handled: %s", messageDesc)
    }
  }()

  return nil
}

func formMessageFromDelivery(delivery ampq.Delivery) (*domain.PutMessage, error) {
  if len(delivery.Body) == 0 {
    return nil, fmt.Errorf("empty delivery body")
  }
  if len(delivery.Headers) == 0 {
    return nil, fmt.Errorf("empty delivery headers")
  }
  message := &domain.PutMessage{}

  err := message.MetaInfo.UnmarshalMap(delivery.Headers)
  if err != nil {
    return nil, fmt.Errorf("cannot unmarshal meta info from map: %v", err)
  }
  message.Content = delivery.Body

  return message, nil
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
