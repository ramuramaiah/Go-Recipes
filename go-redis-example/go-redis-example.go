package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	redis "github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

func testSetGet() {
	err := rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := rdb.Get(ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := rdb.Get(ctx, "key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	// Output: key value
	// key2 does not exist
}

func testPubSub() {
	pubsub := rdb.Subscribe(ctx, "mychannel1")

	// Wait for confirmation that subscription is created before publishing anything.
	_, err := pubsub.Receive(ctx)
	if err != nil {
		panic(err)
	}

	// Go channel which receives messages.
	ch := pubsub.Channel()

	// Publish a message.
	err = rdb.Publish(ctx, "mychannel1", "hello").Err()
	if err != nil {
		panic(err)
	}

	time.AfterFunc(time.Second, func() {
		// When pubsub is closed channel is closed too.
		_ = pubsub.Close()
	})

	// Consume messages.
	for msg := range ch {
		fmt.Println(msg.Channel, msg.Payload)
	}
}

func testTransactions() {
	//Redis uses optimistic concurrency control for transactions

	const maxRetries = 1000

	// Increment transactionally increments key using GET and SET commands.
	increment := func(key string) error {
		// Transactional function.
		txf := func(tx *redis.Tx) error {
			// Get current value or zero.
			n, err := tx.Get(ctx, key).Int()
			if err != nil && err != redis.Nil {
				return err
			}

			// Actual opperation (local in optimistic lock).
			n++

			// Operation is committed only if the watched keys remain unchanged.
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, key, n, 0)
				return nil
			})
			return err
		}

		for i := 0; i < maxRetries; i++ {
			err := rdb.Watch(ctx, txf, key)
			if err == nil {
				// Success.
				return nil
			}
			if err == redis.TxFailedErr {
				// Optimistic lock lost. Retry.
				continue
			}
			// Return any other error.
			return err
		}

		return errors.New("increment reached maximum number of retries")
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := increment("counter3"); err != nil {
				fmt.Println("increment error:", err)
			}
		}()
	}
	wg.Wait()

	n, err := rdb.Get(ctx, "counter3").Int()
	fmt.Println("ended with", n, err)

}

func main() {
	//testSetGet()
	//testPubSub()
	testTransactions()
}
