# Task 015 — Frontend administrativo e dashboard do Agent

status: ativo
owner: frontend + devops

after: 013-configuration-scope-model, 014-scoped-agent-polling

## Objetivo

Criar um frontend acessível pelo navegador para administrar Applications, Environments, Configurations e Flags, além de consultar a saúde do Agent.

## MVP da tela

- listar/criar aplicação;
- listar/criar ambiente da aplicação;
- listar/criar configuração do ambiente;
- listar/criar/editar/remover flags da configuração;
- editar `enabled` e `value` JSON;
- exibir revision, quantidade de flags e última atualização;
- exibir endpoint do Agent e status `READY/DEGRADED`;
- exibir métricas resumidas: revision atual, idade do snapshot, último polling, erros e latência.

## Contratos

O frontend deve usar somente os endpoints `/v1/applications/**` e `/metrics`/`/internal/status`. Não deve acessar PostgreSQL nem depender de CDN. Valores são editados como JSON válido, com mensagem de erro de parsing.

## Dashboard

Adicionar página `/admin` do próprio serviço para facilitar teste local. O dashboard deve atualizar métricas a cada 5 segundos e indicar claramente que os números são locais. Grafana/Prometheus ficam opcionais no Compose; a página não pode depender deles.

## Aceite

- funciona em navegador sem build Node obrigatório;
- fluxo completo cria hierarquia e flag;
- alteração de flag incrementa revision;
- Agent reflete alteração após polling;
- menu de 10 MB pode ser avaliado e renderizado numa tela separada;
- acessibilidade básica, estado de loading/erro e `Content-Security-Policy` sem dependência externa.