all: build

build:
	go build -C cmd -o health-checks

run:
	./cmd/health-checks