# Spec — Perfil e personagens

- Feature: `personagens` · Nível: G · Status: Aprovada
- Design: Claude Design, `Perfil.dc.html` no projeto "RO Lobby Home v2", variação 1a
- Última revisão: 2026-10-06 — rodada 1 da entrevista: cards em formato de carteirinha com
  retrato escolhido de uma lista (4 retratos de exemplo) e nick único em todo o RO Lobby

## 1. Contexto
Com o login pronto, o Usuário precisa cadastrar os personagens com que vai criar lobbies e
se candidatar (P-04 da `candidatura-lobby`). Esta feature cria a página de perfil, onde os
personagens aparecem como carteirinhas, e a API que guarda e valida os personagens.

Horários disponíveis e instâncias de interesse ficam para uma feature seguinte. Lobby e
candidatura ainda não existem, então as travas da spec `candidatura-lobby` sobre mudar a
função ou excluir um personagem (RN-25 e RN-26 de lá) entram junto com elas.

## 2. Atores
| Ator | Descrição | Permissões principais |
|---|---|---|
| Usuário | Quem entrou com o Discord | Ver o próprio perfil, cadastrar, editar, excluir e escolher o principal dos próprios personagens |
| Visitante | Quem não entrou | Nada no perfil; é levado ao login |

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como usuário, quero ver meus personagens em carteirinhas no meu perfil, para ter num só lugar com quem eu jogo. |
| US-02 | P1 | Como usuário, quero adicionar um personagem com nick, classe, nível, função, retrato e link, para usá-lo nos grupos. |
| US-03 | P1 | Como usuário, quero editar um personagem, para manter nível, classe e função em dia. |
| US-04 | P1 | Como usuário, quero excluir um personagem, para tirar o que não uso mais. |
| US-05 | P2 | Como usuário, quero marcar um personagem como principal, para ele aparecer primeiro. |
| US-06 | P1 | Como visitante, quero ser levado ao login ao abrir o perfil, para depois voltar a ele. |
| US-07 | P1 | Como desenvolvedor, quero que a API valide os personagens e sirva o catálogo de classes, para as regras valerem em qualquer cliente. |

## 4. Regras

### Dono e acesso
- **RN-01** — Todo personagem pertence a um único Usuário. Listar, criar, editar, excluir e
  marcar como principal exigem sessão e só valem para os personagens do próprio Usuário.
- **RN-02** — Um personagem de outro Usuário é tratado como inexistente: a API responde
  404, sem revelar que ele existe.
- **RN-03** — A página `/perfil` exige login. O visitante é levado ao login e volta para
  `/perfil` depois de entrar (RN-12 da `login-discord`).

### Campos
- **RN-04** — Nick: obrigatório, de 1 a 24 caracteres depois de tirar os espaços das
  pontas, sem caracteres de controle.
- **RN-05** — O nick é único em todo o RO Lobby, sem diferenciar maiúsculas de minúsculas
  ("Brasa" e "brasa" são o mesmo nick). Vale para todos os usuários, para ninguém se passar
  por outra pessoa.
- **RN-06** — Classe: uma das classes do catálogo do bROWiki (RN-22 da `home-local`). A API
  recusa qualquer outra.
- **RN-07** — Nível: número inteiro de 1 a 275.
- **RN-08** — Função: Tank, Suporte ou Dano.
- **RN-09** — Link externo: opcional. Se informado, é uma URL `https://` válida de até 300
  caracteres. Na tela, abre em outra aba e não passa a página de origem ao site de destino.
- **RN-10** — Retrato: um dos retratos da lista do RO Lobby (4 por enquanto). Sem escolha,
  vale o primeiro da lista. Os retratos são arte original, sem nada da Gravity, e são
  servidos pelo próprio web (RNF-02 da `home-local`).

### Limites e principal
- **RN-11** — Cada Usuário tem no máximo 10 personagens.
- **RN-12** — Quando o Usuário tem personagens, exatamente um deles é o principal. O
  primeiro cadastrado vira o principal.
- **RN-13** — O Usuário pode marcar outro personagem como principal a qualquer momento; o
  anterior deixa de ser.
- **RN-14** — Ao excluir o principal, o personagem mais antigo que sobrar vira o principal.
- **RN-15** — Editar pode mudar qualquer campo, respeitando as mesmas validações. Excluir é
  definitivo e pede confirmação na tela.

### Tela
- **RN-16** — Cada personagem aparece numa carteirinha com retrato, nick, classe, nível,
  função (na cor da função), o selo "Principal" quando for o caso e o link externo quando
  houver, além das ações de editar, excluir e marcar como principal.
