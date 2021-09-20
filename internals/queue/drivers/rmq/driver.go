package rmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/alash3al/exeq/internals/config"
	"github.com/alash3al/exeq/internals/queue"
	"github.com/go-redis/redis/v8"
)

type Redis struct {
	cfg     *config.QueueConfig
	redis   *redis.Client
	rmq     rmq.Connection
	queues  []rmq.Queue
	errChan chan error
}

func (r *Redis) Open(cfg *config.QueueConfig) (queue.Driver, error) {
	opts, err := redis.ParseURL(cfg.DSN)
	if err != nil {
		return nil, err
	}

	r.redis = redis.NewClient(opts)

	if err := r.redis.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	r.cfg = cfg
	r.errChan = make(chan error)

	r.rmq, err = rmq.OpenConnectionWithRedisClient("exeq", r.redis, r.errChan)
	if err != nil {
		return nil, err
	}

	r.queues = make([]rmq.Queue, r.cfg.RetryAttempts+1)

	r.queues[0], err = r.rmq.OpenQueue("exeq.queue.master")
	if err != nil {
		return nil, err
	}

	for i := 0; i < r.cfg.RetryAttempts; i++ {
		r.queues[i+1], err = r.rmq.OpenQueue(fmt.Sprintf("exeq.queue.push#%d", i))
		if err != nil {
			return nil, err
		}

		r.queues[i].SetPushQueue(r.queues[i+1])
	}

	cleaner := rmq.NewCleaner(r.rmq)

	go (func() {
		for range time.Tick(time.Second) {
			if _, err := cleaner.Clean(); err != nil {
				r.errChan <- err
				continue
			}

			if _, err := r.queues[0].PurgeRejected(); err != nil {
				r.errChan <- err
				continue
			}
		}
	})()

	return r, err
}

func (r *Redis) Enqueue(j *queue.Job) error {
	return r.queues[0].Publish(j.String())
}

func (r *Redis) Stats() ([]queue.JobStats, error) {
	return nil, nil
}

func (r *Redis) ListenAndConsume() error {
	dur, err := time.ParseDuration(r.cfg.PollDuration)
	if err != nil {
		return err
	}

	if err := r.queues[0].StartConsuming(int64(r.cfg.WorkersCount)*2, dur); err != nil {
		return err
	}

	for i := 0; i < r.cfg.WorkersCount; i++ {
		r.queues[0].AddConsumerFunc(fmt.Sprintf("workers#%0d", i+1), func(d rmq.Delivery) {
			var j queue.Job
			if err := json.Unmarshal([]byte(d.Payload()), &j); err != nil {
				if err != nil {
					r.errChan <- err
					if err := d.Reject(); err != nil {
						r.errChan <- err
					}

					return
				}
			}

			if err := j.Run(); err != nil {
				r.errChan <- err
				if err := d.Push(); err != nil {
					r.errChan <- err
				}

				return
			}

			if err := d.Ack(); err != nil {
				r.errChan <- err
			}
		})
	}

	select {}
}
