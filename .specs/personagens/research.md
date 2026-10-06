# Research — Perfil e personagens

- Feature: `personagens`
- Data: 2026-10-05

## Objetivo
Dar ao Usuário logado uma página de perfil onde ele cadastra os personagens com que vai
criar lobbies e se candidatar.

## Achados
- **Produto (CLAUDE.md):** o personagem tem nick, classe, nível, função, link externo,
  horários e instâncias de interesse. Domínio: Personagem, Disponibilidade e Instância
  (catálogo).
- **Premissas da spec `candidatura-lobby`:**
  - P-04: o personagem pertence a um único Usuário e tem uma função (Tank, Suporte ou
    Dano);
  - RN-25: a função não pode mudar com candidatura pendente ou vaga ocupada;
  - RN-26: excluir o personagem cancela as candidaturas dele.

  Como lobbies e candidaturas ainda não existem, essas duas regras não têm efeito agora,
  mas a modelagem precisa permitir que elas entrem depois.
- **Login (`login-discord`):** existe Usuário, sessão e `GET /me`. A rota de personagens
  usa a mesma sessão (`Authorization: Bearer`). O PR #13 ainda não está na `main`, então
  esta branch sai da `feature/login-discord`.
- **Catálogo de classes:** `web/src/lib/catalog/classes.ts`, com as 82 classes do bROWiki
  (nome, nível da classe, linha, ícone). Hoje ele só existe no web.
- **Instâncias:** só existem os 5 nomes inventados dos dados fictícios da Home. Não há
  catálogo de instâncias.
- **Claude Design:**
  - Não há tela de perfil desenhada no nível do Home v2.
  - O kit de telas do design system tem uma tela "Personagens" (título, botão "Adicionar
    personagem" e lista de `CharacterCard`) e um diálogo "Adicionar personagem" (nome,
    classe digitada, nível e função principal).
  - O `CharacterCard` mostra o quadrado da função, o nome, o selo "Principal", a função,
    a classe e o nível. O exemplo dos dados ainda mostra um campo "servidor" (BR).
- **Contrato e banco:** OpenAPI (ADR-05) e `pgx` + `sqlc` + `goose` (ADR-04). A próxima
  migração é a `00003`.

## Classificação: G (Grande)
Novas tabelas e regras de domínio, CRUD com autorização por dono, catálogo compartilhado
entre web e API e uma tela nova. Pode crescer bastante se Disponibilidade e instâncias de
interesse entrarem juntas, então o escopo é a primeira pergunta.
