build:
	docker-compose build
up:
	docker-compose up -d --build
down:
	docker-compose down -v --remove-orphans
logs:
	docker-compose --ansi always logs -f --tail 1000
reset:
	docker-compose down -v --remove-orphans
	docker-compose up -d --build
run:
	go run .
test:
	go test './...'
lint:
	golangci-lint run
fmt:
	go fmt 'github.com/VictorAnnell/...'
	go mod tidy
	goimports -w .
