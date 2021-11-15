package worker

import (
	"encoding/json"
	"fmt"
	"userland/store/broker"
	"userland/store/postgres"
)

const KafkaWaitTimeout = -1

func UserLoginLog(
	msgBroker broker.MessageBrokerInterface,
	logStore postgres.LogStoreInterface,
	endChan <-chan int) {
	fmt.Println("worker starting")
	c := msgBroker.GetConsumer()
	err := c.SubscribeTopics([]string{broker.MsgBrokerLogTopicName}, nil)
	if err != nil {
		fmt.Println(err)
	}

	for {
		select {
		case <-endChan:
			return
		default:
			msg, err := c.ReadMessage(KafkaWaitTimeout)
			if err != nil {
				fmt.Println(err)
				continue
			}

			job := broker.UserLoginLogJob{}
			err = json.Unmarshal(msg.Value, &job)
			if err != nil {
				fmt.Println(err)
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
				fmt.Println(err)
				continue
			}

			fmt.Println("success save log")
		}
	}




}
