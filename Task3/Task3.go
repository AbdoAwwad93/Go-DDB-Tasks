package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		sleepTime := time.Duration(rand.Intn(1000)) * time.Millisecond
		time.Sleep(sleepTime)

		fmt.Printf("Worker %d processed job %d\n", id, job)
	}
}

func main() {
	numWorkers := 3
	numJobs := 6
	jobs := make(chan int)
	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, &wg)
	}

	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}

	close(jobs)

	wg.Wait()

	fmt.Println("All jobs processed.")
}
