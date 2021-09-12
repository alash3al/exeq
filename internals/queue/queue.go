package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/alash3al/exeq/internals/config"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
)

type Queue struct {
	redisConn      *redis.Client
	rmqConn        rmq.Connection
	masterQueue    rmq.Queue
	configs        *config.Config
	pushQueues     []rmq.Queue
	allQueuesNames []string
}

func New(configs *config.Config) (*Queue, error) {
	opts, err := redis.ParseURL(configs.Queue.DSN)
	if err != nil {
		return nil, err
	}

	redisConn := redis.NewClient(opts)

	if err := redisConn.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	errChan := make(chan error)

	rmqConn, err := rmq.OpenConnectionWithRedisClient("exeq", redisConn, errChan)
	if err != nil {
		return nil, err
	}

	masterQueueName := "exeq.master"
	masterQueue, err := rmqConn.OpenQueue(masterQueueName)
	if err != nil {
		return nil, err
	}

	allQueuesNames := []string{masterQueueName}

	masterQueueCleaner := rmq.NewCleaner(rmqConn)

	go (func() {
		for {
			time.Sleep(time.Minute * 10)
			log.Println("Purging rejected")
			masterQueueCleaner.Clean()
		}
	})()

	pushQueues := []rmq.Queue{}

	for i := 0; i < configs.Queue.RetryAttempts; i++ {
		pushQueueName := fmt.Sprintf("exeq.push.%d", i)
		pushQueue, err := rmqConn.OpenQueue(pushQueueName)
		if err != nil {
			return nil, err
		}

		allQueuesNames = append(allQueuesNames, pushQueueName)
		pushQueues = append(pushQueues, pushQueue)

		if i < 1 {
			masterQueue.SetPushQueue(pushQueue)
		} else {
			pushQueues[i-1].SetPushQueue(pushQueue)
		}
	}

	return &Queue{
		redisConn:      redisConn,
		rmqConn:        rmqConn,
		masterQueue:    masterQueue,
		pushQueues:     pushQueues,
		configs:        configs,
		allQueuesNames: allQueuesNames,
	}, nil
}

func (q *Queue) Enqueue(j *Job) (string, error) {
	j.ID = xid.New().String()

	if err := q.masterQueue.Publish(j.String()); err != nil {
		return "", err
	}

	return j.ID, nil
}

func (q *Queue) ListenAndConsume() error {
	pollDuration, err := time.ParseDuration(q.configs.Queue.PollDuration)
	if err != nil {
		return err
	}

	if err := q.masterQueue.StartConsuming(
		int64(q.configs.Queue.WorkersCount)*2,
		pollDuration,
	); err != nil {
		return err
	}

	for i := 0; i < q.configs.Queue.WorkersCount; i++ {
		workerName := fmt.Sprintf("worker-%d", i)
		log.Println("starting worker ", workerName)
		_, err := q.masterQueue.AddConsumerFunc(workerName, func(d rmq.Delivery) {
			var job Job

			if err := json.Unmarshal([]byte(d.Payload()), &job); err != nil {
				log.Println(err)

				// it is an invalid payload we must reject it
				if err := d.Reject(); err != nil {
					log.Println(err)
				}
				return
			}

			if err := job.Run(); err != nil {
				log.Println(err)

				// it failed, we should retry it again later
				if err := d.Push(); err != nil {
					log.Println(err)
				}
				return
			}

			if err := d.Ack(); err != nil {
				log.Println(err)
			}
		})

		if err != nil {
			return err
		}
	}

	select {}
}
