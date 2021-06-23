package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	pb "productinfo/client/ecommerce"
	"time"
)

const (
	address = "localhost:50051"
)

func main() {
	// Setting up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	//c := pb.NewProductInfoClient(conn)
	c := pb.NewOrderManagementClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// // Add Product
	//name := "Apple iPhone 11"
	//description := `Meet Apple iPhone 11. All-new dual-camera system with Ultra Wide and Night Mode.`
	//price := float32(1000.0)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//
	//r, err := c.AddProduct(ctx, &pb.Product{
	//	Name: name,
	//	Description: description,
	//	Price: price,
	//		})
	//if err != nil {
	//	log.Fatalf("Could not add product: %v", err)
	//}
	//log.Printf("Product ID: %s added successfully", r.Value)

	// // Get Product
	//product, err := c.GetProduct(ctx, &pb.ProductID{
	//	Value: r.Value,
	//})
	//if err != nil {
	//	log.Fatalf("Could not get product: %v", err)
	//}
	//log.Printf("Product: %v", product.String())

	// // Get Order
	//retrievedOrder, err := c.GetOrder(ctx, &wrappers.StringValue{Value: "106"})
	//log.Print("GetOrder Response -> : ", retrievedOrder)

	// Add Order
	order1 := pb.Order{
		Id: "101",
		Items:	[]string{"iPhone XS", "MacBook Pro"},
		Price: 2300.00,
		Description: "San Jose, CA",
	}
	res, _ := c.AddOrder(ctx, &order1)
	if res != nil {
		log.Print("AddOrder Response -> ", res.Value)
	}

}