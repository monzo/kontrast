DOCKER_REPOSITORY ?= monzo
DOCKER_TAG ?= latest
DOCKER_IMAGE ?= $(DOCKER_REPOSITORY)/kontrast:$(DOCKER_TAG)

ALL_GOARCH = amd64
ALL_GOOS = linux darwin

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: build
build:
	mkdir -p bin/
	dep ensure -v
	go build -o bin/kontrast ./cmd/kontrast

.PHONY: kontrastd
kontrastd:
	mkdir -p bin/
	dep ensure -v
	go build -o bin/kontrastd ./cmd/kontrastd

.PHONY: build-in-docker
build-in-docker:
	dep ensure -vendor-only -v
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/kontrastd ./cmd/kontrastd

.PHONY: build-linux
build-linux:
	mkdir -p bin/
	dep ensure -v
	GOOS=linux go build -o bin/kontrast ./cmd/kontrast

.PHONY: dist
dist:
	$(eval export ALL_GOARCH)
	$(eval export ALL_GOOS)
	mkdir -p bin/
	NAME=kontrast ./dist.sh

.PHONY: docker
docker:
	docker build . -t $(DOCKER_IMAGE)

.PHONY: docker-push
docker-push: docker
	docker push $(DOCKER_IMAGE)

