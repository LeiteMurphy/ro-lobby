# Spec — Endurecimento de segurança

- Feature: `endurecimento` · Nível: M · Status: Aprovada (2026-10-15)
- Notion: https://app.notion.com/p/3fad4a3a5eff81e1a52de5a23728a9ec
- Última revisão: 2026-10-15 — aprovada; os limites da RN-15 são por minuto e não mexem na
  duração da sessão (RN-09 da login-discord continua igual). Nas tasks, o limite de login
  passou a contar a troca do código, porque o início não chega à API (RN-15, CA-05.3)
- Revisão: 2026-10-15 — RN-15 e CA-05.2 aprovados pelo usuário depois da validação do
  ciclo 1: as ações contam por sessão e incluem toda requisição logada que muda algo
  (exceto login e logout); o `/healthz` fica fora do limite geral

## 1. Contexto
O repositório do RO Lobby é público, e a v1 roda no PC do usuário, publicada em
`rolobby.com.br` pelo Cloudflare Tunnel (ADR-08), com usuários reais. O research
(`research.md`) não achou segredo vazado, mas achou pontos fracos:
- a senha do banco de produção é a senha pública do `.env.example`;
- há portas abertas para a rede local;
- a API foi compilada com um Go que tem vulnerabilidades conhecidas;
- o site não envia cabeçalhos de segurança;
- não há limite de tamanho nem de frequência nas requisições;
- o GitHub está sem proteções contra segredo commitado.

Esta entrega fecha esses pontos sem mudar o comportamento do produto para quem usa direito.

## 2. Atores
| Ator | Descrição | Permissões principais |
|---|---|---|
| Operador | O dono do projeto, que mantém o `.env` e publica a pilha `app` no PC | Edita o `.env`, sobe a pilha, faz backup, configura o GitHub |
| Usuário | Jogador logado no RO Lobby | Usa o site normalmente |
| Visitante | Qualquer pessoa na internet, inclusive quem lê o repositório público | Vê as páginas públicas e o código |

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como operador, quero que a produção use uma senha de banco própria e forte, e que a pilha pública não suba com valores de exemplo, para que o repositório público não entregue a chave do banco. |
| US-02 | P1 | Como operador, quero que o banco de dev e o web só escutem no meu PC, para que outros aparelhos da rede não cheguem neles por fora do túnel. |
| US-03 | P1 | Como usuário, quero que a API rode num Go sem vulnerabilidades conhecidas e que a CI avise sobre vulnerabilidades novas, para meus dados não dependerem de falha já corrigida. |
| US-04 | P1 | Como usuário, quero que o site envie cabeçalhos de segurança, para outro site não poder me enganar exibindo o RO Lobby dentro de uma moldura. |
| US-05 | P2 | Como usuário, quero que a API recuse requisições gigantes ou repetidas demais, para o site continuar no ar mesmo com alguém abusando. |
| US-06 | P1 | Como operador, quero que o GitHub barre e avise sobre segredos e dependências vulneráveis, e que a `main` só receba código com a CI verde, para um erro meu não virar vazamento. |
| US-07 | P2 | Como operador, quero um comando de backup com retenção, para guardar só o necessário dos dados dos usuários (LGPD). |
| US-08 | P1 | Como operador, quero um verificador do `.env` e um checklist de segurança no README, para saber se o meu ambiente está seguro antes de publicar. |

## 4. Regras

### Segredos e `.env`
- **RN-01** — A pilha `app` usa a senha `APP_POSTGRES_PASSWORD`, separada da
  `POSTGRES_PASSWORD` do modo de desenvolvimento. A pilha `app` não sobe sem ela.
- **RN-02** — Com origem pública (`APP_ORIGIN` começando com `https://`), a pilha `app` se
  recusa a subir se `APP_POSTGRES_PASSWORD` tiver menos de 24 caracteres ou se
  `APP_POSTGRES_PASSWORD` ou `DISCORD_CLIENT_SECRET` forem iguais aos valores do
  `.env.example`. A mensagem diz qual variável corrigir, sem mostrar o valor.
- **RN-03** — Sem origem pública (localhost e CI), a pilha `app` aceita os valores do
  `.env.example`, para o modo local e a CI continuarem funcionando com
  `cp .env.example .env`.
- **RN-04** — O verificador do `.env` (`scripts/check-env.sh`):
  - aponta variável obrigatória faltando, valor de exemplo em uso com origem pública,
    senha curta e bot pela metade (só o token ou só o servidor, RN-22 da discord-bot);
  - termina com código diferente de zero se houver problema;
  - nunca imprime o valor de nenhuma variável.
