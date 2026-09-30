# RO Lobby

Lobby para jogadores de Ragnarok Online montarem grupos para instâncias difíceis.
Projeto de fã, sem vínculo com a Gravity.

- `backend/`: API em Go (`net/http`, `pgx`, `sqlc`, `goose`)
- `web/`: SvelteKit com `adapter-node`
- `openapi.yaml`: contrato da API, fonte dos tipos do Go e do TypeScript
- `docs/adr/`: decisões de arquitetura
- `.specs/`: specs de cada feature

## Setup local no Windows

### 1. Pré-requisitos

| Ferramenta | Versão | Como instalar |
|---|---|---|
| Git | qualquer recente | `winget install Git.Git` |
| Go | 1.27 (a do `backend/go.mod`) | `winget install GoLang.Go` |
| Node.js | 24.14.0 (a do `.nvmrc`) | `winget install OpenJS.NodeJS --version 24.14.0` ou o `nvm-windows` |
| Docker Desktop | com WSL 2 | `winget install Docker.DockerDesktop` |

Antes do Docker Desktop:

1. Ative a virtualização na BIOS/UEFI, se ainda não estiver ativa.
2. Rode `wsl --install` num PowerShell como administrador e reinicie.

Depois de instalar, abra o Docker Desktop uma vez e espere o engine subir.
Abra um terminal novo para o `docker` entrar no PATH.

> **Smart App Control:** no Windows 11, o Smart App Control bloqueia os executáveis
> sem assinatura que o `go test` e o `go build` geram. O erro é "Uma política de
> Controle de Aplicativo bloqueou este arquivo". Desative em **Segurança do Windows →
> Controle de aplicativos e navegador → Smart App Control → Desativado**. O Windows só
> deixa religar reinstalando o sistema. O Defender continua ativo.

Confira as versões:

```powershell
go version       # go1.27.x
node --version   # v24.14.0
docker version   # Client e Server respondendo
```

### 2. Variáveis de ambiente

Todo o projeto lê um único `.env` na raiz: o Compose, o backend e o web.

```powershell
Copy-Item .env.example .env
```

Os valores do `.env.example` são fictícios e servem para o ambiente local.

> **`.env` antigo?** Se o seu `.env` é de antes da Home, acrescente
> `COMPOSE_PROFILES=dev` e `APP_WEB_PORT=3000`, ou copie o `.env.example` de novo. Sem o
> `COMPOSE_PROFILES`, o `docker compose up` não sobe nenhum serviço.

**Porta ocupada?** Troque `POSTGRES_PORT`, `API_PORT`, `WEB_PORT` ou `APP_WEB_PORT` no `.env`. Se
mudar `POSTGRES_PORT` ou `API_PORT`, ajuste também a `DATABASE_URL` ou a
`API_BASE_URL`.

### 3. Banco de dados

```powershell
docker compose up -d --wait        # sobe o PostgreSQL 18 com volume nomeado
cd backend
go run ./cmd/migrate up            # aplica as migrações
go run ./cmd/migrate status        # confere
```

Os dados ficam no volume `ro-lobby_postgres-data` e sobrevivem a
`docker compose down`. Para apagar tudo: `docker compose down -v`.

Outros comandos de migração: `go run ./cmd/migrate down` reverte a última e
`go run ./cmd/migrate reset` reverte todas.

### 4. Backend

```powershell
cd backend
go run ./cmd/api
```

Em outro terminal:

```powershell
curl.exe http://localhost:8080/healthz
# {"database":"ok","status":"ok"}
```

### 5. Web

```powershell
cd web
npm ci
npm run dev
```

Abra http://localhost:5173. A Home mostra os grupos do dia, com dados fictícios, e
http://localhost:5173/status mostra **API online**.

## Pilha completa em containers (produção local)

A hospedagem, por enquanto, é local ([ADR-06](docs/adr/0006-hospedagem-local.md)). Um
único comando sobe um PostgreSQL próprio, as migrações, a API e o web com o build de
produção, como num servidor:

```powershell
docker compose --profile app up -d --wait --build   # sobe tudo
```

Abra http://localhost:3000. A porta vem de `APP_WEB_PORT`. Só o web fica exposto: a API
e o banco da pilha ficam na rede interna do Compose. As migrações rodam antes da API, e
se uma delas falhar a API não sobe.

```powershell
docker compose --profile app logs -f web api   # acompanha os logs
docker compose --profile app down              # para, mantendo os dados
docker compose --profile app down -v           # para e apaga o banco da pilha
```

O banco da pilha usa o volume `ro-lobby_app-postgres-data`, separado do banco de
desenvolvimento. Para conferir a pilha inteira de uma vez, rode o teste de fumaça, o
mesmo da CI:

```powershell
bash scripts/smoke-app.sh
```

## Testes e verificações

### Backend (em `backend/`)

```powershell
go test ./...                          # unitários, não precisam de banco
$env:DATABASE_URL = "postgres://ro_lobby:ro_lobby_dev@localhost:5432/ro_lobby?sslmode=disable"
go test -tags=integration ./...        # integração, com o PostgreSQL do Compose no ar
gofmt -l .                             # não deve listar nada
golangci-lint run ./...
```

Os testes de integração criam um banco descartável para cada teste e apagam no fim.
O banco de desenvolvimento não é tocado.

Para instalar o `golangci-lint`, na mesma versão da CI:
`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0`.

### Web (em `web/`)

```powershell
npm run lint             # Prettier e ESLint
npm run check            # svelte-check, TypeScript estrito
npm test                 # Vitest
npm run test:coverage    # Vitest com cobertura em web/coverage
npm run build
npx playwright install chromium   # uma vez
npm run test:e2e                  # ponta a ponta: sobe o backend e o web sozinho
```

O ponta a ponta precisa do banco no ar e migrado (passo 3).

## Contrato e código gerado

Toda rota nova começa no `openapi.yaml`. Depois de mudar o contrato, uma migração ou
uma query do `sqlc`, gere o código de novo e faça commit junto:

```powershell
# a partir da raiz do repositório
cd backend; go generate ./...; cd ..   # oapi-codegen e sqlc
cd web; npm run generate; cd ..        # openapi-typescript
```

A CI gera de novo e falha se o resultado for diferente do que foi commitado.

## CI

O workflow `.github/workflows/ci.yml` roda em todo PR para a `main`:

- **Backend**, quando `backend/` muda: gofmt, golangci-lint, testes unitários, migrações,
  testes de integração, build e código gerado.
- **Web**, quando `web/` muda: ESLint, Prettier, svelte-check, Vitest com cobertura,
  build e tipos gerados.
- **Ponta a ponta**, quando o backend ou o web mudam.
- **Pilha app**, quando o backend, o web ou o Compose mudam: constrói as imagens e roda
  o `scripts/smoke-app.sh`. Não publica as imagens.
- O `openapi.yaml` e a própria CI disparam todos.
- **CI ok** junta os resultados. PR só de documentação passa direto.

A cobertura aparece no resumo de cada job e como artefato, sem percentual mínimo.

### Proteção da `main`

No GitHub: **Settings → Branches → Add branch ruleset** (ou *Add rule*) para a `main`:

1. Exigir pull request antes do merge.
2. Exigir status checks e marcar **CI ok**. Não marque os jobs Backend, Web e Ponta a
   ponta: eles ficam pulados em PRs que não mexem na pasta deles.
3. Bloquear force push.
