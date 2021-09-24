package rmq

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/alash3al/exeq/internals/config"
	"github.com/alash3al/exeq/internals/queue"
	"github.com/go-redis/redis/v8"
)

const (
	redisHashMapJobsHistory      = "exeq/jobs/history"
	redisCounterPendingCounter   = "exeq/counters/pending"
	redisCounterRunningCounter   = "exeq/counters/running"
	redisCounterSucceededCounter = "exeq/counters/succeeded"
	redisCounterRetriesCounter   = "exeq/counters/retries"
	redisCounterFailedCounter    = "exeq/counters/failed"
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
		for range time.Tick(time.Minute) {
			// clean the underlying queue
			if _, err := cleaner.Clean(); err != nil {
				r.errChan <- err
			}

			// clean the underlying rejected queue
			if _, err := r.queues[0].PurgeRejected(); err != nil {
				r.errChan <- err
			}

			// clean the old jobs in the history based on the retention period
			{
				retentionPeriod, err := time.ParseDuration(r.cfg.History.RetentionPeriod)
				if err != nil {
					r.errChan <- err
				} else if err == nil && retentionPeriod > time.Second {
					for _, item := range r.redis.HGetAll(context.Background(), redisHashMapJobsHistory).Val() {
						var j queue.Job

						if err := json.Unmarshal([]byte(item), &j); err != nil {
							r.errChan <- err
						}

						if time.Since(j.EnqueuedAt) >= retentionPeriod {
							fmt.Printf("removing %s because sicne is %d \n", j.ID, time.Since(j.EnqueuedAt))
							if err := r.redis.HDel(context.Background(), redisHashMapJobsHistory, j.ID).Err(); err != nil {
								r.errChan <- err
							}
						}
					}
				}
			}
		}
	})()

	return r, err
}

func (r *Redis) Enqueue(j *queue.Job) error {
	j.EnqueuedAt = time.Now()

	if err := r.syncJob(j); err != nil {
		return err
	}

	r.incrPending(1)

	return r.queues[0].Publish(j.String())
}

func (r *Redis) Err() <-chan error {
	return r.errChan
}

func (r *Redis) Stats() (queue.Stats, error) {
	getCounter := func(counter string) int64 {
		i, _ := r.redis.Get(context.Background(), counter).Int64()
		return int64(math.Max(0, float64(i)))
	}

	return queue.Stats{
		Pending:   getCounter(redisCounterPendingCounter),
		Running:   getCounter(redisCounterRunningCounter),
		Succeeded: getCounter(redisCounterSucceededCounter),
		Failed:    getCounter(redisCounterFailedCounter),
		Retries:   getCounter(redisCounterRetriesCounter),
	}, nil
}

func (r *Redis) List() ([]queue.Job, error) {
	all := r.redis.HGetAll(context.Background(), redisHashMapJobsHistory).Val()
	result := []queue.Job{}

	for _, val := range all {
		var j queue.Job

		if err := json.Unmarshal([]byte(val), &j); err != nil {
			return nil, err
		}

		result = append(result, j)
	}

	return result, nil
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
		r.queues[0].AddConsumerFunc(fmt.Sprintf("workers/master/#%0d", i+1), r.worker)
	}

	select {}
}

func (r *Redis) worker(d rmq.Delivery) {
	r.incrPending(-1)
	r.incrRunning(1)
	defer func() {
		r.incrRunning(-1)
	}()

	var j queue.Job
	if err := json.Unmarshal([]byte(d.Payload()), &j); err != nil {
		if err != nil {
			r.errChan <- err
			if err := d.Reject(); err != nil {
				r.errChan <- fmt.Errorf(
					"rejecting the job %s due to => %s",
					d.Payload(),
					err.Error(),
				)
			}

			r.incrFailed(1)

			return
		}
	}

	j.StartedAt = time.Now()

	r.syncJob(&j)

	if err := j.Run(); err != nil {
		j.FinishedAt = time.Now()
		j.Error = err.Error()

		r.errChan <- err

		if err := d.Reject(); err != nil {
			r.errChan <- err
		}

		r.incrFailed(1)
		r.syncJob(&j)

		if r.cfg.RetryAttempts < 1 || j.RetryAttempts >= int64(r.cfg.RetryAttempts) {
			return
		}

		j.ID += fmt.Sprintf("-retry-%d", j.RetryAttempts)
		j.RetryAttempts++

		if err := r.Enqueue(&j); err != nil {
			r.errChan <- err
		} else {
			r.incrRetries(1)
		}

		return
	}

	if err := d.Ack(); err != nil {
		j.FinishedAt = time.Now()
		j.Error = err.Error()

		r.syncJob(&j)

		r.errChan <- err

		return
	}

	j.FinishedAt = time.Now()

	r.syncJob(&j)

	r.incrSucceeded(1)
}

func (r *Redis) syncJob(j *queue.Job) error {
	return r.redis.HSet(context.Background(), redisHashMapJobsHistory, j.ID, j.String()).Err()
}

func (r *Redis) incrPending(delta int64) {
	r.incrCounter(redisCounterPendingCounter, delta)
}

func (r *Redis) incrRunning(delta int64) {
	r.incrCounter(redisCounterRunningCounter, delta)
}

func (r *Redis) incrSucceeded(delta int64) {
	r.incrCounter(redisCounterSucceededCounter, delta)
}

func (r *Redis) incrFailed(delta int64) {
	r.incrCounter(redisCounterFailedCounter, delta)
}

func (r *Redis) incrRetries(delta int64) {
	r.incrCounter(redisCounterRetriesCounter, delta)
}

func (r *Redis) incrCounter(counter string, delta int64) {
	if _, err := r.redis.IncrBy(context.Background(), counter, delta).Result(); err != nil {
		r.errChan <- err
	}
}
