package broker

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"time"
)

const MsgBrokerLogTopicName = "user_login_log"

var broker MessageBrokerInterface

type UserLoginLogJob struct {
	LoggedInAt time.Time `json:"logged_in_at"`
	UserId uint `json:"user_id"`
	SessionId uint `json:"session_id"`
	LoggedInIp string `json:"logged_in_ip"`
}



type MessageBrokerInterface interface {
	SendLog(topic string, log UserLoginLogJob) error
	GetConsumer() *kafka.Consumer
	TearDown()
}


type Broker struct {
	producer *kafka.Producer
	consumer *kafka.Consumer
	logChan chan UserLoginLogJob

}

func (b Broker) SendLog(topic string, log UserLoginLogJob) error {
	jsonByte, err  := json.Marshal(log)
	if err != nil {
		return err
	}

	return b.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value: jsonByte,
	}, nil)
}

func (b Broker) GetConsumer() *kafka.Consumer {
	return b.consumer
}

func (b Broker) TearDown(){
	close(b.logChan)
}

func NewBroker (consumerCfg *kafka.ConfigMap, producerCfg *kafka.ConfigMap) (MessageBrokerInterface, error) {
	if broker != nil {
		return broker, nil
	}

	c, err := kafka.NewConsumer(consumerCfg)
	if err != nil {
		return nil, err
	}

	p, err := kafka.NewProducer(producerCfg)
	if err != nil {
		return nil, err
	}

	broker :=  Broker{
		producer: p,
		consumer: c,
	}

	broker.logChan = make(chan UserLoginLogJob)

	return broker, nil
}
