FILE=docker-compose.yaml

build:
	docker-compose -f $(FILE) build

up:
	docker-compose -f $(FILE) up -d

down:
	docker-compose -f $(FILE) down

restart: down build up

clean:
	docker-compose -f $(FILE) down --volumes --remove-orphans
