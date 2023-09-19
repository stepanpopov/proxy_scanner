# proxy_scanner

Запуск прокси:
- go run ./cmd/proxy/main.go -ca_cert_file ./certs/ca.crt -ca_key_file ./certs/ca.key -listen_addr 8000

Запуск db:
- docker-compose up

Запуск api:
- go run ./cmd/api/main.go
Запуститься на 8080 порту