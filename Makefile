proto:
	protoc user.proto --go_out=./proto --go-grpc_out=./proto

db:
	sqlite3 data.sqlite < user.sql
