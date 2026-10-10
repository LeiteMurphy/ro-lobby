# Tasks — Retrato pela classe

- Spec: `./spec.md` · Design: não há (nível M; as decisões técnicas ficam abaixo)
- Notion: [Épico Retrato pela classe](https://app.notion.com/p/3f5d4a3a5eff8164a870f238ec447500)
- Branch: `feature/retrato-por-classe`

## Decisões técnicas
- **D-01 — Mapa no catálogo.** Cada classe do catálogo em Go ganha o campo `Art`, com o id
  da arte (seção 6). `catalog.ArtOf(classID)` devolve a arte, ou `retrato-1` para classe
  fora do catálogo (RN-04). `GET /classes` passa a mandar `art`.
- **D-02 — A API manda o retrato da classe.** Em toda resposta com personagem, o campo
  `portrait` passa a sair de `catalog.ArtOf(classId)`: personagens, dono e membros do
  lobby, pendentes, pedidos de troca, banco de talentos e "Minhas candidaturas". Assim o
  web não muda nada nas telas que só mostram o retrato (RN-03, RN-05). O enum `Portrait`
  do contrato ganha os 27 ids, e os 4 de hoje continuam válidos.
- **D-03 — Coluna mantida.** `characters.portrait` continua no banco, para a migração de
  volta não perder dado, mas deixa de ser lida. O cadastro grava o padrão e ignora o campo
  `portrait` da entrada, que fica obsoleto no contrato (RN-02).
- **D-04 — Arquivos.** As 27 artes vão para `web/static/portraits/classes/<id>.png`. O
  gerador vai para `scripts/class-art/`, para a arte poder ser refeita. No web,
  `portraitInfo(id)` aponta para o PNG da classe ou para o SVG genérico e devolve o nome da
  linha como texto alternativo.

## Backend

### T-01 — Arte da classe no catálogo e em toda resposta com personagem  [x]
- Cobre: US-01, US-02, RN-01 a RN-05, CA-01.1, CA-01.2, CA-01.3, CA-01.5, CA-01.6, CA-02.1,
  CA-02.2 (API), D-01 a D-03
- Depende de: —
- Paralelizável: não
- Arquivos: `backend/internal/catalog/`, `backend/internal/server/`,
  `backend/internal/characters/`, `openapi.yaml`, gerados
- Pronto quando: testes provam que toda classe do catálogo tem uma arte da lista, a
  tabela da seção 6, a classe fora do catálogo em `retrato-1`, e o `portrait` pela classe
  nas respostas de personagem, lobby, troca, talentos e candidaturas.
- Commit: `9ae85b0`

## Web

### T-02 — Artes no web, fim da escolha de retrato  [x]
- Cobre: US-01, US-02, RN-02, RN-05, RN-06, CA-01.1, CA-01.2 (e2e), CA-01.4, CA-02.1, CA-02.2, RNF-01,
  RNF-02, D-04
- Depende de: T-01
- Paralelizável: não
- Arquivos: `web/static/portraits/classes/`, `scripts/class-art/`,
  `web/src/lib/characters/` (portraits, CharacterDialog, CharacterCard)
- Pronto quando: testes de SSR e e2e provam que o card mostra a arte da linha com o texto
  alternativo, que o diálogo não tem mais a escolha de retrato e mostra a arte da classe
  escolhida, que trocar a classe troca a arte, e que o lobby e o banco de talentos mostram
  as artes; cada PNG tem no máximo 4 KB.
- Commit: `1c5da8c`

## Matriz de cobertura
| Critério | Tasks |
|---|---|
| CA-01.1 | T-01, T-02 |
| CA-01.2 | T-01, T-02 |
| CA-01.3 | T-01 |
| CA-01.4 | T-02 |
| CA-01.5 | T-01 |
| CA-01.6 | T-01 |
| CA-02.1 | T-01, T-02 |
| CA-02.2 | T-01, T-02 |

## Seção 6 — Classe → arte
| Arte | Classes |
|---|---|
| `superaprendiz` | aprendiz, superaprendiz, superaprendiz-ex, hiperaprendiz |
| `espadachim` | espadachim |
| `cavaleiro` | cavaleiro, lorde, cavaleiro-runico, cavaleiro-draconiano |
| `templario` | templario, paladino, guardiao-real, guardiao-imperial |
| `mago` | mago |
| `bruxo` | bruxo, arquimago, arcano, magus |
| `sabio` | sabio, professor, feiticeiro, elementalista |
| `gatuno` | gatuno |
| `mercenario` | mercenario, algoz, sicario, executor |
| `arruaceiro` | arruaceiro, desordeiro, renegado, mandraque |
| `mercador` | mercador |
| `ferreiro` | ferreiro, mestre-ferreiro, mecanico, engenheiro |
| `alquimista` | alquimista, criador, bioquimico, cientista |
| `novico` | novico |
| `sacerdote` | sacerdote, sumo-sacerdote, arcebispo, cardeal |
| `monge` | monge, mestre, shura, inquisidor |
| `arqueiro` | arqueiro |
| `cacador` | cacador, atirador-de-elite, sentinela, falcao-do-vento |
| `bardo` | bardo, menestrel, trovador, maestro |
| `odalisca` | odalisca, cigana, musa, diva |
| `taekwon` | taekwon |
| `mestre-taekwon` | mestre-taekwon, mestre-estelar, mestre-celestial |
| `espiritualista` | espiritualista, ceifador-de-almas, asceta-das-almas |
| `ninja` | ninja, kagerou, oboro, shinkiro, shiranui |
| `justiceiro` | justiceiro, insurgente, guerrilheiro |
| `invocador` | invocador, animista |
| `druida` | druida, karnos, alitea |

## Descobertas
- D-01 ficou como um mapa `classArt` no pacote `catalog`, ao lado da lista de classes, em
  vez de um campo `Art` em cada classe: a tabela da seção 6 fica legível num lugar só. O
  efeito é o mesmo (`catalog.ArtOf`, `art` em `GET /classes`).
- RN-05 pede o nome da linha como texto alternativo em todo lugar. As listas (lobby,
  candidaturas, trocas, formulário) usavam `alt=""`; agora usam `portraitAlt`, que dá o
  nome da linha.
