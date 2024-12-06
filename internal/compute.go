package internal

import (
	"context"
	pb "github.com/pplmx/pb/dist/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

type computeClient struct {
	client     pb.DistributedTaskServiceClient
	connection *grpc.ClientConn
	nodeID     string
}

func newComputeClient(serverAddr string, nodeID string) (*computeClient, error) {
	// Set up a connection to the server
	conn, err := grpc.Dial(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewDistributedTaskServiceClient(conn)
	return &computeClient{
		client:     client,
		connection: conn,
		nodeID:     nodeID,
	}, nil
}

func (c *computeClient) submitTask(taskType pb.Task_TaskType, payload []byte) error {
	// Create a new task
	task := &pb.Task{
		Name:    "Compute Task",
		Type:    taskType,
		Payload: payload,
	}

	// Submit the task
	resp, err := c.client.SubmitTask(context.Background(), &pb.SubmitTaskRequest{
		Task:            task,
		RequesterNodeId: c.nodeID,
	})
	if err != nil {
		return err
	}

	log.Printf("Task submitted successfully. Task ID: %s", resp.TaskId)

	// Watch the task
	go c.watchTask(resp.TaskId)

	return nil
}

func (c *computeClient) watchTask(taskID string) {
	// Create a stream to watch task status
	stream, err := c.client.WatchTasks(context.Background(), &pb.WatchTasksRequest{
		NodeId:  c.nodeID,
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
			log.Printf("Task %s status changed from %v to %v",
				event.TaskStatusChange.Task.Id,
				event.TaskStatusChange.PreviousStatus,
				event.TaskStatusChange.CurrentStatus,
			)
		}
	}
}

func (c *computeClient) cancelTask(taskID string) error {
	resp, err := c.client.CancelTask(context.Background(), &pb.CancelTaskRequest{
		TaskId:          taskID,
		RequesterNodeId: c.nodeID,
	})
	if err != nil {
		return err
	}

	if resp.Success {
		log.Printf("Task %s canceled successfully", taskID)
	} else {
		log.Printf("Failed to cancel task: %s", resp.ErrorMessage)
	}

	return nil
}

func (c *computeClient) close() {
	if c.connection != nil {
		c.connection.Close()
	}
}

func main() {
	// Create a compute client
	client, err := newComputeClient("localhost:50051", "compute-node-1")
	if err != nil {
		log.Fatalf("Failed to create compute client: %v", err)
	}
	defer client.close()

	// Submit a compute task
	err = client.submitTask(
		pb.Task_TASK_TYPE_COMPUTE,
		[]byte("sample compute payload"),
	)
	if err != nil {
		log.Fatalf("Failed to submit task: %v", err)
	}

	// Keep the client running to observe task updates
	time.Sleep(1 * time.Minute)
}
