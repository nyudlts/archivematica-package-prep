build:
	go mod tidy
	go build -o ampp

install:
	cp ampp /usr/local/bin