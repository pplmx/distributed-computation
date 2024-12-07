package main

import (
	"flag"
	"log"

	"github.com/pplmx/distributed-computation/internal"
)

func main() {
	// Define command-line flags
	serviceType := flag.String("service", "", "Type of service to start (compute|registry|controller)")
	flag.Parse()

	// Check which service to start
	switch *serviceType {
	case "compute":
		log.Println("Starting Compute Service")
		internal.StartCompute()
	case "registry":
		log.Println("Starting Registry Service")
		internal.StartRegistry()
	case "controller":
		log.Println("Starting Controller Service")
		internal.StartController()
	default:
		log.Fatalf("Invalid service type. Use one of: compute, registry, controller")
	}
}
