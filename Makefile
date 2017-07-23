
protoc:
	protoc structs.proto -I ./ -I ../../../ --go_out=plugins=grpc:.
