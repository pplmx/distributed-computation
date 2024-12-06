package internal

import (
	"context"
	"fmt"
	pb "github.com/pplmx/pb/dist/v1"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

type distributedTaskServer struct {
	pb.UnimplementedDistributedTaskServiceServer

	// In-memory task storage
	tasks     map[string]*pb.Task
	taskMutex sync.RWMutex
}

func (s *distributedTaskServer) SubmitTask(ctx context.Context, req *pb.SubmitTaskRequest) (*pb.SubmitTaskResponse, error) {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	// Generate a unique task ID if not provided
	if req.Task.Id == "" {
		req.Task.Id = fmt.Sprintf("task-%d", len(s.tasks)+1)
	}

	// Set initial task status
	req.Task.Status = pb.Task_TASK_STATUS_PENDING

	// Store the task
	s.tasks[req.Task.Id] = req.Task

	log.Printf("Task submitted: %s (Type: %v)", req.Task.Id, req.Task.Type)

	return &pb.SubmitTaskResponse{
		Accepted: true,
		TaskId:   req.Task.Id,
	}, nil
}

func (s *distributedTaskServer) WatchTasks(req *pb.WatchTasksRequest, stream pb.DistributedTaskService_WatchTasksServer) error {
	// Simulate task status changes and assignments
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			// In a real implementation, this would be event-driven
			for _, taskID := range req.TaskIds {
				s.taskMutex.RLock()
				task, exists := s.tasks[taskID]
				s.taskMutex.RUnlock()

				if exists {
					// Simulate a task status change
					if task.Status == pb.Task_TASK_STATUS_PENDING {
						oldStatus := task.Status
						task.Status = pb.Task_TASK_STATUS_RUNNING

						// Send task status change event
						statusChangeEvent := &pb.WatchTasksResponse{
							Event: &pb.WatchTasksResponse_TaskStatusChange{
								TaskStatusChange: &pb.TaskStatusChangeEvent{
									Task:           task,
									PreviousStatus: oldStatus,
									CurrentStatus:  task.Status,
								},
							},
						}
						if err := stream.Send(statusChangeEvent); err != nil {
							return err
						}
					}
				}
			}
		}
	}
}

func (s *distributedTaskServer) CancelTask(ctx context.Context, req *pb.CancelTaskRequest) (*pb.CancelTaskResponse, error) {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	task, exists := s.tasks[req.TaskId]
	if !exists {
		return &pb.CancelTaskResponse{
			Success:      false,
			ErrorMessage: "Task not found",
		}, nil
	}

	// Only cancel tasks that are pending or running
	if task.Status == pb.Task_TASK_STATUS_PENDING || task.Status == pb.Task_TASK_STATUS_RUNNING {
		task.Status = pb.Task_TASK_STATUS_CANCELED
		log.Printf("Task canceled: %s", req.TaskId)
		return &pb.CancelTaskResponse{Success: true}, nil
	}

	return &pb.CancelTaskResponse{
		Success:      false,
		ErrorMessage: "Task cannot be canceled in current state",
	}, nil
}

func newDistributedTaskServer() *distributedTaskServer {
	return &distributedTaskServer{
		tasks: make(map[string]*pb.Task),
	}
}

func main() {
	// Create a gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterDistributedTaskServiceServer(grpcServer, newDistributedTaskServer())

	log.Println("Distributed Task Service started on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
