module github.com/go-enki/enki-example

go 1.13

require (
	github.com/go-chi/chi v4.0.2+incompatible // indirect
	github.com/go-chi/cors v1.0.0 // indirect
	github.com/go-chi/render v1.0.1 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/lukasjarosch/enki v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v0.9.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	go.uber.org/zap v1.10.0
	google.golang.org/grpc v1.24.0
)

replace github.com/lukasjarosch/enki => ../enki
