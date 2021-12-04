TOOLS=\
	google.golang.org/protobuf/cmd/protoc-gen-go@v1.26 \
	google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1

install-tools:
	@for tool in $(TOOLS) ; do \
		go install $$tool; \
	done

.PHONY: gen
gen:
	@protoc -I proto \
		--go_out proto/gen \
		--go_opt paths=source_relative \
		--go-grpc_out proto/gen \
		--go-grpc_opt paths=source_relative \
		proto/ojichat.proto

.PHONY: run-server
run-server:
	@go run ./cmd/server/main.go

.PHONY: run-client
run-client:
	@go run ./cmd/client/main.go --name seiichi

.PHONY: deploy
deploy: docker-build docker-push docker-deploy

.PHONY: docker-build
docker-build:
	@docker build -t gcr.io/<your-account-id>/ojichat-stream:latest .

.PHONY: docker-push
docker-push:
	@docker push gcr.io/<your-account-id>/ojichat-stream:latest

.PHONY: docker-deploy
docker-deploy:
	@gcloud run deploy --image gcr.io/<your-account-id>/ojichat-stream:latest --platform managed --region asia-northeast1 --use-http2 --allow-unauthenticated --port 50051

.PHONY: delete
delete:
	@gcloud run services delete ojichat-stream --region asia-northeast1

