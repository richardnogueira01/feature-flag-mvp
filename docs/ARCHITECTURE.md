# Feature Flag Service — Architecture

## 1. Overview

Este projeto implementa uma plataforma de **Feature Flags de alta performance**, desenvolvida em **Go**, projetada para suportar grandes volumes de avaliações com latência extremamente baixa.

Os principais objetivos arquiteturais são:

- Suportar milhões de avaliações de Feature Flags.
- Manter latência mínima no Data Plane.
- Evitar dependências externas no caminho crítico de leitura.
- Permitir execução em qualquer cloud ou ambiente on-premises.
- Manter as instâncias sincronizadas praticamente em tempo real.
- Continuar avaliando flags mesmo durante indisponibilidade do banco ou sistema de mensageria.
- Permitir evolução futura para SDKs com avaliação local.
- Escalar horizontalmente.
- Evitar vendor lock-in.

A arquitetura separa explicitamente:

```text
Control Plane
    ↓
Gerenciamento das Feature Flags

Data Plane
    ↓
Avaliação das Feature Flags
```

---

# 2. Architectural Principles

## 2.1 Cloud Agnostic

O core da aplicação não deve possuir dependências diretas de serviços específicos de cloud providers.

Não devem existir dependências obrigatórias como:

```text
AWS DynamoDB
AWS SNS/SQS
AWS AppConfig

Azure Service Bus
Azure Cosmos DB

Google Pub/Sub
Google Cloud Spanner
```

A implementação padrão deve utilizar tecnologias portáveis:

```text
Go
PostgreSQL
NATS
Docker / OCI
Kubernetes
OpenTelemetry
Prometheus
```

O sistema deve poder executar sem alterações funcionais em:

```text
AWS
Azure
Google Cloud
OpenShift
Kubernetes
Rancher
k3s
Bare Metal
VMs
Docker
On-Premises
```

Integrações específicas de cloud poderão existir futuramente como **adapters opcionais**.

---

# 3. High-Level Architecture

```text
                         ┌──────────────────────┐
                         │      Admin / UI      │
                         └──────────┬───────────┘
                                    │
                                    ▼
                         ┌──────────────────────┐
                         │    Control Plane     │
                         │        Go API        │
                         │                      │
                         │ POST /flags          │
                         │ PUT /flags/{key}     │
                         │ DELETE /flags/{key}  │
                         └──────────┬───────────┘
                                    │
                                    ▼
                              PostgreSQL
                         ┌──────────────────────┐
                         │ feature_flags        │
                         │ flag_history         │
                         │ outbox_events        │
                         └──────────┬───────────┘
                                    │
                                    ▼
                           Outbox Publisher
                                    │
                                    ▼
                              NATS / JetStream
                                    │
                 ┌──────────────────┼──────────────────┐
                 │                  │                  │
                 ▼                  ▼                  ▼
          ┌────────────┐     ┌────────────┐     ┌────────────┐
          │ Data Plane │     │ Data Plane │     │ Data Plane │
          │   Task A   │     │   Task B   │     │   Task C   │
          │            │     │            │     │            │
          │ RAM Cache  │     │ RAM Cache  │     │ RAM Cache  │
          │ version 42 │     │ version 42 │     │ version 42 │
          └──────┬─────┘     └──────┬─────┘     └──────┬─────┘
                 │                  │                  │
                 └──────────────────┼──────────────────┘
                                    │
                              Flag Evaluation
```

---

# 4. Control Plane

O **Control Plane** é responsável pela administração das Feature Flags.

Responsabilidades:

- criação;
- edição;
- exclusão;
- ativação/desativação;
- versionamento;
- auditoria;
- histórico;
- autenticação;
- autorização;
- publicação de alterações.

Exemplos iniciais:

```http
POST   /v1/flags
PUT    /v1/flags/{key}
DELETE /v1/flags/{key}
GET    /v1/flags
GET    /v1/flags/{key}
```

O Control Plane **não participa do caminho crítico de avaliação**.

Sua latência é menos crítica que a do Data Plane.

---

# 5. Data Plane

O **Data Plane** é responsável exclusivamente pela avaliação extremamente rápida das Feature Flags.

Exemplo:

