## Instruções

#### 1 - Copie o arquivo `.env.example` para `.env`
Ajuste as variáveis que julgar necessário:

#### Variáveis de ambiente default:

- REDIS_HOST=redis # Host do Redis
- REDIS_PORT=6379 # Porta do Redis
- REDIS_DB=0 # Banco do Redis
- APP_PORT=8080 # Porta do servidor
- DEFAULT_RATE_LIMIT=5 # Rate limit padrão por segundo
- JWT_SECRET="dEz9XMDH3lK1b8vcHaQ4b8D7I53Et6pm" # Chave secreta do JWT
- JWT_EXPIRES_IN=18000 # Tempo de expiração do token em segundos
- BAN_TIME=3600 # Tempo de banimento em segundos

No arquivo `cmd/server/main.go` linha 35, é possível alterar o rate limit do superUser. Valor padrão 100.

#### 2 - Inicie o servidor com o comando abaixo:
`docker compose up -d`

Aguarde o servidor iniciar. Pode demorar alguns segundos por causa do `go mod tidy`.

#### 3 - Utilize o seguinte comando para obter um token de acesso com limite de 100 request por segundo:
```bash
curl --location 'http://127.0.0.1:8080/auth/login' \
    --header 'Content-Type: application/json' \
    --header 'Accept: application/json' \
    --data '{"user": "superUser"}'
```

Obs.: Qualquer outro nome de usuário diferente de "superUser" terá o rate limit padrão.

#### 4 - Utilize o endereço `http://127.0.0.1:8080/` com método GET para realizar requisições.
O token deve ser passado no header API_KEY (sem o prefixo Bearer). Exemplo:

```bash
curl --location 'http://127.0.0.1:8080/' \
    --header 'API_KEY: TOKEN_OBTIDO_NO_PASSO_3' \
    --header 'Content-Type: application/json' \
    --header 'Accept: application/json'
```
Obs.: O header API_KEY é opcional.

#### 5 - Executar testes
`docker compose run --rm api go test . -v`