- **RN-17** — As carteirinhas aparecem com o principal primeiro e os outros por ordem de
  cadastro.
- **RN-18** — O menu do usuário na barra superior ganha "Meu perfil", acima de "Sair".
- **RN-19** — O formulário de personagem mostra os erros de validação da API junto do
  campo, em pt-BR, sem perder o que já foi digitado.

### Catálogo
- **RN-20** — O catálogo de classes passa a viver na API, como fonte única. `GET /classes`
  devolve as classes, não exige sessão, e o formulário do web usa essa lista.

## 5. Critérios de aceite

### US-01 — Ver os personagens
```gherkin
CA-01.1 — Carteirinhas  [US-01, RN-16, RN-17]
Given um Usuário com os personagens "Lirien" (principal) e "Faísca"
When ele abre /perfil
Then vê uma carteirinha para cada personagem, com "Lirien" primeiro
  And cada carteirinha mostra retrato, nick, classe, nível e função
  And a de "Lirien" mostra o selo "Principal"

CA-01.2 — Perfil vazio  [US-01, RN-16]
Given um Usuário sem personagens
When ele abre /perfil
Then vê "Você ainda não tem personagens" e o botão "Adicionar personagem"

CA-01.3 — Só os próprios personagens  [US-01, RN-01]
Given dois Usuários, cada um com um personagem
When um deles abre /perfil
Then vê só o próprio personagem

CA-01.4 — Link externo  [US-01, RN-09, RN-16]
Given um personagem com o link "https://exemplo.com/char/123"
When o Usuário vê a carteirinha
Then o link abre em outra aba, com rel="noopener noreferrer"
```

### US-02 — Adicionar
```gherkin
CA-02.1 — Primeiro personagem vira principal  [US-02, RN-12]
Given um Usuário sem personagens
When ele adiciona "Brasa", Guardião Real, nível 172, Tank
Then o personagem aparece no perfil como principal

CA-02.2 — Retrato padrão e escolhido  [US-02, RN-10]
Given o formulário de novo personagem
When o Usuário salva sem escolher retrato
Then o personagem fica com o primeiro retrato da lista
  And quando ele escolhe outro retrato, esse é o que aparece na carteirinha

CA-02.3 — Nick repetido em outro Usuário  [US-02, RN-05]
Given o nick "Brasa" já cadastrado por outro Usuário
When um Usuário tenta cadastrar "brasa"
Then o cadastro é recusado com a mensagem "Esse nick já está em uso"

CA-02.4 — Classe fora do catálogo  [US-02, RN-06]
Given uma tentativa de cadastro com a classe "Paladino Supremo"
When a API recebe o pedido
Then responde 422 com o erro no campo classe

CA-02.5 — Nível fora da faixa  [US-02, RN-07]
Given uma tentativa de cadastro com nível 0 ou 276
When a API recebe o pedido
Then responde 422 com o erro no campo nível

CA-02.6 — Nick vazio ou longo demais  [US-02, RN-04]
Given uma tentativa de cadastro com nick "   " ou com 25 caracteres
When a API recebe o pedido
Then responde 422 com o erro no campo nick

CA-02.7 — Link inválido  [US-02, RN-09]
Given uma tentativa de cadastro com o link "http://exemplo.com" ou "javascript:alert(1)"
When a API recebe o pedido
Then responde 422 com o erro no campo link

CA-02.8 — Limite de 10  [US-02, RN-11]
Given um Usuário com 10 personagens
When ele tenta adicionar mais um
Then o cadastro é recusado com a mensagem "Você já tem 10 personagens"
  And o botão "Adicionar personagem" aparece desabilitado com essa dica

CA-02.9 — Erro junto do campo  [US-02, RN-19]
Given o formulário com um nick já em uso
When o Usuário salva
Then a mensagem aparece junto do campo nick
  And os outros campos continuam preenchidos
```

### US-03 — Editar
```gherkin
CA-03.1 — Editar campos  [US-03, RN-15]
Given o personagem "Faísca", Feiticeiro, nível 165
When o Usuário muda o nível para 170 e a classe para Elementalista
Then a carteirinha mostra Elementalista e nível 170

CA-03.2 — Manter o próprio nick  [US-03, RN-05]
Given o personagem "Faísca"
When o Usuário edita outro campo e mantém o nick
Then a edição é aceita

CA-03.3 — Personagem de outro Usuário  [US-03, RN-02]
Given um personagem de outro Usuário
When alguém tenta editá-lo pela API
Then a API responde 404
  And o personagem continua igual
```

