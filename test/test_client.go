package main

import (
	"context"
	"log"

	"github.com/hopiconb/gofast-imp/invoicer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:8089", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()

	client := invoicer.NewInvoicerClient(conn)

	resp, err := client.Create(context.Background(), &invoicer.CreateRequest{
		Amount:    &invoicer.Amount{Amount: 12345, Currency: "USD"},
		From:      "Company A",
		To:        "Company B",
		VATnumber: "VAT1234567",
	})
	if err != nil {
		log.Fatalf("error calling Create: %v", err)
	}

	log.Printf("Received PDF: %v", resp.Pdf)
	log.Printf("Received DOCX: %v", resp.Docx)
}
