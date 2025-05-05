package main

import (
	"context"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/hopiconb/gofast-imp/invoicer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var users = []string{"cosmin", "ana", "maria", "andrei", "david", "elena", "george", "ioana"}
var companies = []string{"Company A", "Company B", "Company C", "Company D", "Company E"}
var vatNumbers = []string{"VAT123", "VAT456", "VAT789", "VAT321", "VAT654", "VAT987"}

func randomFrom(list []string) string {
	return list[rand.Intn(len(list))]
}

func sendRequest(client invoicer.InvoicerClient, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	req := &invoicer.CreateRequest{
		User:      randomFrom(users),
		From:      randomFrom(companies),
		To:        randomFrom(companies),
		VATnumber: randomFrom(vatNumbers),
		Amount:    &invoicer.Amount{Amount: int64(rand.Intn(10000) + 1), Currency: "USD"},
	}

	_, err := client.Create(ctx, req)
	if err != nil {
		log.Printf("❌ error: %v", err)
	}
}

func workerPool(client invoicer.InvoicerClient, jobs <-chan int, wg *sync.WaitGroup, burstSize int) {
	ctx := context.Background()
	for range jobs {
		for i := 0; i < burstSize; i++ {
			wg.Add(1)
			go sendRequest(client, wg, ctx)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	conn, err := grpc.Dial("localhost:8089",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("❌ could not connect: %v", err)
	}
	defer conn.Close()

	client := invoicer.NewInvoicerClient(conn)
	rand.Seed(time.Now().UnixNano())

	const (
		totalBursts = 50000 // Number of batches sent to the server
		concurrency = 300   // Worker goroutines
		burstSize   = 7     // Requests per burst (per job)
	)

	jobs := make(chan int, concurrency)
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < concurrency; i++ {
		go workerPool(client, jobs, &wg, burstSize)
	}

	for i := 0; i < totalBursts; i++ {
		jobs <- i
		if (i+1)%500 == 0 {
			log.Printf("🚀 Sent %d bursts (%d total requests)...", i+1, (i+1)*burstSize)
		}
	}

	close(jobs)
	wg.Wait()

	log.Println("🏁 Stress test completed.")
}
