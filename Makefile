generate_grpc_code:
	protoc \
	--go_out=invoicer \
	--go_opt=paths=source_relative \
	--go-grpc_out=invoicer \
	--go-grpc_opt=paths=source_relative \
	invoicer.proto

generate_grpc_docker:
	sudo docker run --rm -v "$$(pwd)":/app -w /app namely/protoc-all \
	--proto_path=. \
	--go_out=protoc-gen-go=. \
	--go-grpc_out=protoc-gen-go-grpc=. \
	--with-gogofaster=false \
	--go_opt=paths=source_relative \
	--go-grpc_opt=paths=source_relative \
	--go_source_relative \
	--out=./invoicer \
	--protofiles=invoicer.proto \
	--plugins=go,go-grpc
build:
	docker compose build

up:
	docker compose up

down:
	docker compose down

logs:
	docker compose logs -f go-server