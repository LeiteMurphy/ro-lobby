# Research — Servidor do jogo

- Feature: `servidor`
- Data: 2026-10-08

## Objetivo
O Ragnarok LATAM tem dois servidores, Freya e Nidhogg (grafia a confirmar). Um personagem
existe em um só servidor, e um grupo só joga junto se todos estiverem no mesmo. O RO Lobby
precisa saber o servidor do personagem e do lobby, e a Home precisa filtrar por ele.

## Achados
- **Personagens (`personagens`):**
  - tabela `characters` (migração `00003`) sem servidor;
  - RN-05: nick único em todo o RO Lobby, sem diferenciar maiúsculas (índice
    `characters_nick_key`). No jogo, o nick é único por servidor, então a regra pode
    passar a valer por servidor;
  - RN-16: a carteirinha mostra retrato, nick, classe, nível, função e link;
  - RN-25 da `candidatura-lobby`: personagem com vínculo ativo não muda função nem nível.
    O servidor seria um terceiro campo travado.
- **Lobbies (`lobbies`):**
  - tabela `lobbies` (migração `00004`) sem servidor. `owner_character_id` é
    `ON DELETE SET NULL`, então o servidor do lobby precisa ser guardado no próprio lobby,
    e não só lido do personagem do dono;
  - RN-08: o dono escolhe um dos seus personagens; RN-17: instância e personagem do dono
    não mudam na edição;
  - RN-22: o card da Home mostra horário, instância, vagas, anfitrião, classe, nível
    mínimo e composição.
- **Candidatura (`candidatura-lobby`):** RN-01 (personagem do candidato), RN-19 (dono
  troca o próprio personagem), RN-20/RN-23/RN-36 (pedido de troca do membro). Todas
  passam a exigir o mesmo servidor do lobby.
- **Home (`home-local`):** RN-12 e RN-13 definem os filtros (instância, vaga, nível
  mínimo, faixa de horário) e as contagens. O filtro de servidor entra ali
  (`web/src/lib/home/filters.ts`).
- **Dados existentes:** o projeto só roda local (ADR-06), sem produção. A migração pode
  dar um valor padrão aos personagens e lobbies que já existem.
- **Grafia:** a busca web não trouxe fonte oficial com os nomes dos servidores. O usuário
  escreveu "Niddhog"; a forma mitológica é "Nidhogg". Confirmar com o usuário.

## Arquivos-chave
- `backend/migrations/00003_characters.sql`, `00004_lobbies.sql` (nova migração `00010`)
- `backend/queries/characters.sql`, `lobbies.sql`, `applications.sql`, `swap_requests.sql`
- `backend/internal/{characters,lobbies,applications,catalog}`
- `openapi.yaml` e `web/src/lib/api/schema.gen.ts`
- `web/src/lib/characters`, `web/src/lib/lobbies`, `web/src/lib/home/filters.ts`,
  `web/src/routes/{perfil,lobbies,candidaturas}`

## Classificação
**M (Médio).** Um campo novo em duas tabelas e regras novas em três specs, passando por
API, contrato e web. Não há decisão de arquitetura: o servidor é um catálogo fixo como
função. Fases: Specify → Tasks → Execute → Validate.

## Atualização de 2026-10-14
Retomada depois do `main` avançar. Entraram nesse meio-tempo, e a spec passa a cobrir:
- banco de talentos (afinidade, `/talentos`);
- compartilhar lobby (convite e Open Graph);
- publicação em `rolobby.com.br` (ADR-08). Os dados que já existem são de jogadores reais,
  então a migração marca os personagens "a conferir".

A nova migração é a `00010` (a `discord-bot` fica com a `00011`). A grafia "Nidhogg" foi
confirmada.
