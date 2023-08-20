build:
	go build -o bin/ping-monitor main.go

run:
	go run main.go


compile:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=arm go build -o bin/main-linux-arm ping-monitor.go
	GOOS=linux GOARCH=arm64 go build -o bin/main-linux-arm64 ping-monitor.go
	GOOS=freebsd GOARCH=386 go build -o bin/main-freebsd-386 ping-monitor.go
