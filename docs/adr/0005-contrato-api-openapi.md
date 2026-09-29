# ADR-05 — Contrato da API: OpenAPI como fonte

- Status: Aceito
- Data: 2026-09-29

## Contexto
O backend em Go (ADR-02) e o web em SvelteKit (ADR-01) conversam por HTTP/JSON. Se os
tipos forem escritos à mão dos dois lados, eles se desencontram sem ninguém perceber.
O projeto é portfólio de QA, então um contrato verificável tem valor por si.

## Opções consideradas
- (a) `openapi.yaml` como fonte. O `oapi-codegen` gera os tipos e a interface do
  servidor em Go (com suporte a `net/http`), e o `openapi-typescript` gera os tipos
  do web.
- (b) Tipos escritos à mão nos dois lados.
- (c) Gerar o OpenAPI a partir de anotações no código Go.

## Decisão
(a) OpenAPI como fonte do contrato, com geração de código para Go e TypeScript.

## Consequências
- \+ Mudar o contrato quebra a compilação do lado que ficou desatualizado.
- \+ O contrato é um documento versionado, que o validador e os testes de contrato
  podem usar.
- \+ A interface gerada do servidor funciona com os handlers da biblioteca padrão.
- − Toda rota nova começa no `openapi.yaml`, o que é um passo a mais.
- − O código gerado precisa ficar sincronizado com o contrato, então a CI confere.
