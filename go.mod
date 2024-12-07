module github.com/pplmx/distributed-computation

go 1.23.3

require (
	github.com/pplmx/pb/dist v0.0.0-20241206125115-b0306a987a92
	github.com/rs/zerolog v1.33.0
	google.golang.org/grpc v1.68.1
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/net v0.32.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241206012308-a4fef0638583 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)

// replace github.com/pplmx/pb/dist => ../pb/dist
