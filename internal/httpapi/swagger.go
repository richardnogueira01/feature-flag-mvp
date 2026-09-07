package httpapi

import "net/http"

const openAPISpec = `openapi: 3.0.3
info:
  title: Feature Flag MVP API
  version: 0.1.0
paths:
  /v1/flags:
    get: {responses: {'200': {description: Lista de flags}}}
    post: {responses: {'201': {description: Flag criada}, '400': {description: Inválido}, '409': {description: Conflito}}}
  /v1/flags/{key}:
    get: {responses: {'200': {description: Flag}, '404': {description: Não encontrado}}}
    put: {responses: {'200': {description: Atualizada}, '404': {description: Não encontrado}}}
    delete: {responses: {'204': {description: Removida}, '404': {description: Não encontrado}}}
  /v1/evaluate/{key}: {get: {responses: {'200': {description: Resultado}, '404': {description: Não encontrado}}}}
  /healthz: {get: {responses: {'200': {description: UP}}}}
  /readyz: {get: {responses: {'200': {description: READY}, '503': {description: NOT_READY}}}}
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
