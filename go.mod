module github.com/pplmx/distributed-computation

go 1.24.0

require (
	github.com/pplmx/pb/dist v0.0.0-20241206125115-b0306a987a92
	github.com/rs/zerolog v1.33.0
	google.golang.org/grpc v1.79.3
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

// replace github.com/pplmx/pb/dist => ../pb/dist
