# Task: 015-admin-frontend-observability

- Status: done (com ressalvas documentadas abaixo)
- Concluída em: 2026-09-13
- Branch: task/015-admin-frontend-observability

## Resultado

Página `/admin`, servida pelo próprio binário `feature-flag-mvp` (`internal/httpapi/admin.go`),
sem build Node e sem CDN — um único documento HTML com CSS e JS inline, `Content-Security-Policy:
default-src 'self'` (só relaxa `connect-src` para permitir consultar a URL do Agent, que roda em
outra origem/porta).

Cobre o MVP da spec:

- listar/criar application, environment, configuration;
- listar/criar/editar (`enabled` e `value` JSON, com validação e mensagem de erro de parse)
  /remover flags de uma configuration, usando os endpoints da task 013;
- exibe revision, quantidade de flags e última atualização da configuration selecionada
  (precisou adicionar `updated_at` ao `ConfigurationSnapshot`, que ainda não existia);
- dashboard "Control Plane" (status, revision, gaps, resyncs, outbox, avaliações e latência
  média, lidos de `/internal/status` e `/metrics`) e painel "Agent" (URL configurável, persistida
  em `localStorage`, status `READY`/`DEGRADED` via `/readyz`), atualizados a cada 5s;
- link para o teste de payload de 10 MiB já existente (`/test/large-payload`), como pedido pela
  spec ("numa tela separada" — não foi reimplementado, só linkado);
- acessibilidade básica: labels associados aos inputs, região `aria-live` para mensagens,
  controles nativos (botão/checkbox), sem widgets custom.

O endpoint `/readyz` do Agent (`cmd/feature-flag-agent`) ganhou corpo JSON
(`{status,ready,revision}`) e `Access-Control-Allow-Origin: *` — sem isso o dashboard não
conseguiria ler o status do Agent a partir de outra origem/porta (não há dado sensível nesses
dois endpoints).

## Limitações conhecidas, por causa das tasks deferidas (014/017/018/019)

- "Agent reflete alteração após polling" só é observável para a configuração `legacy/default`
  hoje — o Agent atual sincroniza `/v1/flags` (API plana), não é escopado por
  application/environment/configuration; isso é exatamente o que a task 014 (deferida) resolve.
- "Métricas... último polling, erros e latência" do Agent não existem ainda — o binário do Agent
  não expõe `/metrics` nem contadores de polling (isso também é escopo da 014/019, deferidas). O
  painel mostra isso explicitamente em vez de fabricar números.

## Evidências

- `gofmt -l .`, `go vet ./...`, `go build ./...`: limpos.
- `go test ./...` e `go test -race ./...`: OK, incluindo um teste novo (`admin_test.go`) que
  confirma a CSP estrita e a ausência de qualquer referência a CDN na página.
- Testes de integração contra PostgreSQL real (`TEST_DATABASE_URL`): ainda passam com o campo
  `updated_at` novo.
- **Fluxo completo validado de ponta a ponta via HTTP real**: subi o binário `feature-flag-mvp`
  contra um PostgreSQL real e o binário `feature-flag-agent` de verdade, e reproduzi via `curl`
  exatamente as chamadas que o JS da página faz — criar application/environment/configuration,
  criar flag com `enabled` e com `value` objeto, `PATCH` enabled, `PUT` value, listar snapshot
  (confirmando `revision`/`updated_at`/`flags` corretos a cada mutação), `DELETE`, e os casos de
  erro (nome inválido → 400, duplicado → 409, escopo inexistente → 404) — todos com o formato de
  resposta que o `api()` do JS espera. Também confirmei `/readyz` do Agent real com o CORS header
  e o JSON que o painel lê.
- **Não validado visualmente num navegador**: a ferramenta de browser automatizado disponível
  nesta sessão falhou por um bug de ambiente (tentou criar um diretório de perfil com um caminho
  estilo Windows, `D:`, dentro de um ambiente Linux) — reportado como limitação, não contornado.
  A verificação acima cobre toda a lógica de dados (contratos HTTP, parsing, renderização
  determinística a partir deles), mas não cobre CSS/layout/render real.

## Pendências para o dono

- Abrir `/admin` num navegador de verdade para conferir layout/responsividade — a lógica foi
  validada, o visual não.
- Ao implementar a task 014 no futuro, revisitar o texto do painel "Agent" nesta página (hoje
  avisa explicitamente que polling/erros/latência por Agent ainda não existem).
