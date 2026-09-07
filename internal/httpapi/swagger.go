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
        '200':
          description: Lista de flags
    post:
      summary: Cria uma flag
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateFlagRequest'
            example:
              key: timeout_ms
              value: 1500
      responses:
        '201':
          description: Flag criada
        '400':
          description: Requisição inválida
        '409':
          description: Flag já existe
  /v1/flags/{key}:
    parameters:
      - name: key
        in: path
        required: true
        description: Identificador da flag
        schema:
          type: string
          minLength: 1
          maxLength: 128
    get:
      summary: Consulta uma flag
      responses:
        '200':
          description: Flag encontrada
        '404':
          description: Flag não encontrada
    put:
      summary: Atualiza o value da flag
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ValueRequest'
            example:
              value:
                retries: 3
      responses:
        '200':
          description: Flag atualizada
    patch:
      summary: Ativa ou desativa a flag
      description: Altera somente enabled e preserva value.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/EnabledRequest'
            examples:
              activate:
                value:
                  enabled: true
              deactivate:
                value:
                  enabled: false
      responses:
        '200':
          description: Estado atualizado
        '400':
          description: enabled é obrigatório
        '404':
          description: Flag não encontrada
    delete:
      summary: Remove uma flag
      responses:
        '204':
          description: Flag removida
  /v1/evaluate/{key}:
    get:
      summary: Avalia uma flag
      parameters:
        - name: key
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Flag ativa com value
        '204':
          description: Flag desativada, sem conteúdo
        '404':
          description: Flag não encontrada
components:
  schemas:
    CreateFlagRequest:
      type: object
      required:
        - key
      properties:
        key:
          type: string
          minLength: 1
          maxLength: 128
        enabled:
          type: boolean
          description: Compatibilidade legada; padrão true quando value é não booleano
        value:
          description: JSON arbitrário, como string, número, objeto ou lista
    ValueRequest:
      type: object
      required:
        - value
      properties:
        value:
          description: JSON arbitrário
    EnabledRequest:
      type: object
      required:
        - enabled
      properties:
        enabled:
          type: boolean
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
