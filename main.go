package main

import (
	"flag"

	"github.com/pplmx/distributed-computation/internal"
	"github.com/rs/zerolog/log"
)

func main() {
	// Define command-line flags
	serviceType := flag.String("service", "", "Type of service to start (compute|registry|controller)")
	flag.Parse()

	// Check which service to start
	switch *serviceType {
	case "compute":
		log.Info().Msg("Starting Compute Service")
		internal.StartCompute()
	case "registry":
		log.Info().Msg("Starting Registry Service")
		internal.StartRegistry()
	case "controller":
		log.Info().Msg("Starting Controller Service")
		internal.StartController()
	default:
		log.Fatal().Msg("Invalid service type. Use one of: compute, registry, controller")
	}
}
