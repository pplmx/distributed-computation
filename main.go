package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/pplmx/pb/dist/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DistributedNode struct {
	ID          string
	Name        string
	Address     string
	Port        uint32
	Registry    pb.RegistryServiceClient
	TaskService pb.DistributedTaskServiceClient
}

func newDistributedNode(id, name, address string, port uint32) *DistributedNode {
	// Connect to Registry Service
	registryConn, err := grpc.Dial(
		"localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to registry service: %v", err)
	}

	// Connect to Distributed Task Service
	taskConn, err := grpc.Dial(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to task service: %v", err)
	}

	// Create node
	node := &DistributedNode{
		ID:          id,
		Name:        name,
		Address:     address,
		Port:        port,
		Registry:    pb.NewRegistryServiceClient(registryConn),
		TaskService: pb.NewDistributedTaskServiceClient(taskConn),
	}

	return node
}

func (n *DistributedNode) register() error {
	// Prepare node registration
	nodeProto := &pb.Node{
		Id:      n.ID,
		Name:    n.Name,
		Address: n.Address,
		Port:    n.Port,
		Status:  pb.Node_NODE_STATUS_ACTIVE,
	}

	// Attempt to register the node
	resp, err := n.Registry.Register(context.Background(), &pb.RegisterRequest{
		Node: nodeProto,
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		log.Printf("Node registration failed: %s", resp.ErrorMessage)
		return fmt.Errorf("registration failed: %s", resp.ErrorMessage)
	}

	log.Printf("Node %s registered successfully", n.ID)
	return nil
}

func (n *DistributedNode) startHeartbeat(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			_, err := n.Registry.Heartbeat(context.Background(), &pb.HeartbeatRequest{
				NodeId: n.ID,
			})
			if err != nil {
				log.Printf("Heartbeat failed for node %s: %v", n.ID, err)
			}
		}
	}()
}

func (n *DistributedNode) submitComputeTask() error {
	// Create a compute task
	task := &pb.Task{
		Name:    "Sample Compute Task",
		Type:    pb.Task_TASK_TYPE_COMPUTE,
		Payload: []byte("Distributed computing payload"),
	}

	// Submit the task
	resp, err := n.TaskService.SubmitTask(context.Background(), &pb.SubmitTaskRequest{
		Task:            task,
		RequesterNodeId: n.ID,
	})
	if err != nil {
		return err
	}

	log.Printf("Task submitted successfully. Task ID: %s", resp.TaskId)

	// Watch the task (in a real scenario, this would be more robust)
	go n.watchTask(resp.TaskId)

	return nil
}

func (n *DistributedNode) watchTask(taskID string) {
	stream, err := n.TaskService.WatchTasks(context.Background(), &pb.WatchTasksRequest{
		NodeId:  n.ID,
		TaskIds: []string{taskID},
	})
	if err != nil {
		log.Printf("Error creating watch stream: %v", err)
		return
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving task update: %v", err)
			return
		}

		switch event := resp.Event.(type) {
		case *pb.WatchTasksResponse_TaskStatusChange:
			log.Printf("Task %s status changed: %v -> %v",
				event.TaskStatusChange.Task.Id,
				event.TaskStatusChange.PreviousStatus,
				event.TaskStatusChange.CurrentStatus,
			)
		}
	}
}

func main() {
	// Create a distributed node
	node := newDistributedNode(
		"node-001",         // Node ID
		"Compute Node 001", // Node Name
		"localhost",        // Address
		50053,              // Port
	)

	// Register the node
	if err := node.register(); err != nil {
		log.Fatalf("Failed to register node: %v", err)
	}

	// Start periodic heartbeats
	node.startHeartbeat(15 * time.Second)

	// Submit a compute task
	if err := node.submitComputeTask(); err != nil {
		log.Fatalf("Failed to submit compute task: %v", err)
	}

	// Keep the application running
	select {}
}
