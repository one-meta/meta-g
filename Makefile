build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o meta-g main.go

run:
	go run main.go -merge=true