# Research — Hospedagem local e Home

- Feature: `home-local`
- Data: 2026-09-30

## Objetivo
Rodar o projeto inteiro em `localhost` como se fosse um servidor, com a Home desenhada
no Claude Design no lugar da página de status provisória.

## Achados
- **Design:** projeto "RO Lobby Home v2" no Claude Design
  (`claude.ai/design/p/6b28b143-4bcb-4fa0-a225-b0abbb41dad7`), arquivo `Home v2.dc.html`,
  com um design system próprio ("RO Lobby Design System"). Não usa a marca Credifit.
- **Tela Home v2:**
  - barra superior com símbolo e "RO Lobby", botão "Criar lobby" (ícone no celular) e
    o personagem conectado;
  - seletor de 14 dias, com a contagem de grupos por dia;
  - barra lateral de filtros: instância, vaga para (Tank, Suporte, Dano, com contagem),
    nível mínimo e faixa de horário (com contagem). Abaixo de 900 px, vira uma gaveta à
    esquerda;
  - destaque "Próximo grupo com vaga de <função>", com capa da instância, horário,
    tempo relativo, anfitrião, nível mínimo, composição e as ações "Candidatar como
    <função>" e "Ver grupo";
  - grade de cards de lobby: capa por instância, horário, tempo relativo, nome da
    instância, vagas ocupadas/total, anfitrião com a classe, nível mínimo, composição
    por função e a ação "Candidatar", o selo "Candidatura pendente" ou o selo "Lotado";
  - borda esquerda do card na cor da função com mais vagas, e cinza quando lotado;
  - estados vazios: "Nenhum grupo neste dia" e "Nenhum grupo com esses filtros".
- **Design system:**
  - tokens de cor, tipografia, espaçamento, movimento e base em CSS;
  - fontes Chakra Petch, Hanken Grotesk e JetBrains Mono, pelo Google Fonts;
  - ícones Lucide;
  - componentes em JSX (Button, IconButton, Icon, Badge, Select, Checkbox, Radio,
    DaySelector, RoleComposition, Drawer, entre outros);
  - assets SVG: fundo "céu pixel", ícones das 5 instâncias e símbolo do RO Lobby.
- **Restrições do próprio design system:** sem logos, artes, sprites ou nomes oficiais
  do Ragnarok. Nomes de instância e classe são inventados (Torre sem fim, Arcebispo).
  O símbolo `c1-symbol` é uma arte original em pixel (balão com três pontos nas cores das
  funções). Textos em pt-BR, com "você", sem exclamação e sem emoji.
- **Estado do código:**
  - `web/` tem só a página de status em `/`, que consome o `/healthz`;
  - não há tabelas de domínio nem API de lobbies, então a Home só pode mostrar dados
    fictícios por enquanto;
  - não há login, então não existe "personagem conectado".
- **Fundação:** a RN-03 diz que o Compose sobe só o PostgreSQL, e as imagens Docker do
  backend e do web ficaram fora de escopo. Rodar tudo em containers é uma mudança nova,
  que precisa de spec e do ADR-06 (hospedagem), hoje pendente no CLAUDE.md.
- **SvelteKit com `adapter-node`:** em produção precisa da variável `ORIGIN` (proteção
  CSRF das actions) e lê a porta de `PORT`.

## Classificação: M (Médio)
Várias camadas (Compose, imagens, web), uma decisão de arquitetura (ADR-06) e nenhuma
regra de domínio nova. A Home é de apresentação, com dados fictícios.
