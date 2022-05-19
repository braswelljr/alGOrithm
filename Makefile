clean:
	rm -rf ./**/*.exe ./**/*.zip ./**/*.rar

test:
	go test $(go list ./... | grep -v /utils)
