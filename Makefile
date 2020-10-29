default:
	$(MAKE) all

windows:
	mkdir -p dist/windows
	env GOOS=windows GOARCH=amd64 go build -o dist/windows/tokenmachine.exe main.go

linux:
	mkdir -p dist/linux
	env GOOS=linux GOARCH=amd64 go build -o dist/linux/tokenmachine main.go

darwin:
	mkdir -p dist/darwin
	env GOOS=darwin GOARCH=amd64 go build -o dist/darwin/tokenmachine main.go

all:
	$(MAKE) windows
	$(MAKE) darwin
	$(MAKE) linux