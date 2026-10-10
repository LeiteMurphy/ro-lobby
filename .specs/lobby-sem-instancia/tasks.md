# Tasks — Lobby sem instância

- Spec: `./spec.md` · Design: não há (nível M; as decisões técnicas ficam abaixo)
- Notion: [Épico](https://app.notion.com/p/3f5d4a3a5eff81bb924bcb3af470ab83)
- Branch: `feature/lobby-sem-instancia`

## Decisões técnicas
- **D-01 — Instância nula no lobby.** `instance_id` e `instance_name` passam a aceitar nulo:
  - `instance_id` nulo é "sem instância";
  - `instance_name` guarda o título, ou fica nulo sem título;
  - `instance_level` fica 1, então o CHECK `min_level >= instance_level` continua valendo
    (RN-03);
  - um CHECK novo limita o título a 1–40 caracteres quando não há instância (RN-02).
  Os lobbies existentes não mudam (RNF-01).
- **D-02 — Contrato.** `LobbyInput` e `LobbyUpdate` ganham `anyInstance` (boolean) e `title`.
  Com `anyInstance`, o `instanceId` é ignorado. No `Lobby`, `instance.id` fica nulo, e
  `instance.name` traz o título ou "Qualquer instância". Assim o card, o detalhe, o convite
  e o preview mostram o nome sem lógica nova (RN-04).
- **D-03 — Afinidade.** A consulta de afinidade recebe `instance_id` vazio para o lobby sem
  instância, e a condição de instância vale para todos (RN-07).
- **D-04 — Filtro da Home.** O lobby da Home ganha `anyInstance`. O filtro tem o valor
  especial `__none__` ("Sem instância definida"), e a lista de instâncias do filtro deixa
  de fora os títulos livres (RN-05).

## Backend

### T-01 — Lobby sem instância na API, no banco e na afinidade  [x]
- Cobre: US-01, US-03, RN-01, RN-02, RN-03, RN-06, RN-07, RN-08, CA-01.1 a CA-01.4 (API),
  CA-03.1 (API), CA-04.1, RNF-01, D-01 a D-03
- Depende de: —
- Paralelizável: não
- Arquivos: `backend/migrations/00009_lobby_any_instance.sql`, `backend/queries/`,
  `backend/internal/lobbies/`, `backend/internal/talents/`, `backend/internal/server/`,
  `openapi.yaml`, gerados
- Pronto quando: a migração aplica e reverte com um lobby antigo; testes de integração
  provam criar sem instância com e sem título, recusar título de 41, nível de 1 até o do
  dono, trocar de instância para sem instância e de volta na edição, e afinidade com
  personagem de outra instância.
- Commit: 42363ae

## Web

### T-02 — Formulário, nome no card e filtro da Home  [x]
- Cobre: US-01, US-02, US-03, RN-01 a RN-06, CA-01.1 a CA-01.5, CA-02.1, CA-03.1, RNF-03,
  D-04
- Depende de: T-01
- Paralelizável: não
- Arquivos: `web/src/lib/lobbies/` (form, LobbyForm, toHome), `web/src/lib/home/`
  (types, filters, FilterPanel), `routes/lobbies/novo`, `routes/lobbies/[id]/editar`
- Pronto quando: testes de SSR e e2e provam a opção "Sem instância definida" com o título
  e o nível 1, o padrão com instância, o título no card, no detalhe e no convite, o
  filtro "Sem instância definida" e a troca na edição.
- Commit: e3a1e4a

## Matriz de cobertura
| Critério | Tasks |
|---|---|
| CA-01.1 | T-01, T-02 |
| CA-01.2 | T-01, T-02 |
| CA-01.3 | T-01 |
| CA-01.4 | T-01, T-02 |
| CA-01.5 | T-02 |
| CA-02.1 | T-02 |
| CA-03.1 | T-01, T-02 |
| CA-04.1 | T-01 |

## Descobertas
- (nenhuma)
