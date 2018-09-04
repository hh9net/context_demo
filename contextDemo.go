// Run this application.
// Try to change behavior and look what it will give you.
// Also you may catch deadlock in this application. Try to fix it by your own.
// Valid code without bug is in presentation.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var GlobalResource chan int
var kkk int
var mu sync.Mutex

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	//ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	wg := &sync.WaitGroup{}

	GlobalResource = NewResource(ctx, wg)

	wg.Add(1)
	go Worker(ctx, wg)
	wg.Add(1)
	go Worker(ctx, wg)

	time.Sleep(time.Second * 5)
	cancel()
	wg.Wait()
	fmt.Println("Main goroutine stoping")
	fmt.Println(kkk, "=================")
}

func Worker(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		fmt.Println("Worker was stoped")
		wg.Done()
	}()
	fmt.Println("Worker was started")

	childWG := &sync.WaitGroup{}

	for i := 0; i < 30; i++ {
		childWG.Add(1)
		go SubWorker(ctx, childWG, i)
	}

	// Will wait until main would not call cancel()
	<-ctx.Done()
	fmt.Println("Worker receiver ctx.Done signal. Waiting for SubWorkers")
	childWG.Wait()
}

func SubWorker(ctx context.Context, wg *sync.WaitGroup, index int) {
	defer func() {
		fmt.Printf("SubWorker %d was stoped\n", index)
		wg.Done()
	}()
	fmt.Printf("SubWorker #%d was started\n", index)

	for {
		select {
		case <-ctx.Done():
			sleepTime := rand.Intn(200)
			time.Sleep(time.Millisecond * time.Duration(sleepTime))
			fmt.Printf("SubWorker #%d Received ctx.Done signal. Stoping in %d Milliseconds\n", index, sleepTime)
			//time.Sleep(time.Millisecond * time.Duration(sleepTime))
			return
		default:
			if random, ok := <-GlobalResource; ok {
				mu.Lock()
				kkk++
				mu.Unlock()
				fmt.Printf("Tick from SubWorker #%d with value %d\n", index, random)
			}
		}
	}
}

func NewResource(ctx context.Context, wg *sync.WaitGroup) chan int {
	out := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Resource received ctx.Done signal")
				close(out)
				return
			case out <- rand.Intn(200):
			}
		}
	}()
	return out
}