```http
GET /v1/evaluate/{key}
```

Resposta mínima:

```json
{
  "enabled": true
}
```

O fluxo deve ser:

```text
Request
   │
   ▼
HTTP Server
   │
   ▼
Router
   │
   ▼
Local Memory
   │
   ▼
Response
```

O fluxo NÃO deve ser:

```text
Request
   ↓
Redis
   ↓
PostgreSQL
   ↓
Response
```

Nem:

```text
Request
   ↓
PostgreSQL
   ↓
Response
```

### Regra arquitetural

> Nenhuma dependência externa deve existir no hot path normal de avaliação de uma Feature Flag.

Isso inclui:

- PostgreSQL;
- NATS;
- Redis/Valkey;
- APIs externas;
- serviços específicos de cloud.

---

# 6. Source of Truth

A implementação padrão utilizará:

```text
PostgreSQL
```

como **Source of Truth**.

Motivos:

- open source;
- amplamente suportado;
- cloud agnostic;
- suporte a transações ACID;
- excelente ecossistema;
- disponível como serviço gerenciado em praticamente qualquer cloud;
- fácil execução on-premises;
- suporte nativo a JSONB;
- maturidade operacional.

O PostgreSQL não será dimensionado para suportar milhões de avaliações por segundo.

Ele será utilizado principalmente para:

```text
CRUD
auditoria
histórico
versionamento
configuração
outbox
```

---

# 7. Feature Flag Model

Modelo inicial:

```text
FeatureFlag

id
namespace
application
environment
key
enabled
value
version
created_at
updated_at
```

Exemplo:

```json
{
  "id": "019...",
  "namespace": "bank",
  "application": "payments",
  "environment": "production",
  "key": "new-checkout",
  "enabled": true,
  "value": true,
  "version": 42
}
```

Uma flag deve ser identificável de forma única através de:

```text
namespace
+
application
+
environment
+
key
```

Exemplo lógico:

```text
bank/production/payments/new-checkout
```

---

# 8. Local In-Memory Cache

Cada instância do Data Plane mantém seu próprio snapshot das Feature Flags em memória.

Exemplo conceitual:

```go
type Snapshot struct {
    Version uint64
    Flags   map[string]Flag
}

var current atomic.Pointer[Snapshot]
```

Uma avaliação executa aproximadamente:

```go
snapshot := current.Load()

flag, exists := snapshot.Flags[key]
```

Nenhum lock deve ser necessário no caminho normal de leitura.

---

# 9. Immutable Snapshots

O mapa atual não deve ser alterado enquanto estiver sendo utilizado pelas requisições.

Quando uma configuração for atualizada:

```text
Snapshot v41
      │
      │ clone + alteração
      ▼
Snapshot v42
      │
      │ atomic swap
      ▼
current → v42
```

Requisições concorrentes enxergarão:

```text
v41 completa
```

ou:

```text
v42 completa
```

Nunca um estado parcialmente atualizado.

Isso permite leitura altamente concorrente com praticamente nenhuma contenção.

---

# 10. Versioning

Toda alteração de configuração deve produzir uma versão monotonicamente crescente.

Exemplo:

```text
41
42
43
44
```

Eventos devem carregar essa versão:

```json
{
  "type": "FEATURE_FLAGS_UPDATED",
  "version": 44,
  "namespace": "bank",
  "application": "payments",
  "environment": "production"
}
```

Uma instância mantém:

```text
currentVersion = 43
```

Ao receber:

```text
version = 44
```

ela aplica a atualização.

Se receber novamente:

```text
version = 43
```

o evento é ignorado.

---

# 11. Detecting Lost Events

Se uma instância possuir:

```text
version = 41
```

e receber:

```text
version = 43
```

ela detectará que perdeu uma atualização.

Nesse caso:

```text
event v43
    │
    ▼
expected v42
    │
    ▼
version mismatch
    │
    ▼
FULL RESYNC
```

A instância deve buscar o snapshot atual e substituir atomicamente seu estado local.

---

# 12. Real-Time Distribution

Alterações devem ser distribuídas utilizando **push**, e não polling como mecanismo primário.

Implementação padrão:

```text
NATS
+
JetStream
```

NATS foi escolhido inicialmente por:

