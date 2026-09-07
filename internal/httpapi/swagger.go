package httpapi

import "net/http"

const openAPISpec = `openapi: 3.0.3
info: {title: Feature Flag MVP API, version: 0.1.0}
servers: [{url: http://localhost:8080}]
paths:
  /v1/flags:
    post:
      summary: Cria uma flag
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/FlagRequest'}, example: {key: timeout_ms, value: 1500}}}}
      responses: {'201': {description: Criada}}
    get: {responses: {'200': {description: Lista}}}
  /v1/flags/{key}:
    parameters: [{name: key, in: path, required: true, schema: {type: string}, example: timeout_ms}]
    get: {responses: {'200': {description: Encontrada}, '404': {description: Inexistente}}}
    put:
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/ValueRequest'}, example: {value: {retries: 3}}}}}
      responses: {'200': {description: Atualizada}}
    delete: {responses: {'204': {description: Removida}}}
    patch:
      summary: Ativa ou desativa uma flag sem alterar seu value
      requestBody: {required: true, content: {application/json: {schema: {type: object, required: [enabled], properties: {enabled: {type: boolean}}}, example: {enabled: false}}}}
      responses: {'200': {description: Estado atualizado}, '400': {description: enabled obrigatório}, '404': {description: Inexistente}}}
  /v1/evaluate/{key}:
    get:
      parameters: [{name: key, in: path, required: true, schema: {type: string}}]
      responses: {'200': {description: Valor avaliado}}
components:
  schemas:
    FlagRequest: {type: object, required: [key], properties: {key: {type: string}, value: {}, enabled: {type: boolean}}}
    ValueRequest: {type: object, required: [value], properties: {value: {}}}
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
		_, _ = w.Write([]byte(`<!doctype html><html><head><title>Feature Flag MVP Swagger</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({url:'/swagger.yaml',dom_id:'#swagger-ui'})</script></body></html>`))
	})
}
