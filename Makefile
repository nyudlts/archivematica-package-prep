tidy:
	go mod tidy

build:
	go mod tidy
	go build -o build/ampp

install:
	sudo cp build/ampp /usr/local/bin/