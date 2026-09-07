# Task: 004-nats-sync-readiness

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: in-progress
- Branch: task/feature-flag-mvp-mvp
- Dependências: 003-postgres-outbox

## Contexto

Instâncias do Data Plane recebem revisões via NATS/JetStream. A próxima revisão esperada é local + 1;
revisão maior indica gap e exige full resync. Snapshot válido continua servindo durante falhas externas.

## Implementado

- Dependência nats.go.
- messaging.Publisher compatível com persistence.Publisher.
- messaging.Subscriber com durable consumer, manual ack, Term em payload inválido e Nak em erro.
- syncer.Syncer com Apply, Resync, Fetcher e Ready.
- Revisões antigas são ignoradas; gaps iniciam resync.
- endpoints /healthz, /readyz e /internal/status.
- main registra os endpoints e usa Syncer como fonte de readiness.
- Testes unitários do sincronizador e status.

## Ainda necessário

- Conectar subscriber real ao ciclo de vida do main.
- Integrar payload da outbox como snapshot completo distribuível.
- Configurar stream/consumer NATS por variáveis sem secrets.
- Fornecer Fetcher real para full resync.
- Criar testes de integração com NATS/JetStream.

## Aceite

- [ ] Subscriber real recebe, aplica e confirma eventos.
- [ ] Gap dispara resync e não perde consistência.
- [ ] Nova instância não fica ready antes do sync.
- [ ] Testes NATS reais passam.
- [ ] Marcar done e mover para docs/history/.

## Evidências atuais

- gofmt -w cmd internal: OK
- go test ./...: OK
- go vet ./...: OK
- go build ./...: OK
- git diff --check: OK
