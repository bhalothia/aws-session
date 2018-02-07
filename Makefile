.PHONY: clean bin test

bin/aws-assume: *.go
	go build -o $@ .

bin: bin/aws-assume-Linux bin/aws-assume-Darwin bin/aws-assume-Windows.exe 

bin/aws-assume-Linux: *.go
	env GOOS=linux go build -o $@ .
bin/aws-assume-Darwin: *.go
	env GOOS=darwin go build -o $@ .
bin/aws-assume-Windows.exe: *.go
	env GOOS=windows go build -o $@ .

clean:
	rm -rf bin/*

test:
	go test -race $(shell go list ./... | grep -v /vendor/)
