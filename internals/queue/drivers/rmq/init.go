package rmq

import "github.com/alash3al/exeq/internals/queue"

func init() {
	queue.Register("rmq", &Redis{})
}
