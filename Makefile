clean:
	rm -rf ./**/*.exe

test:
	go test $(go list ./... | grep -v /utils)
