build:
	go build -o marlinraker -ldflags="-s -w" -trimpath src/main.go