- baixa latência;
- footprint reduzido;
- facilidade operacional;
- suporte a pub/sub;
- suporte a persistência através do JetStream;
- funcionamento em Kubernetes;
- funcionamento em Docker;
- funcionamento on-premises;
- independência de cloud provider.

Kafka poderá ser implementado futuramente através de adapter.

---

# 13. Broadcast Semantics

Alterações de Feature Flags precisam chegar a **todas as instâncias relevantes**.

Não deve ser utilizado um modelo no qual várias instâncias competem pela mesma mensagem e somente uma recebe a atualização.

O comportamento desejado é:

```text
                 FLAG UPDATED
                      │
            ┌─────────┼─────────┐
            ▼         ▼         ▼
          Task A    Task B    Task C
```

Todas devem convergir para a mesma versão.

---

# 14. Consistency Model

O sistema utilizará:

```text
Strong local consistency
+
Near-real-time distributed convergence
```

Dentro de uma instância, uma request sempre utiliza um snapshot consistente.

Entre instâncias, poderá existir uma janela extremamente pequena durante a propagação.

Exemplo:

```text
Task A → v42
Task B → v41
Task C → v41

       ↓ alguns ms

Task A → v42
Task B → v42
Task C → v42
```

Consistência absoluta instantânea entre máquinas distribuídas não será requisito padrão, pois exigiria coordenação no caminho crítico e prejudicaria disponibilidade e latência.

---

# 15. Scheduled Atomic Activation

Para casos excepcionais que exijam mudança aproximadamente simultânea entre todas as instâncias, poderá existir:

```text
activateAt
```

Exemplo:

```json
{
  "version": 45,
  "activateAt": "2026-09-07T20:00:05Z"
}
```

O snapshot poderá ser distribuído antecipadamente.

```text
20:00:04

Task A → recebeu v45
Task B → recebeu v45
Task C → recebeu v45

           ↓

20:00:05

Task A → ativa v45
Task B → ativa v45
Task C → ativa v45
```

Essa funcionalidade depende de sincronização adequada dos relógios dos hosts.

Não será necessária para o fluxo padrão.

---

# 16. Transactional Outbox

Uma alteração não deve seguir simplesmente:

```text
UPDATE PostgreSQL

        ↓

publish NATS
```

Existe uma condição de falha:

```text
PostgreSQL COMMIT = SUCCESS

NATS publish = FAILURE
```

Nesse cenário o banco teria uma configuração que as instâncias nunca receberiam.

Será utilizado o padrão:

**Transactional Outbox**.

Dentro da mesma transação:

```text
BEGIN

UPDATE feature_flags

INSERT INTO outbox_events

COMMIT
```

Depois:

```text
Outbox Publisher
      │
      ▼
NATS
```

Após confirmação da publicação:

```text
outbox event → processed
```

Isso garante que alterações persistidas possam ser eventualmente distribuídas mesmo diante de falhas temporárias do broker.

---

# 17. Instance Startup

Uma instância nunca deve começar atendendo requests antes de possuir um snapshot válido.

Startup:

```text
STARTING
    │
    ▼
Load persisted/local snapshot
    │
    ▼
Connect distribution channel
    │
    ▼
Synchronize latest version
    │
    ▼
Build memory snapshot
    │
    ▼
READY
```

Somente após a sincronização:

```text
Readiness = TRUE
```

---

# 18. Readiness

Endpoint:

```http
GET /ready
```

Exemplo:

```json
{
  "ready": true,
  "configVersion": 1452
}
```

Instâncias sem configuração válida devem retornar:

```http
503 Service Unavailable
```

e não devem receber tráfego.

---

# 19. Health and Synchronization

Cada instância deverá expor informações sobre seu estado.

Exemplo:

```http
GET /internal/status
```

Resposta:

```json
{
  "status": "UP",
  "configVersion": 1452,
  "configHash": "sha256:...",
  "lastUpdate": "2026-09-07T19:42:10Z"
}
```

Isso permite detectar:

```text
Task A → v1452 / ABC
Task B → v1452 / ABC
Task C → v1451 / XYZ
```

e identificar imediatamente uma instância fora de sincronia.

---

# 20. Snapshot Hash

