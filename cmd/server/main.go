package main

import (
	"context"
	"log"
	"net"

	"github.com/hopiconb/gofast-imp/invoicer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type myInvoicerServer struct {
	invoicer.UnimplementedInvoicerServer
}

func (s myInvoicerServer) Create(ctx context.Context, req *invoicer.CreateRequest) (*invoicer.CreateResponse, error) {
	// Get the peer (client) info from the context
	var clientIP string
	if p, ok := peer.FromContext(ctx); ok {
		clientIP = p.Addr.String()
	}

	log.Printf("[📦] Client \"%s\" (IP: %s) sent an invoice request of %d %s to \"%s\" (VAT: %s)",
		req.From, clientIP, req.Amount.Amount, req.Amount.Currency, req.To, req.VATnumber)

	return &invoicer.CreateResponse{
		Pdf:  []byte(req.From),
		Docx: []byte("test"),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8089")
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	log.Println("🚀 gRPC server listening on port 8089...")

	serverRegistrar := grpc.NewServer()
	service := &myInvoicerServer{}

	invoicer.RegisterInvoicerServer(serverRegistrar, service)
	log.Println("🔗 Invoicer service registered")

	err = serverRegistrar.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
