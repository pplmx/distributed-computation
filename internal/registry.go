package internal

import (
	"context"
	pb "github.com/pplmx/pb/dist/v1"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
	"time"
)

type registryServer struct {
	pb.UnimplementedRegistryServiceServer

	// In-memory node registry
	nodes     map[string]*pb.Node
	nodeMutex sync.RWMutex

	// Heartbeat tracking
	nodeLastHeartbeat map[string]time.Time
}

func (s *registryServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	s.nodeMutex.Lock()
	defer s.nodeMutex.Unlock()

	node := req.Node

	// Check if node already exists
	if _, exists := s.nodes[node.Id]; exists {
		return &pb.RegisterResponse{
			Success:      false,
			ErrorMessage: "Node already registered",
		}, nil
	}

	// Set initial node status
	node.Status = pb.Node_NODE_STATUS_ACTIVE

	// Store the node
	s.nodes[node.Id] = node
	s.nodeLastHeartbeat[node.Id] = time.Now()

	log.Printf("Node registered: %s (Address: %s:%d)", node.Id, node.Address, node.Port)

	return &pb.RegisterResponse{Success: true}, nil
}

func (s *registryServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	s.nodeMutex.Lock()
	defer s.nodeMutex.Unlock()

	node, exists := s.nodes[req.NodeId]
	if !exists {
		return &pb.HeartbeatResponse{
			Success:      false,
			ErrorMessage: "Node not found in registry",
		}, nil
	}

	// Update last heartbeat time
	s.nodeLastHeartbeat[req.NodeId] = time.Now()

	// Ensure node status is active
	if node.Status != pb.Node_NODE_STATUS_ACTIVE {
		oldStatus := node.Status
		node.Status = pb.Node_NODE_STATUS_ACTIVE

		// Notify about status change
		go s.notifyNodeStatusChange(node, oldStatus)
	}

	return &pb.HeartbeatResponse{Success: true}, nil
}

func (s *registryServer) WatchRegistry(req *pb.WatchRegistryRequest, stream pb.RegistryService_WatchRegistryServer) error {
	// Simulate registry events
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			s.checkNodeHeartbeats(stream)
			time.Sleep(10 * time.Second)
		}
	}
}

func (s *registryServer) checkNodeHeartbeats(stream pb.RegistryService_WatchRegistryServer) {
	s.nodeMutex.Lock()
	defer s.nodeMutex.Unlock()

	now := time.Now()
	for nodeID, lastHeartbeat := range s.nodeLastHeartbeat {
		// Consider a node inactive if no heartbeat for 30 seconds
		if now.Sub(lastHeartbeat) > 30*time.Second {
			node := s.nodes[nodeID]
			if node.Status == pb.Node_NODE_STATUS_ACTIVE {
				// oldStatus := node.Status // keep the old
				node.Status = pb.Node_NODE_STATUS_INACTIVE

				// Send shutdown event
				shutdownEvent := &pb.WatchRegistryResponse{
					Event: &pb.WatchRegistryResponse_Shutdown{
						Shutdown: &pb.ShutdownEvent{
							NodeId: nodeID,
							Reason: "Heartbeat timeout",
						},
					},
				}

				if err := stream.Send(shutdownEvent); err != nil {
					log.Printf("Error sending shutdown event: %v", err)
				}
			}
		}
	}
}

func (s *registryServer) notifyNodeStatusChange(node *pb.Node, oldStatus pb.Node_NodeStatus) {
	// In a real implementation, this would broadcast to interested parties
	log.Printf("Node %s status changed from %v to %v",
		node.Id,
		oldStatus,
		node.Status,
	)
}

func newRegistryServer() *registryServer {
	return &registryServer{
		nodes:             make(map[string]*pb.Node),
		nodeLastHeartbeat: make(map[string]time.Time),
	}
}

func main() {
	// Create a gRPC server
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRegistryServiceServer(grpcServer, newRegistryServer())

	log.Println("Registry Service started on :50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
