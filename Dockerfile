# Etapa de build
FROM golang:1.22.2-alpine as build
WORKDIR /app

# Argumento para especificar o diretório a ser buildado
ARG CMD_DIR=.

# Copia os arquivos para o diretório de trabalho
COPY . .

# Compila o binário
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o location ${CMD_DIR}

# Etapa final
FROM golang:1.22.2-alpine
WORKDIR /app

# Copia o binário compilado para a imagem final
COPY --from=build /app/location .

# Certifica-se de que o binário tem permissões de execução
RUN ["chmod", "+x", "/app/location"]

# Define a porta exposta
EXPOSE 8080

# Define o ponto de entrada
ENTRYPOINT ["/app/location"]
