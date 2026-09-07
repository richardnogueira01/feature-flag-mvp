# Task: 004-nats-sync-readiness

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: in-progress
- Branch: task/feature-flag-mvp-mvp
- Dependências: 003-postgres-outbox

## Contexto

Instâncias do Data Plane recebem revisões via NATS/JetStream. A próxima revisão esperada é local + 1;
revisão maior indica gap e exige full resync pela fonte de verdade. Snapshot válido continua servindo
durante falhas externas.

## Implementado

- Dependência nats.go.
- messaging.Publisher compatível com persistence.Publisher.
- syncer.Syncer com Apply, Resync e Fetcher injetável.
- Revisões antigas são ignoradas.
- Gaps iniciam full resync.
- Ready só fica true após snapshot válido.
- Testes unitários de sequência, revisão antiga e gap.

## Ainda necessário

- Criar subscriber JetStream com durable consumer e ack explícito.
- Integrar o evento da outbox ao formato de snapshot distribuído.
- Implementar endpoint /internal/status e endpoint /readyz.
- Garantir que resync concorrente faça apenas uma requisição.
- Criar testes de integração com NATS/JetStream.

## Aceite

- [ ] Subscriber real recebe, aplica e confirma eventos.
- [ ] Evento atrasado não altera snapshot.
- [ ] Gap dispara resync e não perde consistência.
- [ ] Nova instância não fica ready antes do sync.
- [ ] Testes NATS reais passam.
- [ ] Marcar done e mover para docs/history/.

## Evidências atuais

- go get github.com/nats-io/nats.go@v1.39.1: OK
- gofmt -w internal: OK
- go test ./...: OK
- go vet ./...: OK
- go build ./...: OK
- git diff --check: OK
