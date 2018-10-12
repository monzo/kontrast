.PHONY: all clean test build

clean:
	rm -f bin/

build:
	mkdir -p bin/
	dep ensure -v
	go build -o bin/kryp .

build-linux:
	mkdir -p bin/
	dep ensure -v
	GOOS=linux go build -o bin/kryp .

docker: build-linux
	docker build . -t milesbxf/kryp
