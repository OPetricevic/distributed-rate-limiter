build:
	go build -o bin/node.exe ./cmd/node/
run:
	go run ./cmd/node/
test: 
	go test -v ./...
clean:
	rmdir /s /q bin 2>nul || true
	del test.exe 2>nul || true
: