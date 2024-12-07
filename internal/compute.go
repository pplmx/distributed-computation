package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	pb "github.com/pplmx/pb/dist/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// computeClient wraps gRPC client and connection details for distributed task management
type computeClient struct {
	client     pb.DistributedTaskServiceClient
	connection *grpc.ClientConn
	nodeID     string
}

// newComputeClient initializes a new gRPC client for distributed tasks
func newComputeClient(serverAddr, nodeID string) (*computeClient, error) {
	// Set up a connection to the server with error handling
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

// submitTask submits a compute task to the server and starts a watcher for updates
func (c *computeClient) submitTask(taskType pb.Task_TaskType, payload []byte) error {
	// Context with timeout to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	task := &pb.Task{
		Name:    "Compute Task",
		Type:    taskType,
		Payload: payload,
	}

	resp, err := c.client.SubmitTask(ctx, &pb.SubmitTaskRequest{
		Task:            task,
		RequesterNodeId: c.nodeID,
	})
	if err != nil {
		return err
	}

	log.Info().Msg(fmt.Sprintf("Task submitted successfully. Task ID: %s", resp.TaskId))

	// Start watching task status in a goroutine
	go c.watchTask(resp.TaskId)
	return nil
}

// watchTask listens for updates on a task's status
func (c *computeClient) watchTask(taskID string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := c.client.WatchTasks(ctx, &pb.WatchTasksRequest{
		NodeId:  c.nodeID,
		TaskIds: []string{taskID},
	})
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Failed to create watch stream: %v", err))
		return
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Error().Msg(fmt.Sprintf("Error receiving task update: %v", err))
			return
		}

		if statusChange, ok := resp.Event.(*pb.WatchTasksResponse_TaskStatusChange); ok {
			log.Info().Msg(fmt.Sprintf("Task %s status changed: %v -> %v",
				statusChange.TaskStatusChange.Task.Id,
				statusChange.TaskStatusChange.PreviousStatus,
				statusChange.TaskStatusChange.CurrentStatus,
			))
		}
	}
}

// cancelTask attempts to cancel a running task
func (c *computeClient) cancelTask(taskID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.CancelTask(ctx, &pb.CancelTaskRequest{
		TaskId:          taskID,
		RequesterNodeId: c.nodeID,
	})
	if err != nil {
		return err
	}

	if resp.Success {
		log.Info().Msg(fmt.Sprintf("Task %s canceled successfully", taskID))
	} else {
		log.Error().Msg(fmt.Sprintf("Failed to cancel task %s: %s", taskID, resp.ErrorMessage))
	}
	return nil
}

// close safely closes the gRPC connection
func (c *computeClient) close() {
	if c.connection != nil {
		c.connection.Close()
	}
}

// StartCompute demonstrates a complete lifecycle of task submission and observation
func StartCompute() {
	client, err := newComputeClient("localhost:50051", "compute-node-1")
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("Failed to create compute client: %v", err))
	}
	defer client.close()

	payload := []byte("sample compute payload")
	err = client.submitTask(pb.Task_TASK_TYPE_COMPUTE, payload)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("Failed to submit task: %v", err))
	}

	log.Info().Msg(fmt.Sprintf("Compute client is observing task updates"))
	time.Sleep(1 * time.Minute) // Simulate runtime to keep the application alive
}
