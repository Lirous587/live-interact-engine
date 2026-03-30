PROTO_DIR := proto
PROTO_SRC := \
    $(wildcard proto/danmaku/*.proto) \
    $(wildcard proto/payment/*.proto) \
    $(wildcard proto/room/*.proto)
GO_OUT := .

.PHONY: generate-proto
generate-proto:
	protoc --proto_path=$(PROTO_DIR) --go_out=$(GO_OUT) --go-grpc_out=$(GO_OUT) $(PROTO_SRC)