- **RN-05** — Nenhum segredo real (senhas, tokens, Client Secret) aparece em arquivo
  versionado, log, mensagem de erro ou saída de script.

### Rede
- **RN-06** — A porta do PostgreSQL de desenvolvimento e a porta do web da pilha `app` são
  publicadas só em `127.0.0.1`. O túnel continua chegando ao web pela rede interna do
  Compose.
- **RN-07** — O banco e a API da pilha `app` continuam sem porta publicada (RN-04 da
  publicacao-tunel).

### Dependências
- **RN-08** — A API é compilada com um Go fixo, sem vulnerabilidades conhecidas da
  biblioteca padrão alcançáveis pelo código. A imagem de build usa a versão exata, sem tag
  móvel.
- **RN-09** — A CI falha se o `govulncheck` achar vulnerabilidade alcançável no backend, ou
  se o `npm audit` achar vulnerabilidade alta ou crítica nas dependências de produção do web.
- **RN-10** — O Dependabot verifica Go, npm, GitHub Actions e imagens Docker uma vez por
  semana.

### Cabeçalhos do web
- **RN-11** — Toda resposta do web leva:
  - `X-Frame-Options: DENY`;
  - `X-Content-Type-Options: nosniff`;
  - `Referrer-Policy: strict-origin-when-cross-origin`;
  - `Permissions-Policy: camera=(), microphone=(), geolocation=()`.
- **RN-12** — Com origem `https://`, as respostas também levam
  `Strict-Transport-Security: max-age=31536000`. Em `http://localhost`, esse cabeçalho não
  é enviado.

### Limites da API
- **RN-13** — A API recusa com 413 um corpo de requisição maior que 64 KiB.
- **RN-14** — O servidor da API tem limite de tempo para ler a requisição (15 s), escrever a
  resposta (30 s) e manter conexão ociosa (120 s), além do limite de 5 s já existente para
  os cabeçalhos.
- **RN-15** — A API limita a frequência de requisições:
  - **login** (a troca do código no retorno do Discord; o início só redireciona e não
    chega à API): 10 por minuto por IP;
  - **ações** (toda requisição de quem está logado que muda algo — criar, editar, excluir
    ou decidir — exceto o login e o logout): 30 por minuto por sessão. Quem está logado em
    mais de um aparelho tem um limite em cada um;
  - **geral**: 300 por minuto por IP em qualquer rota, exceto a verificação de saúde
    (`/healthz`).

  Passado o limite, a resposta é 429 com `Retry-After`, e o web mostra "Muitas tentativas.
  Aguarde um pouco e tente de novo.". O intervalo próprio do "Avisar no Discord" (RN-27 da
  discord-bot) continua valendo, e o aviso também conta como ação.
- **RN-16** — O IP de quem acessa é o que o Cloudflare informa (`CF-Connecting-IP`) quando
  a requisição chega pelo túnel. Fora do túnel, é o endereço da conexão. O web repassa
  esse IP à API, e a API só aceita o repasse vindo da rede interna (a API não tem porta
  publicada, RN-07).

### GitHub e operação
- **RN-17** — O repositório tem ligados:
  - secret scanning e push protection;
  - alertas e atualizações de segurança do Dependabot;
  - proteção da `main`: exige o check `ci-ok` verde, bloqueia force push e exclusão, e não
    exige revisão de outra pessoa.
- **RN-18** — O backup (`scripts/backup.sh`) faz o `pg_dump` do banco da pilha `app` em
  `~/ro-lobby-backups/` e mantém só os 10 arquivos mais recentes. Ele não imprime dados dos
  usuários.
- **RN-19** — A senha de produção é trocada no banco existente (`ALTER ROLE`) só depois de
  um backup, e a pilha volta a subir com a senha nova sem perder dados.

## 5. Critérios de aceite

