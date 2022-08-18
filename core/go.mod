module core

go 1.15

require (
	dynatrace.com v1.0.0
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
	google.golang.org/protobuf v1.28.0
)

replace dynatrace.com => ./internal/proto/dynatrace.com
