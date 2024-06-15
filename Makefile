# Verifica a existência do arquivo .env antes de executar qualquer target
.env:
	@if [ ! -f .env ]; then \
		echo "Erro: arquivo .env não encontrado!"; \
		exit 1; \
	fi

# Os outros targets dependem da regra .env
all: .env test
	docker-compose up --build -d

down: .env
	docker-compose down

test: .env
	go test ./...