### US-04 — Excluir
```gherkin
CA-04.1 — Excluir com confirmação  [US-04, RN-15]
Given o personagem "Faísca"
When o Usuário escolhe excluir e confirma
Then o personagem some do perfil
  And o nick "Faísca" fica livre para outro cadastro

CA-04.2 — Cancelar a exclusão  [US-04, RN-15]
Given o personagem "Faísca"
When o Usuário escolhe excluir e cancela
Then o personagem continua no perfil

CA-04.3 — Excluir o principal  [US-04, RN-14]
Given "Lirien" (principal, o segundo cadastrado), "Brasa" (o primeiro) e "Faísca"
When o Usuário exclui "Lirien"
Then "Brasa", o mais antigo que sobrou, vira o principal

CA-04.4 — Personagem de outro Usuário  [US-04, RN-02]
Given um personagem de outro Usuário
When alguém tenta excluí-lo pela API
Then a API responde 404
  And o personagem continua existindo
```

### US-05 — Principal
```gherkin
CA-05.1 — Trocar o principal  [US-05, RN-12, RN-13]
Given "Lirien" é o principal e "Faísca" não
When o Usuário marca "Faísca" como principal
Then "Faísca" passa a ser o único principal e aparece primeiro
```

### US-06 — Acesso
```gherkin
CA-06.1 — Visitante vai para o login  [US-06, RN-03]
Given um visitante
When ele abre /perfil
Then é levado ao login com destino /perfil

CA-06.2 — Menu "Meu perfil"  [US-06, RN-18]
Given um Usuário logado
When ele abre o menu do usuário
Then vê "Meu perfil", que leva a /perfil, acima de "Sair"
```

### US-07 — API e catálogo
```gherkin
CA-07.1 — Catálogo de classes  [US-07, RN-20]
Given a API no ar
When se pede GET /classes sem sessão
Then a resposta tem as 82 classes do catálogo

CA-07.2 — Rotas exigem sessão  [US-07, RN-01]
Given nenhuma sessão
When se pede para listar, criar, editar ou excluir personagens
Then a API responde 401

CA-07.3 — Nick único no banco  [US-07, RN-05]
Given dois pedidos simultâneos com o nick "Brasa"
When os dois chegam à API
Then só um personagem é criado
```

## 6. Casos de borda
- Nick com espaços nas pontas: " Brasa " é salvo como "Brasa" [RN-04]
- Nick com acento: "Faísca" e "Faisca" são nicks diferentes [RN-05]
- Excluir o único personagem: o Usuário fica sem principal até cadastrar outro [RN-12]
- Classe que sair do catálogo no futuro: os personagens já salvos continuam, e a edição
  pede uma classe válida [RN-06]
- Sessão vencida no meio da edição: a API responde 401 e o web leva ao login [RN-01, RN-03]

## 7. Requisitos não funcionais
- **RNF-01** — Acessibilidade: carteirinhas, formulário e diálogo de confirmação funcionam
  pelo teclado, com foco visível e rótulos em todos os campos.
- **RNF-02** — Nenhum recurso carregado de fora do servidor do web (RNF-02 da `home-local`).
- **RNF-03** — Todo teste cita o ID do critério de aceite (CLAUDE.md).
- **RNF-04** — As regras dos campos valem na API; o web repete as mais simples só para
  avisar antes de enviar.

## 8. Fora de escopo
- Horários disponíveis (Disponibilidade) e instâncias de interesse.
- Upload de imagem e a lista definitiva de retratos.
- Confirmar no jogo que o personagem é do Usuário, e contestar um nick já cadastrado.
- As travas da `candidatura-lobby` sobre mudar a função ou excluir personagem com
  candidatura (entram com lobby e candidatura).
- Perfil público de outros jogadores.
- Campo de servidor (o foco é só o Ragnarok LATAM).

## 9. Perguntas em aberto
- Nenhuma.

## 10. Decisões tomadas na entrevista
- Escopo → só personagem; horários e instâncias de interesse numa feature seguinte.
- Design → tela de perfil desenhada no Claude Design antes de implementar.
- Onde fica → `/perfil`, pelo item "Meu perfil" no menu do usuário.
- Listagem → cards em formato de carteirinha, com retrato escolhido de uma lista (4 de
  exemplo agora; a lista definitiva vem depois).
- Campos → nick 1–24, classe do catálogo, nível 1–275, função, link `https://` opcional,
  sem servidor.
- Nick → único em todo o RO Lobby, sem diferenciar maiúsculas, para evitar que alguém se
  passe por outra pessoa.
- Limites → até 10 personagens; primeiro vira principal; excluir o principal passa para o
  mais antigo.
- Editar e excluir → tudo editável; excluir com confirmação; travas de candidatura depois.
- Catálogo de classes → na API (`GET /classes`), como fonte única.
