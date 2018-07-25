.PHONY: clean bin test

bin/aws-assume:
	go build -o $@ .


bin: bin/linux/aws-assume bin/darwin/aws-assume bin/windows/aws-assume.exe

bin/linux/aws-assume: 
	env GOOS=linux go build -o $@ .
bin/darwin/aws-assume: 
	env GOOS=darwin go build -o $@ .
bin/windows/aws-assume.exe:
	env GOOS=windows go build -o $@ .


dist: dist/aws-assume-linux.zip dist/aws-assume-darwin.zip dist/aws-assume-windows.zip
	
dist/aws-assume-linux.zip: bin/linux/aws-assume
	mkdir -p $(dir $@) && cd bin/linux && pwd  && zip -r $(abspath $@) .
dist/aws-assume-darwin.zip: bin/darwin/aws-assume
	mkdir -p $(dir $@) && cd bin/darwin && pwd  && zip -r $(abspath $@) .
dist/aws-assume-windows.zip: bin/windows/aws-assume.exe 
	mkdir -p $(dir $@) && cd bin/windows && pwd  && zip -r $(abspath $@) .


clean:
	rm -rf bin/*
