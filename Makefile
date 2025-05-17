start_lb:
	go run cmd/balancer/main.go cmd/balancer/config.json

deploy_rate:
	docker compose down -v
	docker compose up --build