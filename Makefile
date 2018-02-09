.PHONY: clean bin test

bin/aws-assume: *.go
	go build -o $@ .


bin: bin/linux/aws-assume bin/darwin/aws-assume bin/windows/aws-assume.exe

bin/linux/aws-assume: *.go
	env GOOS=$(shell basename `dirname $@`) go build -o $@ .
bin/darwin/aws-assume: *.go
	env GOOS=$(shell basename `dirname $@`) go build -o $@ .
bin/windows/aws-assume.exe: *.go
	env GOOS=$(shell basename `dirname $@`) go build -o $@ .


dist: dist/aws-assume-linux.zip dist/aws-assume-darwin.zip dist/aws-assume-windows.zip
	
dist/aws-assume-linux.zip: bin/linux/aws-assume
	mkdir -p $(dir $@) && cd bin/$(shell echo $@ | sed 's|.*-\(.*\).zip|\1|') && pwd  && zip -r $(abspath $@) .
dist/aws-assume-darwin.zip: bin/darwin/aws-assume
	mkdir -p $(dir $@) && cd bin/$(shell echo $@ | sed 's|.*-\(.*\).zip|\1|') && pwd  && zip -r $(abspath $@) .
dist/aws-assume-windows.zip: bin/windows/aws-assume.exe 
	mkdir -p $(dir $@) && cd bin/$(shell echo $@ | sed 's|.*-\(.*\).zip|\1|') && pwd  && zip -r $(abspath $@) .


clean:
	rm -rf bin/*


test:
	go test -race $(shell go list ./... | grep -v /vendor/)