Além da versão, snapshots deverão possuir um hash determinístico.

Exemplo:

```text
SHA-256
```

Dessa maneira:

```text
version = 1452
hash    = ABC123
```

representa exatamente o mesmo conteúdo em todas as instâncias.

O hash ajuda na detecção de corrupção ou divergências inesperadas.

---

# 21. Failure Strategy

Uma das principais propriedades do sistema será:

> A avaliação de Feature Flags deve continuar funcionando durante falhas do Control Plane.

Exemplo:

```text
PostgreSQL DOWN
NATS DOWN
Control Plane DOWN

             ↓

Data Plane

Task A → RAM Snapshot
Task B → RAM Snapshot
Task C → RAM Snapshot

             ↓

Evaluation continues
```

Nesse cenário, as instâncias utilizam o último estado conhecido.

O sistema prioriza:

```text
Availability of evaluation
```

sobre:

```text
ability to modify flags
```

---

# 22. Local Persistent Snapshot

Opcionalmente, cada Data Plane poderá persistir o último snapshot válido localmente.

Exemplo:

```text
/var/lib/featureflags/snapshot.json
```

Isso permite recuperação mesmo quando uma instância reinicia durante indisponibilidade do Control Plane.

Fluxo:

```text
Process starts
      │
      ▼
Load local snapshot
      │
      ▼
Populate RAM
      │
      ▼
Try remote synchronization
```

A política exata sobre quando uma instância poderá ficar `READY` usando apenas um snapshot persistido deverá ser configurável.

---

# 23. Redis / Valkey

Redis/Valkey **não fará parte inicialmente do hot path**.

Não utilizar:

```text
GET
 ↓
Redis
 ↓
Response
```

como mecanismo padrão.

A arquitetura inicial será:

```text
GET
 ↓
Local RAM
 ↓
Response
```

Valkey poderá futuramente ser implementado como L2 cache opcional caso exista uma necessidade concreta.

---

# 24. Performance Strategy

O Data Plane deve minimizar:

- allocations;
- locks;
- serialização;
- payload;
- chamadas de rede;
- syscalls desnecessárias;
- dependências externas.

Objetivo arquitetural:

```text
HTTP
 ↓
Routing
 ↓
Memory Lookup
 ↓
Minimal Serialization
 ↓
HTTP Response
```

O principal custo esperado em cargas extremamente altas será:

```text
HTTP
TCP
TLS
network
serialization
kernel
load balancing
```

e não o lookup da Feature Flag.

---

# 25. Initial Resource Profile

Configuração inicial recomendada para Data Plane:

```yaml
resources:
  requests:
    cpu: "1"
    memory: "512Mi"

  limits:
    cpu: "4"
    memory: "1Gi"
```

A configuração definitiva deverá ser determinada através de benchmarks.

O sistema deverá possuir testes para medir pelo menos:

```text
requests/sec
p50 latency
p95 latency
p99 latency
p99.9 latency
CPU
RAM
allocations/op
bytes/op
error rate
```

Não devem existir garantias de RPS baseadas somente em estimativas teóricas.

---

# 26. Horizontal Scaling

O Data Plane será stateless do ponto de vista de persistência central.

Isso permite:

```text
1 instance
   ↓
10 instances
   ↓
100 instances
```

sem alterar o modelo arquitetural.

Cada instância possui somente uma cópia local do snapshot.

---

# 27. Deployment

Os componentes oficiais deverão ser distribuídos como imagens OCI.

Exemplo:

```text
featureflags/control-plane
featureflags/data-plane
```

O projeto deverá fornecer inicialmente:

```text
Dockerfile
docker-compose.yml
```

e posteriormente:

```text
Helm Chart
```

Permitindo execução em:

```text
Docker Compose
Kubernetes
EKS
AKS
GKE
OpenShift
Rancher
k3s
Bare Metal Kubernetes
```

---

# 28. Observability

A implementação deve ser vendor-neutral.

Padrão:

```text
OpenTelemetry
```

Métricas:

```text
Prometheus compatible
```

Métricas importantes:

```text
flag_evaluations_total

flag_evaluation_duration

snapshot_version

snapshot_age_seconds

snapshot_update_total

snapshot_update_failures_total

event_processing_total

event_processing_failures_total

outbox_pending_events

outbox_publish_failures_total
```

