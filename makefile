build_local:
	docker build -t moscow_events_backend:local .
build_hub:
	docker build -t udinsemen/moscow_events_backend:v1.0.0 .
deploy:
	scp -rv .env root@94.154.11.47:/etc/docker_compose
