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
> `COMPOSE_PROFILES`, o `docker compose up` não sobe nenhum serviço. Se é de antes do
> endurecimento, acrescente `APP_POSTGRES_PASSWORD` (veja [Segurança](#segurança)).

Para conferir o `.env` sem mostrar nenhum valor:

```powershell
sh scripts/check-env.sh
```

**Porta ocupada?** Troque `POSTGRES_PORT`, `API_PORT`, `WEB_PORT` ou `APP_WEB_PORT` no `.env`. Se
mudar `POSTGRES_PORT` ou `API_PORT`, ajuste também a `DATABASE_URL` ou a
`API_BASE_URL`.

#### Aplicativo do Discord (login)

O login usa um aplicativo do Discord ([ADR-07](docs/adr/0007-login-discord-sessao.md)). A API
não sobe sem `DISCORD_CLIENT_ID` e `DISCORD_CLIENT_SECRET`.

1. Em https://discord.com/developers/applications, crie um aplicativo (ex.: "RO Lobby (dev)").
2. Na aba **OAuth2**, cadastre os dois Redirects:
   - `http://localhost:3000/auth/discord/callback` (pilha app)
   - `http://localhost:5173/auth/discord/callback` (modo de desenvolvimento)
3. Copie o **Client ID** e gere o **Client Secret**. Coloque os dois no seu `.env`:

   ```dotenv
   DISCORD_CLIENT_ID=<seu Client ID>
   DISCORD_CLIENT_SECRET=<seu Client Secret>
   ```

O Client Secret só existe no `.env`, que o git ignora, e só a API recebe esse valor. Nunca
cole o secret em issue, commit ou conversa; se isso acontecer, gere outro no Developer
Portal (**OAuth2 → Reset Secret**). Os testes e a CI usam um Discord falso e não precisam
de aplicativo.

#### Sala do grupo no Discord (bot, opcional)

O bot do RO Lobby abre, para cada lobby, uma sala privada (categoria com texto e voz) no
servidor oficial, só para os membros do grupo ([ADR-09](docs/adr/0009-bot-discord-rest-outbox.md),
spec `.specs/discord-bot/`). Sem as duas variáveis abaixo, a sala fica desligada e o resto do
RO Lobby funciona normalmente. Só uma delas impede a API de subir.

1. **Servidor oficial.** Crie um servidor no Discord. Com o **Modo de desenvolvedor** ligado
   (Configurações de usuário → Avançado), clique com o botão direito no ícone do servidor e
   use **Copiar ID do servidor**. O ID do servidor oficial é `1560117933705797643`.
2. **Bot.** No mesmo aplicativo do login, abra a aba **Bot**, clique em **Reset Token** e copie
   o token. O usuário do bot tem o mesmo ID do aplicativo (`DISCORD_CLIENT_ID`).
3. **Convite do bot.** Na aba **OAuth2 → URL Generator**, marque o escopo `bot` e as
   permissões **View Channels**, **Manage Channels**, **Manage Roles**, **Send Messages**,
   **Read Message History**, **Connect**, **Speak** e **Create Instant Invite**. Abra a URL
   gerada e adicione o bot ao servidor oficial.
4. **Redirect da conexão.** Na aba **OAuth2**, cadastre também os Redirects da conexão ao
   servidor, que pede `identify` e `guilds.join`:
   - `http://localhost:3000/perfil/discord/retorno` (pilha app)
   - `http://localhost:5173/perfil/discord/retorno` (modo de desenvolvimento)
   - `https://rolobby.com.br/perfil/discord/retorno` (publicação pelo túnel)
5. **Variáveis.** No seu `.env`:

   ```dotenv
   DISCORD_BOT_TOKEN=<token do bot>
   DISCORD_GUILD_ID=1560117933705797643
   ```

   A pilha `app` também usa `APP_ORIGIN` (ou `http://localhost:3000`) para o link do lobby nas
   mensagens do bot.

O token do bot é segredo, como o Client Secret: só existe no `.env` e só a API recebe esse
valor (RNF-01). Nunca cole o token em issue, commit ou conversa. Se ele vazar, gere outro na
aba **Bot → Reset Token** e troque no `.env`. Os testes e a CI usam o Discord falso, com as
rotas do bot, e não precisam do servidor.

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

## Publicar para testes (túnel)

Para mostrar o RO Lobby a amigos, o PC publica a pilha app em `https://rolobby.com.br`
pelo Cloudflare Tunnel ([ADR-08](docs/adr/0008-publicacao-tunel-cloudflare.md)). O túnel
sai do PC até o Cloudflare: nenhuma porta do roteador é aberta e o IP de casa não aparece.

**Limites:** o site só fica no ar com o PC ligado (sem hibernar) e o Docker rodando, e o
banco fica no volume do Docker. O backup é manual, com `bash scripts/backup.sh` (veja
[Segurança](#segurança)).

### Uma vez só

1. **DNS no Cloudflare.** Em dash.cloudflare.com, adicione `rolobby.com.br` no plano Free
   e troque os servidores DNS no registro.br pelos dois nameservers que o Cloudflare
   mostrar. Espere o domínio ficar *Active*.
2. **Túnel.** Em *Zero Trust → Networks → Tunnels*, crie um túnel do tipo *Cloudflared*.
   Copie o token (o texto longo depois de `--token` no comando de instalação). Em *Public
   hostnames*, adicione `rolobby.com.br`, sem subdomínio nem caminho, com o serviço
   `HTTP` e a URL `web:3000`.
3. **Discord.** No aplicativo em discord.com/developers/applications, em *OAuth2 →
   Redirects*, adicione `https://rolobby.com.br/auth/discord/callback`.
4. **`.env`.** Descomente e preencha:
   ```dotenv
   APP_ORIGIN=https://rolobby.com.br
   CLOUDFLARE_TUNNEL_TOKEN=<token do passo 2>
   ```
   Com a origem pública, a pilha só sobe com segredos de verdade: confira com
   `sh scripts/check-env.sh` antes de subir.

### Subir e parar

```powershell
docker compose -f docker-compose.yml -f docker-compose.tunnel.yml --profile app up -d --wait --build
docker compose -f docker-compose.yml -f docker-compose.tunnel.yml --profile app logs -f tunnel
docker compose -f docker-compose.yml -f docker-compose.tunnel.yml --profile app down
```

Com `APP_ORIGIN` preenchido, use o site por `https://rolobby.com.br`, inclusive no PC: os
formulários recusam `http://localhost:3000` (proteção CSRF do SvelteKit). Para voltar ao
modo só local, comente `APP_ORIGIN` e suba sem o arquivo do túnel.

O túnel fica em `docker-compose.tunnel.yml`, fora do arquivo principal, para o modo local
e a CI não dependerem do token. A configuração é conferida por:

```powershell
bash scripts/check-compose-tunnel.sh
```

## Segurança

O repositório é público: tudo o que está nele, qualquer pessoa lê. A segurança do RO Lobby
não depende de esconder o código. Ela depende de os segredos ficarem só no `.env`, que
nunca vai para o git.

### Segredos do `.env`

| Variável | O que é | Onde trocar se vazar |
|---|---|---|
| `DISCORD_CLIENT_SECRET` | Segredo do aplicativo do login | Discord Developer Portal → OAuth2 → Reset Secret |
| `DISCORD_BOT_TOKEN` | Token do bot da sala | Discord Developer Portal → Bot → Reset Token |
| `CLOUDFLARE_TUNNEL_TOKEN` | Token do túnel | Cloudflare Zero Trust → Tunnels → o túnel → Refresh token |
| `APP_POSTGRES_PASSWORD` | Senha do banco da pilha app | Veja *Trocar a senha do banco* abaixo |

Para gerar uma senha forte (só letras e números, porque ela vai dentro da URL do banco):

```powershell
openssl rand -base64 48 | tr -dc 'A-Za-z0-9' | head -c 32
```

O `POSTGRES_PASSWORD` é só do banco de desenvolvimento e pode ficar com o valor do
`.env.example`. A pilha app usa o `APP_POSTGRES_PASSWORD`. Com `APP_ORIGIN` público, o
serviço `env-check` impede a pilha de subir com senha curta (menos de 24 caracteres) ou com
valores do `.env.example`, e diz qual variável corrigir sem mostrar o valor.

**Se um segredo vazar** (commit, print, mensagem): troque o valor no painel da tabela acima,
atualize o `.env` e suba a pilha de novo. Apagar o commit não basta, porque o valor já foi
visto. O GitHub barra o push de tokens conhecidos (*push protection*) e avisa sobre os que
passarem (*secret scanning*).

### Antes de publicar

1. `sh scripts/check-env.sh` sem problemas.
2. Backup do banco: `bash scripts/backup.sh` (com migração nova:
   `bash scripts/backup.sh antes-<número>`). O backup vai para `~/ro-lobby-backups/`, e só os
   10 mais recentes ficam guardados. Os arquivos têm dados dos usuários: não compartilhe nem
   envie para a nuvem.
3. Subir pela `main` atualizada, nunca por uma branch de feature.

### Trocar a senha do banco

O PostgreSQL só lê a senha do ambiente na primeira vez que cria o volume. Num banco que já
existe, a troca é feita dentro dele:

```powershell
bash scripts/backup.sh antes-troca-senha
docker exec -it ro-lobby-app-postgres-1 psql -U ro_lobby -d ro_lobby -c "\password ro_lobby"
# digite a senha nova duas vezes; ela não aparece na tela nem no histórico
```

Depois, ponha a mesma senha em `APP_POSTGRES_PASSWORD` no `.env` e suba a pilha de novo. Se
a API não conectar, a senha do `.env` está diferente da do banco: repita o `\password`.

### O que já vem protegido

- **Rede:** o banco e a API da pilha não publicam porta. O banco de desenvolvimento e o web
  só escutam em `127.0.0.1`, e o túnel chega ao web pela rede interna do Compose.
- **Cabeçalhos:** o web envia `X-Frame-Options`, `X-Content-Type-Options`,
  `Referrer-Policy`, `Permissions-Policy` e, com https, `Strict-Transport-Security`.
- **Limites da API:** corpo de até 64 KiB; 300 requisições por minuto por IP, 10 trocas de
  código de login por minuto por IP e 30 ações por minuto por sessão. Passado o limite, a
  API responde 429 e o site mostra "Muitas tentativas".
- **Dependências:** a CI roda `govulncheck` e `npm audit`, e o Dependabot abre PRs de
  atualização toda semana.
- **Sessões e tokens:** a sessão fica no banco só como hash, o cookie é `HttpOnly`,
  `Secure` e `SameSite=Lax`, e o token do Discord é descartado depois do uso.

### Cloudflare (painel, uma vez só)

Em dash.cloudflare.com, no domínio `rolobby.com.br`:

1. **SSL/TLS → Overview:** modo *Full*.
2. **SSL/TLS → Edge Certificates:** ligar *Always Use HTTPS*.
3. **Security → Bots:** ligar *Bot Fight Mode*.
4. **Security → WAF → Rate limiting rules:** uma regra para o caminho `/auth/` com 20
   requisições por minuto por IP, como segunda barreira antes da API.

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

O ponta a ponta precisa do PostgreSQL no ar (passo 3). Ele usa um banco só dele,
`ro_lobby_e2e`, que o `go run ./cmd/migrate fresh` apaga, recria e migra a cada execução.
O `fresh` só aceita bancos terminados em `_e2e`, para nunca apagar o banco de
desenvolvimento.

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

- **Backend**, quando `backend/` muda: gofmt, golangci-lint, `govulncheck`, testes
  unitários, migrações, testes de integração, build e código gerado.
- **Web**, quando `web/` muda: `npm audit`, ESLint, Prettier, svelte-check, Vitest com
  cobertura, build, tipos gerados e a conferência de que o Client Secret não aparece no
  build.
- **Ponta a ponta**, quando o backend ou o web mudam.
- **Pilha app**, quando o backend, o web ou o Compose mudam: testa o verificador do
  `.env` e o backup, confere o Compose, constrói as imagens e roda o
  `scripts/smoke-app.sh`. Não publica as imagens.
- O `openapi.yaml` e a própria CI disparam todos.
- **CI ok** junta os resultados. PR só de documentação passa direto.

A cobertura aparece no resumo de cada job e como artefato, sem percentual mínimo.

### Proteção da `main`

No GitHub: **Settings → Branches → Add branch ruleset** (ou *Add rule*) para a `main`:

1. Exigir pull request antes do merge.
2. Exigir status checks e marcar **CI ok**. Não marque os jobs Backend, Web e Ponta a
   ponta: eles ficam pulados em PRs que não mexem na pasta deles.
3. Bloquear force push e exclusão da branch.

Sem revisão obrigatória de outra pessoa, porque o projeto tem um mantenedor só. Na aba
**Settings → Code security**, deixe ligados *Secret scanning*, *Push protection*,
*Dependabot alerts* e *Dependabot security updates*.
