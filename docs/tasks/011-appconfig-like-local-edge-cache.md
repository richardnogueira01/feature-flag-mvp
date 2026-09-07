# Task 011 — Camada local inspirada no AWS AppConfig

status: concluída
owner: devops
date: 2026-09-07

## Objetivo

Disponibilizar um caminho local de leitura de flags que reduza chamadas repetidas ao processo Go e ao banco, mantendo a API administrativa sem cache. O consumidor usa `http://localhost:8081`; a API direta continua disponível em `http://localhost:8080`.

## Contexto técnico

O processo Go já mantém o snapshot em memória e publica atualizações via NATS. A nova camada Nginx funciona como cache de borda local:

`cliente -> edge:8081 -> app:8080 -> snapshot em memória`

Não cachear PATCH, POST, DELETE, `/healthz`, Swagger ou rotas administrativas. O TTL de avaliação é 1 segundo para limitar a janela de propagação após alteração de uma flag.

## Implementação

- Criar `infra/nginx/nginx.conf` com upstream persistente para `app:8080`.
- Cachear apenas `GET`/`HEAD` de `/v1/evaluate/*`, incluindo `Accept-Encoding` na chave.
- Ativar `proxy_cache_lock`, atualização em background e resposta stale durante falha transitória do upstream.
- Cachear `204` pelo mesmo TTL curto, para flag desabilitada não gerar resposta grande.
- Expor `X-Edge-Cache` para diagnóstico (`MISS`, `HIT`, `STALE`).
- Expor a porta local `8081` no Compose.

## Critérios de aceite

1. `docker compose up -d --build` inicia `edge` junto com `app`.
2. `GET http://localhost:8081/v1/evaluate/<flag>` mantém o contrato da API direta.
3. A segunda chamada idêntica apresenta `X-Edge-Cache: HIT` após a primeira gerar `MISS`.
4. PATCH/POST continuam chegando diretamente à API e não são armazenados.
5. Após alteração da flag, a nova resposta aparece no máximo após o TTL.
6. Falha transitória do app permite leituras armazenadas durante a janela de stale.
7. Documentar que este cache local não comprova capacidade de produção em escala Netflix.

## Validação manual

```powershell
$base = 'http://localhost:8081'
curl.exe -i "$base/v1/evaluate/menu_itau_mobile"
curl.exe -i "$base/v1/evaluate/menu_itau_mobile"
docker compose ps
docker compose logs edge --tail 50
```

## Limitações

- O cache é por instância do Nginx; produção precisa de distribuição/replicação e invalidação explícita.
- O TTL cria consistência eventual de até 1 segundo.
- Payloads de 10 MB devem preferencialmente ser referenciados por versão/URL e entregues por CDN/object storage, não baixados em cada decisão.