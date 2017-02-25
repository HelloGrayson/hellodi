
.PHONY: install
install:
	glide install
	go install ./vendor/go.uber.org/thriftrw
	go install ./vendor/go.uber.org/yarpc/encoding/thrift/thriftrw-plugin-yarpc

.PHONY: generate
generate:
	thriftrw --plugin=yarpc hello.thrift

.PHONY: run
run:
	go build .
	./hellodi
