build:
	docker-compose build
up:
	docker-compose up -d --build
down:
	docker-compose down -v --remove-orphans
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
fmt:
	go fmt
