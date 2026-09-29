# ADR-04 — Acesso ao banco e migrações: `pgx`, `sqlc` e `goose`

- Status: Aceito
- Data: 2026-09-29

## Contexto
O banco é PostgreSQL, com datas em UTC. As specs pedem operações atômicas, por exemplo
dois aceites concorrentes que não podem lotar uma função (candidatura, RNF-03). Por
isso é preciso controlar as transações e o SQL de perto, e o código de acesso tem de
ser testável contra um Postgres real.

## Opções consideradas
- (a) `pgx` + `sqlc` + `goose`: SQL escrito à mão, código Go tipado gerado a partir
  dele, migrações em arquivos SQL versionados.
- (b) GORM: ORM com migração automática.
- (c) `database/sql` puro, com SQL e mapeamento escritos à mão.

## Decisão
(a) `pgx` como driver, `sqlc` para gerar o código de consulta e `goose` para as
migrações.

## Consequências
- \+ O SQL fica explícito e revisável, e dá para usar `SELECT ... FOR UPDATE` e
  transações quando a regra pedir atomicidade.
- \+ Um erro de tipo entre o SQL e o Go aparece na geração, não em produção.
- \+ As migrações são arquivos SQL versionados, com `up` e `down`.
- − Consulta dinâmica (filtros opcionais) exige mais trabalho do que num ORM.
- − O código gerado precisa ficar sincronizado com o SQL, então a CI confere.
