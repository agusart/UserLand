package worker

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"userland/store/broker"
	"userland/store/postgres"
)

func UserLoginLog(
	ctx context.Context,
	msgBroker broker.BrokerInterface,
	logStore postgres.LogStoreInterface,
	endChan <-chan int) {
	c := msgBroker.GetConsumer()
	err := c.SubscribeTopics([]string{broker.BrokerLogTopicName}, nil)
	if err != nil {
		log.Err(err)
	}

	for {
		select {
		case <-endChan:
			return
		default:
			msg, err := c.ReadMessage(-1)
			if err != nil {
				continue
			}

			job := broker.UserLoginLogJob{}
			err = json.Unmarshal([]byte(msg.String()), &job)
			if err != nil {
				continue
			}

			logUserLogin := postgres.UserLog {
				UserId: job.UserId,
				SessionId: job.SessionId,
				RemoteIp : job.LoggedInIp,
				CreatedAt: job.LoggedInAt,
			}

			err = logStore.WriteUserLog(logUserLogin)
			if err != nil {
				log.Err(err)
			}

		}
	}

}
