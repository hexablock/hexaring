
deps:
	go get -d .

test:
	go test -v -cover .

protoc:
	protoc structs.proto -I ./ -I ../../../ --go_out=plugins=grpc:.
