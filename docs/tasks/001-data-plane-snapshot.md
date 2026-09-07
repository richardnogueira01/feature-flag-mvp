# Task: 001-data-plane-snapshot

- Projeto: feature-flag-mvp
- Tipo: feature
- Status: planned
- Branch: task/feature-flag-mvp-data-plane-snapshot
- Dependências: nenhuma

## Contexto

O repositório é um serviço Go inicialmente vazio. O Data Plane avalia flags sem PostgreSQL, NATS,
Redis ou rede. Um snapshot é um conjunto completo de flags com revisão monotônica; leitores nunca
podem observar estado parcial.

## Objetivo

Criar o núcleo de avaliação em memória usando cópia defensiva e troca atômica.

## Implementação

- Criar go.mod com módulo github.com/richardnogueira01/feature-flag-mvp.
- Criar internal/snapshot.
- Definir Flag com Key string e Enabled bool; Snapshot com Revision uint64 e Flags map.
- Expor NewStore, Publish, Evaluate e Revision.
- Publish ignora nil, copia o mapa e publica o snapshot completo atomicamente.
- Evaluate retorna enabled, found e revision sem I/O.

## Testes e aceite

- [ ] Testar flag existente, inexistente e store sem snapshot.
- [ ] Testar que mutar o mapa original não altera o snapshot.
- [ ] Testar leitura/publicação concorrentes com race detector.
- [ ] Passar gofmt, go test ./..., go test -race ./... e go vet ./....
- [ ] Marcar done, registrar resultados e mover esta spec para docs/history/.
