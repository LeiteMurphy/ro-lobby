# Research — Lobbies

- Feature: `lobbies`
- Data: 2026-10-06

## Objetivo
Permitir que o Usuário crie um lobby (grupo para uma instância, num horário, com vagas por
função) e trocar os dados fictícios da Home pelos lobbies reais.

## Achados
- **Premissas da spec `candidatura-lobby` (aprovada) que esta feature precisa cumprir:**
  - P-01: o lobby tem um dono, início e fim em UTC e vagas por função, com no máximo 12
    no total;
  - P-02: o dono escolhe um dos seus personagens, que ocupa uma vaga da função dele desde
    a criação;
  - P-03: o lobby está *aberto* (antes do início e não cancelado), *iniciado* ou
    *cancelado*.
- **Home (`home-local`):**
  - os lobbies vêm de um módulo único de dados fictícios (`web/src/lib/home/fixtures.ts`),
    acessado por uma função que a feature de lobby troca pela API (RN-08);
  - o card mostra horário, instância, vagas ocupadas/total, anfitrião, classe, nível
    mínimo e composição por função;
  - filtros por instância, vaga, nível mínimo e faixa de horário; 14 dias no seletor;
    fuso `America/Sao_Paulo` (RN-09);
  - "Criar lobby", "Candidatar" e "Ver grupo" ainda estão "Disponível em breve" (RN-18).
- **Instâncias:** não há catálogo. A Home usa 5 nomes inventados, com capa e ícone
  próprios. A regra do `CLAUDE.md` libera nomes de classe do bROWiki, mas não fala de
  nomes de instância.
- **Personagens (`personagens`):** prontos, com função, nível e principal. Personagem tem
  dono único; a API filtra por dono.
- **Claude Design:** o kit do design system tem as telas de detalhe do lobby e os
  diálogos, ainda não desenhados no nível do Home v2.
- **Banco e contrato:** próxima migração é a `00004`; OpenAPI com oapi-codegen (enums com
  prefixo do tipo).

## Classificação: G (Grande)
Entidade nova com ciclo de vida (aberto, iniciado, cancelado), regras de vagas e horário,
catálogo de instâncias, telas novas (criar e detalhe) e a troca da fonte de dados da Home.
A candidatura fica fora: ela tem spec própria e entra depois.
