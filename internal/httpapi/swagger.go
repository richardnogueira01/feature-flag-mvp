package httpapi

import "net/http"

const openAPISpec = `openapi: 3.0.3
info:
  title: Feature Flag MVP API
  version: 0.1.0
servers:
  - url: http://localhost:8080
paths:
  /v1/flags:
    get:
      summary: Lista todas as flags
      responses:
        '200': {description: Lista de flags}
    post:
      summary: Cria uma flag
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateFlagRequest'
            example: {key: checkout, enabled: true}
      responses:
        '201': {description: Flag criada}
        '400': {description: JSON ou key inválida}
        '409': {description: Flag já existe}
  /v1/flags/{key}:
    parameters:
      - name: key
        in: path
        required: true
        description: Identificador da flag
        schema: {type: string, minLength: 1, maxLength: 128}
        example: checkout
    get:
      summary: Consulta uma flag
      responses:
        '200': {description: Flag encontrada}
        '404': {description: Flag não encontrada}
    put:
      summary: Atualiza uma flag
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: '#/components/schemas/UpdateFlagRequest'}
            example: {enabled: false}
      responses:
        '200': {description: Flag atualizada}
        '400': {description: JSON ou key inválida}
        '404': {description: Flag não encontrada}
    delete:
      summary: Remove uma flag
      responses:
        '204': {description: Flag removida}
        '404': {description: Flag não encontrada}
  /v1/evaluate/{key}:
    get:
      summary: Avalia uma flag no snapshot local
      parameters:
        - name: key
          in: path
          required: true
          description: Identificador da flag a avaliar
          schema: {type: string, minLength: 1}
          example: checkout
      responses:
        '200': {description: Resultado da avaliação}
        '404': {description: Flag não encontrada}
  /healthz:
    get: {summary: Liveness, responses: {'200': {description: Serviço ativo}}}
  /readyz:
    get: {summary: Readiness, responses: {'200': {description: Snapshot sincronizado}, '503': {description: Ainda não sincronizado}}}
  /internal/status:
    get: {summary: Status operacional, responses: {'200': {description: Status}}}
components:
  schemas:
    CreateFlagRequest:
      type: object
      required: [key, enabled]
      properties:
        key: {type: string, minLength: 1, maxLength: 128}
        enabled: {type: boolean}
    UpdateFlagRequest:
      type: object
      required: [enabled]
      properties:
        enabled: {type: boolean}
    Flag:
      type: object
      properties:
        Key: {type: string}
        Enabled: {type: boolean}
        Revision: {type: integer, format: int64}
`

func OpenAPISpecHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write([]byte(openAPISpec))
	})
}

func SwaggerUIHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html lang="pt-BR"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Feature Flag MVP - Swagger</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>window.onload=()=>SwaggerUIBundle({url:'/swagger.yaml',dom_id:'#swagger-ui'});</script></body></html>`))
	})
}
