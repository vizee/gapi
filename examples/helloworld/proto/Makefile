GAPI_PROTO ?= .

proto:
	@protoc -I . -I $(GAPI_PROTO) --descriptor_set_out=helloworld.pd \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative helloworld.proto
