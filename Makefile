.PHONY: all clean test build

clean:
	rm -f bin/

build:
	# mkdir -p bin/
	# dep ensure -v
	go build -o bin/petrel .
