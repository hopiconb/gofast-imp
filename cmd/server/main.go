package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/hopiconb/gofast-imp/invoicer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type myInvoicerServer struct {
	invoicer.UnimplementedInvoicerServer
	cassSess *gocql.Session
}

// --- Prometheus Metrics ---
var (
	invoiceCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "invoices_created_total",
			Help: "Total number of invoices created",
		},
		[]string{"user"},
	)

	cassandraLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "cassandra_query_duration_seconds",
			Help:    "Histogram of response time for Cassandra queries",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func (s *myInvoicerServer) Create(ctx context.Context, req *invoicer.CreateRequest) (*invoicer.CreateResponse, error) {
	var clientIP string
	if p, ok := peer.FromContext(ctx); ok {
		clientIP = p.Addr.String()
	}

	log.Printf("[📦] Client \"%s\" (IP: %s) sent an invoice request of %d %s to \"%s\" (VAT: %s)",
		req.From, clientIP, req.Amount.Amount, req.Amount.Currency, req.To, req.VATnumber)

	// Batch insert into Cassandra
	batch := gocql.NewBatch(gocql.LoggedBatch)

	// Create a batch for multiple invoices (you can modify the batch size depending on your load)
	batch.Query(`
		INSERT INTO invoicedata.invoices (id, username, from_company, to_company, vat, amount, currency, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		gocql.TimeUUID(), req.User, req.From, req.To, req.VATnumber, req.Amount.Amount, req.Amount.Currency, time.Now(),
	)

	// Add more queries if needed (simulating multiple inserts)
	// batch.Query( ... )

	start := time.Now()
	err := s.cassSess.ExecuteBatch(batch)
	cassandraLatency.Observe(time.Since(start).Seconds())

	if err != nil {
		return nil, fmt.Errorf("failed to insert invoice into Cassandra: %v", err)
	}

	invoiceCounter.WithLabelValues(req.User).Inc()
	log.Printf("🧾 Invoice request saved to Cassandra for user '%s'", req.User)

	return &invoicer.CreateResponse{
		Pdf:  []byte(req.From),
		Docx: []byte("test"),
	}, nil
}

func main() {
	// Register Prometheus metrics
	prometheus.MustRegister(invoiceCounter)
	prometheus.MustRegister(cassandraLatency)

	// Serve /metrics
	go func() {
		log.Println("📊 Prometheus metrics server listening on :2112/metrics")
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()

	// Connect to Cassandra
	var cassSess *gocql.Session
	for i := 0; i < 20; i++ {
		cluster := gocql.NewCluster("cassandra")
		cluster.Port = 9042
		cluster.Consistency = gocql.Quorum
		cluster.Timeout = 5 * time.Second

		var err error
		cassSess, err = cluster.CreateSession()
		if err == nil {
			break
		}
		log.Println("⏳ Waiting for Cassandra...")
		time.Sleep(3 * time.Second)
	}
	if cassSess == nil {
		log.Fatalf("❌ Could not connect to Cassandra")
	}
	defer cassSess.Close()

	err := cassSess.Query(`
		CREATE KEYSPACE IF NOT EXISTS invoicedata WITH replication = {
			'class': 'SimpleStrategy', 'replication_factor': 1
		}`).Exec()
	if err != nil {
		log.Fatalf("❌ Failed to create Cassandra keyspace: %v", err)
	}

	err = cassSess.Query(`
		CREATE TABLE IF NOT EXISTS invoicedata.invoices (
			id UUID PRIMARY KEY,
			username TEXT,
			from_company TEXT,
			to_company TEXT,
			vat TEXT,
			amount INT,
			currency TEXT,
			created_at TIMESTAMP
		)`).Exec()
	if err != nil {
		log.Fatalf("❌ Failed to create Cassandra table: %v", err)
	}

	// Start gRPC server
	lis, err := net.Listen("tcp", ":8089")
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	log.Println("🚀 gRPC server listening on port 8089...")

	serverRegistrar := grpc.NewServer()
	invoicer.RegisterInvoicerServer(serverRegistrar, &myInvoicerServer{
		cassSess: cassSess, // Pass cassSess here
	})
	log.Println("🔗 Invoicer service registered")

	if err := serverRegistrar.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
