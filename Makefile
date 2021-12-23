run: main.go
	go run main.go

build: main.go
	go build main.go

rm_exe:
	rm -rf *.exe

install:
	go install main.go