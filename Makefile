build:
	docker-compose build
up:
	docker-compose up -d
down:
	docker-compose down -v
logs:
	docker-compose logs -f
reset:
	docker-compose down -v
	docker-compose up -d
run:
	go run .
test:
	go test -v
lint:
	golangci-lint run
format:
	go fmt
