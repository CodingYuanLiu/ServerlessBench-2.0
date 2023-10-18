.PHONY: all build clean
	BIN_FILE=hello
BIN_FILE=cli
goBindata=$(shell which go-bindata)

all: build
build:
    ifeq ($(goBindata),)
			go get -u github.com/jteeuwen/go-bindata/...
			sudo apt-get install go-bindata
    endif
		mkdir resources
		cp templates/testcase.yaml resources
		cp templates/unitTestTemplate.yaml resources
		cp -r templates/kn/* resources
		cp -r templates/fission/* resources
		cp -r templates/openfaas/* resources
		cp -r templates/openwhisk/* resources
		zip -r resources/faastemplate.zip resources/faastemplate
		zip -r resources/python.zip resources/python
		go-bindata -o=myres.go  resources
		rm -rf internal/util/myres.go
		touch internal/util/myres.go
		sed '26c package util' myres.go > internal/util/myres.go
		rm -rf myres.go resources
	   @go build -o "${BIN_FILE}" ./cmd/cli
clean:
	   @go clean
	        rm --force ${BIN_FILE}
