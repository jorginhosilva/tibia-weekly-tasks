# Build em 3 etapas: compila o front-end (Node), compila o servidor (Go)
# e junta tudo numa imagem final pequena.

# 1) Front-end: gera frontend/dist (HTML pré-renderizado + assets)
FROM node:22-alpine AS front
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# 2) Servidor Go
FROM golang:1.26-alpine AS back
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /tibia-tasks .

# 3) Imagem final
FROM alpine:3.20
RUN adduser -D app
WORKDIR /app
COPY --from=back /tibia-tasks ./tibia-tasks
COPY --from=front /app/frontend/dist ./frontend/dist
USER app
EXPOSE 8080
# As credenciais do banco vêm de variáveis de ambiente (DB_USER, DB_PASSWORD,
# DB_HOST, DB_PORT, DB_NAME), configuradas no painel da hospedagem.
CMD ["./tibia-tasks", "web"]
