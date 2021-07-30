package main

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	pb "productinfo/service/ecommerce"
)

const (
	port = ":50051"
	orderBatchSize = 3
)

var orderMap = make(map[string]pb.Order)

// server is used to implement ecommerce/product_info.
type server struct {
	productMap map[string]*pb.Product
	orderMap map[string]*pb.Order
}

// AddProduct implements ecommerce.AddProduct
func (s *server) AddProduct(ctx context.Context, in *pb.Product) (*pb.ProductID, error) {
	out, err := uuid.NewV4()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error while generating Product ID", err)
	}
	in.Id = out.String()
	if s.productMap == nil {
		s.productMap = make(map[string]*pb.Product)
	}
	s.productMap[in.Id] = in

	return &pb.ProductID{Value: in.Id}, status.New(codes.OK, "").Err()
}

// GetProduct implements ecommerce.GetProduct
func (s *server) GetProduct(ctx context.Context, in *pb.ProductID) (*pb.Product, error) {
	value, exists := s.productMap[in.Value]
	if exists {
		return value, status.New(codes.OK, "").Err()
	}

	return nil, status.Errorf(codes.NotFound, "Product does not exist.", in.Value)
}

// GetOrder service implementation.
func (s *server) GetOrder(ctx context.Context, orderId *wrappers.StringValue) (*pb.Order, error) {

	// Service Implementation
	ord, exist := orderMap[orderId.Value]
	if !exist {
		return nil, status.Errorf(codes.NotFound, "Order does not exist. Order ID: %s", orderId)
	}

	return &ord, status.New(codes.OK, "").Err()
}

// AddOrder service implementation.
func (s *server) AddOrder(ctx context.Context, orderReq *pb.Order) (*wrappers.StringValue, error) {
	log.Printf("Order Added. ID : %v", orderReq.Id)

	// Service Implementation
	orderMap[orderReq.Id] = *orderReq

	return &wrappers.StringValue{Value: "Order Added: " + orderReq.Id}, nil
}

// ProcessOrders is bi-directional Streaming RPC.
func (s *server) ProcessOrders(stream pb.OrderManagement_ProcessOrdersServer) error {

	batchMarker := 1
	var combinedShipmentMap = make(map[string]pb.CombinedShipment)
	for {
		orderId, err := stream.Recv()
		log.Printf("Reading Proc order : %s", orderId)
		if err == io.EOF {

			// Client has sent all the messages.
			// Send remaining shipments.
			log.Printf("EOF : %s", orderId)
			for _, shipment := range combinedShipmentMap {
				if err := stream.Send(&shipment); err != nil {
					return err
				}
			}

			return nil
		}
		if err != nil {
			log.Println(err)
			return err
		}

		// Logic to organize orders into shipments,
		// based on the destination.
		destination := orderMap[orderId.GetValue()].Destination
		shipment, found := combinedShipmentMap[destination]

		if found {
			ord := orderMap[orderId.GetValue()]
			shipment.OrdersList = append(shipment.OrdersList, &ord)
			combinedShipmentMap[destination] = shipment
		} else {
			comShip := pb.CombinedShipment{
				Id: "cmb - " + (orderMap[orderId.GetValue()].Destination),
				Status: "Processed!",
			}
			ord := orderMap[orderId.GetValue()]
			comShip.OrdersList = append(shipment.OrdersList, &ord)
			combinedShipmentMap[destination] = comShip
			log.Print(len(comShip.OrdersList), comShip.GetId())
		}

		if batchMarker == orderBatchSize {

			// Stream combined orders to the client in batches.
			for _, comb := range combinedShipmentMap {
				log.Printf("Shipping : %v -> %v" , comb.Id, len(comb.OrdersList))

				// Send combined shipment to the client.
				if err := stream.Send(&comb); err != nil {
					return err
				}
			}
			batchMarker = 0
			combinedShipmentMap = make(map[string]pb.CombinedShipment)
		} else {
			batchMarker++
		}
	}
}

func main() {
	initSampleData()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	//pb.RegisterProductInfoServer(s, &server{})
	pb.RegisterOrderManagementServer(s, &server{})

	log.Println("Starting gRPC listener on port " + port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func initSampleData() {
	orderMap["102"] = pb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Mac Book Pro"}, Destination: "Mountain View, CA", Price: 1800.00}
	orderMap["103"] = pb.Order{Id: "103", Items: []string{"Apple Watch S4"}, Destination: "San Jose, CA", Price: 400.00}
	orderMap["104"] = pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub"}, Destination: "Mountain View, CA", Price: 400.00}
	orderMap["105"] = pb.Order{Id: "105", Items: []string{"Amazon Echo"}, Destination: "San Jose, CA", Price: 30.00}
	orderMap["106"] = pb.Order{Id: "106", Items: []string{"Amazon Echo", "Apple iPhone XS"}, Destination: "Mountain View, CA", Price: 300.00}
}

