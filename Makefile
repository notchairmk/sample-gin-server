.PHONY: build
build:
	go build -o ./bin/sample-server

.PHONY: build-client
build-client:
	cd client; \
		npm install; \
		npm run build


.PHONY: run
run:
	go run ./...