Tracing não deve adicionar overhead significativo ao hot path e deverá possuir sampling configurável.

---

# 29. Code Architecture

O projeto deverá seguir separação clara entre domínio e infraestrutura.

Estrutura inicial:

```text
cmd/
  control-plane/
  data-plane/

internal/

  domain/
    flag.go
    snapshot.go

  application/
    create_flag.go
    update_flag.go
    delete_flag.go
    evaluate_flag.go

  ports/
    flag_repository.go
    event_publisher.go
    snapshot_store.go

  adapters/

    postgres/

    memory/

    nats/

    kafka/

    valkey/

  transport/

    http/
```

O domínio não deverá importar adapters.

---

# 30. Interfaces

Exemplo:

```go
type FlagRepository interface {
    Get(ctx context.Context, key FlagKey) (*Flag, error)
    Save(ctx context.Context, flag *Flag) error
    Delete(ctx context.Context, key FlagKey) error
}
```

Distribuição:

```go
type EventPublisher interface {
    Publish(ctx context.Context, event FlagChanged) error
}
```

Snapshot:

```go
type SnapshotStore interface {
    Current() *Snapshot
    Replace(snapshot *Snapshot)
}
```

Isso permitirá adapters futuros como:

```text
PostgreSQL
CockroachDB
DynamoDB

NATS
Kafka
SNS
Google Pub/Sub
Azure Service Bus

Memory
Valkey
```

sem acoplamento do domínio.

---

# 31. SDK Architecture — Future

A evolução natural da plataforma será disponibilizar SDKs.

Inicialmente:

```text
Go SDK
Java SDK
```

Posteriormente:

```text
.NET
Python
Node.js
```

Em vez de:

```text
Application
    │
    │ HTTP
    ▼
Feature Flag Data Plane
```

teremos:

```text
Application
┌─────────────────────────┐
│                         │
│ Feature Flag SDK        │
│                         │
│ Local Snapshot          │
│                         │
│ flags.Enabled("x")      │
│                         │
└─────────────────────────┘
```

A avaliação passa a ser local.

Exemplo:

```go
if flags.Enabled("new-checkout") {
    // new implementation
}
```

Isso elimina do caminho crítico:

```text
HTTP
DNS
load balancer
network
TLS
Data Plane
```

permitindo que milhões ou até dezenas de milhões de avaliações ocorram diretamente dentro das aplicações consumidoras sem gerar carga proporcional na plataforma.

---

# 32. Data Plane Role After SDK Adoption

Com SDKs, o Data Plane evoluirá de:

```text
Evaluation Server
```

para principalmente:

```text
Configuration Distribution Service
```

Sua responsabilidade será distribuir:

```text
snapshots
updates
versions
```

para os SDKs.

Isso permite que:

```text
100.000.000 evaluations/s
```

nas aplicações não signifiquem:

```text
100.000.000 requests/s
```

na plataforma.

---

# 33. Security — Future Decision

O desenho deverá permitir posteriormente:

```text
TLS
mTLS
OAuth2/OIDC
API Keys
RBAC
namespace isolation
audit logs
secret rotation
```

Nenhum mecanismo específico de identidade de cloud deverá ser obrigatório.

Integrações com IAM de provedores poderão existir como adapters.

---

# 34. Initial Technology Decisions

| Área | Decisão inicial |
|---|---|
| Language | Go |
| HTTP | `net/http` ou router minimalista |
| Source of Truth | PostgreSQL |
| Event Distribution | NATS + JetStream |
| Hot Cache | Local RAM |
| Cache Strategy | Immutable snapshots |
| Atomicity | Atomic pointer swap |
| L2 Cache | Não inicialmente |
| Redis/Valkey | Opcional |
| Event Reliability | Transactional Outbox |
| Versioning | Monotonic revision |
| Integrity | Snapshot hash |
| Observability | OpenTelemetry |
| Metrics | Prometheus |
| Packaging | OCI/Docker |
| Local deployment | Docker Compose |
| Orchestration | Kubernetes |
| Kubernetes packaging | Helm |
| Cloud dependencies | Nenhuma obrigatória |

---

