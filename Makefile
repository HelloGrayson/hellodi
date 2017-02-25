
.PHONY: install
install:
	glide install
	go install ./vendor/go.uber.org/thriftrw
	go install ./vendor/go.uber.org/yarpc/encoding/thrift/thriftrw-plugin-yarpc
	go get -u -f github.com/yarpc/yab

.PHONY: generate
generate:
	thriftrw --plugin=yarpc hello.thrift

.PHONY: run
run:
	go build .
	./hellodi

.PHONY: yab
yab:
	yab -p http://localhost:8086 --disable-thrift-envelope \
		-t hello.thrift hello Hello::echo -r '{"echo": {"message": "yo", "count": 5}}'
