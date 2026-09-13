package httpapi

import "net/http"

func LargePayloadTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html lang="pt-BR"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Feature Flag - Teste 10 MiB</title><style>body{font:16px system-ui;max-width:900px;margin:32px auto;padding:0 16px}button{padding:10px 16px;font-size:16px}input{padding:9px;width:300px}pre{background:#f4f4f4;padding:16px;white-space:pre-wrap}</style></head><body><h1>Teste de payload grande</h1><p>Gera um valor string de exatamente 10 MiB (10.485.760 bytes) no navegador.</p><label>Key: <input id="key" value="payload_10mb"></label> <button id="run">Executar teste</button><pre id="out">Aguardando...</pre><script>
const MiB=10*1024*1024; const out=document.getElementById('out');
function report(name,status,bytes,start,extra=''){out.textContent+=name+': HTTP '+status+' | '+bytes+' bytes | '+(performance.now()-start).toFixed(2)+' ms '+extra+'\n';}
document.getElementById('run').onclick=async()=>{out.textContent='Gerando payload...\n';const key=document.getElementById('key').value;const value='x'.repeat(MiB);const body=JSON.stringify({key,value});const bytes=new TextEncoder().encode(body).byteLength;let start=performance.now();let response=await fetch('/v1/flags',{method:'POST',headers:{'content-type':'application/json'},body});report('POST /v1/flags',response.status,bytes,start);start=performance.now();response=await fetch('/v1/flags/'+encodeURIComponent(key));await response.arrayBuffer();report('GET /v1/flags/'+key,response.status,bytes,start);start=performance.now();response=await fetch('/v1/evaluate/'+encodeURIComponent(key));await response.arrayBuffer();report('GET /v1/evaluate/'+key,response.status,bytes,start);out.textContent+='Payload value: '+value.length+' bytes (10 MiB)\n'+out.textContent;};
</script></body></html>`))
	})
}