# 35. Architecture Invariants

As seguintes regras devem ser consideradas invariantes arquiteturais.

### INV-01 — No External I/O During Evaluation

A avaliação normal de uma flag nunca depende de banco, cache distribuído ou broker.

```text
evaluate → memory
```

### INV-02 — Source of Truth Is Not the Serving Layer

PostgreSQL armazena estado persistente, mas não atende avaliações em massa.

### INV-03 — Snapshots Are Immutable

Um snapshot publicado para readers nunca é modificado.

### INV-04 — Updates Are Atomic

Uma instância muda de uma versão completa para outra versão completa.

### INV-05 — Every Configuration Has a Version

Não existe alteração de configuração sem nova revisão.

### INV-06 — Missed Updates Must Be Detectable

Uma instância deve conseguir identificar gaps de versão e executar resync.

### INV-07 — Evaluation Survives Control Plane Failure

Falhas de PostgreSQL, NATS ou Control Plane não devem interromper avaliações que já possuem snapshot válido.

### INV-08 — Infrastructure Is Replaceable

Dependências externas são implementadas através de ports/adapters.

### INV-09 — Cloud Provider Is Not Part of the Domain

O domínio não conhece AWS, Azure, GCP ou qualquer outro provedor.

### INV-10 — New Instances Must Synchronize Before Receiving Traffic

Uma instância sem configuração válida não deve entrar no balanceamento.

---

# 36. Initial Non-Goals

A primeira versão não precisa implementar:

- targeting complexo por usuário;
- experimentação A/B;
- analytics de conversão;
- machine learning;
- multi-region active-active;
- SDK para todas as linguagens;
- Redis obrigatório;
- integração específica com cloud providers;
- UI administrativa sofisticada.

O foco inicial será:

```text
Correctness
      ↓
Consistency
      ↓
Reliability
      ↓
Low Latency
      ↓
Horizontal Scalability
      ↓
Portability
```

---

# 37. Initial MVP

O primeiro MVP deverá demonstrar:

```text
1. Criar Feature Flag

2. Atualizar Feature Flag

3. Excluir Feature Flag

4. Persistir no PostgreSQL

5. Registrar evento através de Transactional Outbox

6. Publicar alteração no NATS

7. Distribuir alteração para todas as instâncias

8. Atualizar snapshot local atomicamente

9. Avaliar Feature Flag exclusivamente em RAM

10. Detectar divergência de versão

11. Realizar full resync

12. Continuar avaliando durante indisponibilidade do Control Plane

13. Expor métricas

14. Executar benchmark de carga
```

---

# 38. Target Architecture

A visão de longo prazo é:

```text
                         ┌────────────────────┐
                         │       UI/API       │
                         └─────────┬──────────┘
                                   │
                                   ▼
                         ┌────────────────────┐
                         │   Control Plane    │
                         └─────────┬──────────┘
                                   │
                                   ▼
                             PostgreSQL
                                   │
                            Transactional
                                Outbox
                                   │
                                   ▼
                              NATS/Kafka
                                   │
                                   ▼
                       Configuration Distribution
                                   │
                 ┌─────────────────┼─────────────────┐
                 │                 │                 │
                 ▼                 ▼                 ▼
            Java SDK           Go SDK           .NET SDK
                 │                 │                 │
            Local RAM          Local RAM         Local RAM
                 │                 │                 │
                 ▼                 ▼                 ▼
             evaluate()        evaluate()        evaluate()
                 │                 │                 │
                 └──────── NO NETWORK ──────────────┘
```

A plataforma central distribui **estado**.

As aplicações executam **avaliações localmente**.

Esse modelo permite escalar a quantidade de avaliações sem aumentar proporcionalmente a infraestrutura central.

---

# 39. Core Architectural Goal

O objetivo final da plataforma é permitir que uma Feature Flag seja avaliada da forma mais próxima possível de:

```go
flag := snapshot.Flags[key]
```

mantendo ao mesmo tempo:

```text
durabilidade
versionamento
auditoria
distribuição em tempo real
tolerância a falhas
consistência
observabilidade
portabilidade
```

A complexidade do sistema deve existir **fora do caminho crítico de leitura**.

> **Reads should be boring. Distribution can be complex.**