.PHONY: all clean test build

DOCKER_REPOSITORY ?= monzo
DOCKER_TAG ?= latest
DOCKER_IMAGE ?= $(DOCKER_REPOSITORY)/kontrast:$(DOCKER_TAG)

clean:
	rm -rf bin/

build:
	mkdir -p bin/
	go build -o bin/kontrast ./cmd/kontrast

kontrastd:
	mkdir -p bin/
	go build -o bin/kontrastd ./cmd/kontrastd

build-in-docker:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/kontrastd ./cmd/kontrastd

build-linux:
	mkdir -p bin/
	GOOS=linux go build -o bin/kontrast ./cmd/kontrast

docker:
	docker build . -t $(DOCKER_IMAGE)

docker-push: docker
	docker push $(DOCKER_IMAGE)

