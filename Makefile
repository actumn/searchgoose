build:
	go build -o bin/searchgoose main.go

run:
	go run main.go


compile:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=arm go build -o bin/searchgoose-linux-arm main.go
	GOOS=linux GOARCH=arm64 go build -o bin/searchgoose-linux-arm64 main.go
	GOOS=linux GOARCH=amd64 go build -o bin/searchgoose-linux-amd64 main.go
	GOOS=linux GOARCH=386 go build -o bin/searchgoose-linux-386 main.go
	GOOS=freebsd GOARCH=386 go build -o bin/searchgoose-freebsd-386 main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/searchgoose-freebsd-amd64 main.go
	GOOS=freebsd GOARCH=amd64 go build -o bin/searchgoose-freebsd-amd main.go
	GOOS=windows GOARCH=386 go build -o bin/searchgoose-windows-386.exe main.go
	GOOS=windows GOARCH=amd64 go build -o bin/searchgoose-windows-amd64.exe main.go

docker:
	docker build -t actumn/searchgoose:latest .