### US-01 — Senha própria da produção
```gherkin
CA-01.1 — Pilha pública com senha de exemplo não sobe  [US-01, RN-02]
Given um .env com APP_ORIGIN=https://rolobby.com.br
  And APP_POSTGRES_PASSWORD igual ao valor do .env.example
When o operador sobe a pilha app
Then a subida falha antes das migrações
  And a mensagem cita APP_POSTGRES_PASSWORD sem mostrar o valor

CA-01.2 — Senha curta é recusada com origem pública  [US-01, RN-02]
Given APP_ORIGIN=https://rolobby.com.br e APP_POSTGRES_PASSWORD com 23 caracteres
When o operador sobe a pilha app
Then a subida falha citando o mínimo de 24 caracteres

CA-01.3 — Client Secret de exemplo é recusado com origem pública  [US-01, RN-02]
Given APP_ORIGIN=https://rolobby.com.br e DISCORD_CLIENT_SECRET igual ao .env.example
When o operador sobe a pilha app
Then a subida falha citando DISCORD_CLIENT_SECRET

CA-01.4 — Localhost e CI aceitam o exemplo  [US-01, RN-03]
Given um .env copiado do .env.example, sem APP_ORIGIN
When a pilha app sobe
Then todos os serviços ficam saudáveis

CA-01.5 — Produção com senha nova mantém os dados  [US-01, RN-19]
Given o banco de produção com os usuários atuais e um backup recém-feito
When o operador troca a senha e sobe a pilha com APP_POSTGRES_PASSWORD nova
Then a API conecta e os mesmos usuários, personagens e lobbies continuam lá
```

### US-02 — Portas só no PC
```gherkin
CA-02.1 — Banco de dev só em 127.0.0.1  [US-02, RN-06]
Given o modo de desenvolvimento no ar
When se lista a porta publicada do serviço postgres
Then ela é 127.0.0.1:5432 e não 0.0.0.0

CA-02.2 — Web da pilha só em 127.0.0.1 e o túnel funciona  [US-02, RN-06, RN-07]
Given a pilha app com o túnel no ar
When se lista a porta publicada do web
Then ela é 127.0.0.1:3000
  And https://rolobby.com.br responde 200
  And app-postgres e api não têm porta publicada
```

### US-03 — Dependências sem vulnerabilidade conhecida
```gherkin
CA-03.1 — Go fixo e sem vulnerabilidade alcançável  [US-03, RN-08]
Given o backend
When se roda o govulncheck
Then nenhuma vulnerabilidade alcançável é relatada
  And o Dockerfile usa a imagem do Go com versão exata

CA-03.2 — CI barra vulnerabilidade  [US-03, RN-09]
Given o workflow de CI
When o govulncheck ou o npm audit --omit=dev --audit-level=high falham
Then o job falha e o ci-ok fica vermelho

CA-03.3 — Dependabot configurado  [US-03, RN-10]
Given o arquivo .github/dependabot.yml
Then ele cobre gomod (backend), npm (web), github-actions e docker, com frequência semanal
```

### US-04 — Cabeçalhos de segurança
```gherkin
CA-04.1 — Cabeçalhos em toda página  [US-04, RN-11]
Given o web no ar
When um visitante pede a Home, uma página de lobby ou uma rota inexistente
Then a resposta tem X-Frame-Options DENY, X-Content-Type-Options nosniff,
     Referrer-Policy strict-origin-when-cross-origin e a Permissions-Policy da RN-11

CA-04.2 — HSTS só com https  [US-04, RN-12]
Given o web com origem https://rolobby.com.br
When chega qualquer resposta
Then ela tem Strict-Transport-Security max-age=31536000
  And com origem http://localhost:3000 o cabeçalho não aparece
```

### US-05 — Limites da API
```gherkin
CA-05.1 — Corpo grande demais  [US-05, RN-13]
Given um usuário logado
When ele envia à API um corpo de 65 KiB
Then a resposta é 413 e nada é gravado

CA-05.2 — Limite de ações por sessão  [US-05, RN-15]
Given uma sessão que fez 30 ações no último minuto
When ela faz a 31ª
Then a resposta é 429 com Retry-After
  And o web mostra "Muitas tentativas. Aguarde um pouco e tente de novo."
  And outra sessão no mesmo minuto não é afetada

CA-05.3 — Limite de login por IP  [US-05, RN-15, RN-16]
Given um IP que concluiu a troca do código de login 10 vezes no último minuto
When ele tenta a 11ª
Then a API responde 429 e o web leva ao aviso de login com falha
  And outro IP consegue entrar normalmente

CA-05.4 — O limite volta depois da janela  [US-05, RN-15]
Given um usuário bloqueado pelo limite de ações
When passa o tempo do Retry-After
Then a próxima ação é aceita

CA-05.5 — IP vem do Cloudflare só pelo túnel  [US-05, RN-16]
Given duas requisições vindas pelo túnel com CF-Connecting-IP diferentes
Then a API conta cada uma no IP informado
  And uma requisição direta à API com um cabeçalho de IP forjado, vinda de fora da rede
      interna, não é possível porque a API não tem porta publicada

CA-05.6 — Timeouts do servidor  [US-05, RN-14]
Given a configuração do servidor da API
Then ReadTimeout é 15 s, WriteTimeout é 30 s, IdleTimeout é 120 s e ReadHeaderTimeout é 5 s
```

### US-06 — Proteções do GitHub
```gherkin
CA-06.1 — Segurança do repositório ligada  [US-06, RN-17]
Given o repositório LeiteMurphy/ro-lobby
When se consulta a configuração pela API do GitHub
Then secret scanning, push protection e as atualizações de segurança do Dependabot estão
     enabled

CA-06.2 — main protegida  [US-06, RN-17]
Given a branch main
When se consulta a proteção
Then o check ci-ok é obrigatório, force push e exclusão estão bloqueados
  And nenhuma revisão aprovada é exigida
```

### US-07 — Backup com retenção
```gherkin
CA-07.1 — Backup e retenção  [US-07, RN-18]
Given 10 backups em ~/ro-lobby-backups
When o operador roda scripts/backup.sh
Then um arquivo novo é criado com o banco da pilha app
  And o mais antigo é apagado, sobrando 10
  And a saída do script não contém dados de usuários
```

### US-08 — Verificador do `.env` e checklist
```gherkin
CA-08.1 — .env bom passa  [US-08, RN-04]
Given um .env completo, com origem pública e segredos fora do exemplo
When o operador roda scripts/check-env.sh
Then o script termina com código 0 e diz que o .env está pronto

CA-08.2 — .env com problemas é apontado sem vazar valor  [US-08, RN-04, RN-05]
Given um .env com origem pública, senha de exemplo e só o DISCORD_BOT_TOKEN preenchido
When o operador roda scripts/check-env.sh
Then o script lista os três problemas pelo nome da variável
  And termina com código diferente de zero
  And nenhum valor do .env aparece na saída

CA-08.3 — Checklist no README  [US-08]
Given o README
Then há uma seção de segurança com: gerar segredos fortes, rodar o check-env, fazer backup
     antes de publicar, o que fazer se um segredo vazar (trocar no Discord e no Cloudflare)
     e as configurações recomendadas do Cloudflare (Always Use HTTPS, SSL Full, regra de
     limite em /auth)
```

## 6. Casos de borda
- O operador esqueceu de criar `APP_POSTGRES_PASSWORD` no `.env` → a pilha `app` não sobe,
  com mensagem pedindo a variável [RN-01].
- Troca de senha no meio do caminho (`ALTER ROLE` feito, mas o `.env` não atualizado) → a API
  não conecta. O passo a passo do README manda atualizar o `.env` e subir de novo, e o
  backup permite voltar [RN-19].
- Vários usuários atrás do mesmo IP (casa, lan house) → o limite geral por IP (300/min)
  comporta; as ações contam por sessão [RN-15].
- A API reinicia → os contadores de limite zeram. Isso é aceito, porque os contadores ficam
  em memória [RN-15].
- Uma rota de vulnerabilidade do Go só alcançável pelo Discord falso de teste → conta como
  alcançável e precisa ser resolvida [RN-08].

## 7. Requisitos não funcionais
- **RNF-01** — Os testes e a CI continuam rodando sem segredos reais.
- **RNF-02** — O limite de requisições não adiciona dependência externa (sem Redis).
- **RNF-03** — Nada nesta entrega muda o que o usuário vê ao usar o site normalmente.

## 8. Fora de escopo
- E-mail pessoal nos commits antigos: exige reescrever o histórico. O usuário pediu para
  voltar nisso depois.
- Página de privacidade e exclusão de conta (LGPD): vira uma feature própria.
- Usuário do banco com menos poder (A-10): o usuário pediu uma explicação antes de decidir.
  Pode entrar numa revisão desta spec.
- Content-Security-Policy (CSP): o usuário pediu uma explicação antes de decidir. Pode entrar
  numa revisão desta spec.
- Configurações do painel do Cloudflare: ficam como checklist no README, porque não são
  versionáveis aqui.

## 9. Perguntas em aberto
- [ ] Papel do banco sem superusuário (A-10): explicar ao usuário e decidir.
- [ ] CSP: explicar ao usuário e decidir.

## 10. Decisões tomadas na entrevista
- Senha de produção separada → sim, e eu gero a senha no `.env` sem mostrar e troco no banco
  com backup antes.
- Papel do banco com menos poder → separar, explicar ao final.
- Limite de requisições → no código, na API, por usuário e por IP. O Cloudflare entra como
  checklist.
- CSP → explicar ao final antes de decidir. Os outros cabeçalhos entram.
- GitHub (secret scanning, push protection, Dependabot, proteção da `main` sem revisão
  obrigatória) e backup com retenção de 10 → sim.
