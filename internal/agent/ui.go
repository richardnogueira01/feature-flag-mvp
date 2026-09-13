package agent

import "net/http"

func UserMenuHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html lang="pt-BR"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Near-cache — menu grande</title><style>body{font:15px system-ui;max-width:1100px;margin:24px auto;padding:0 16px;background:#f7f8fa;color:#17202a}button,input{font:inherit;padding:9px;margin:4px}button{cursor:pointer}#status{white-space:pre-wrap;background:#fff;border:1px solid #ddd;padding:12px;border-radius:8px}.menu{background:#101820;color:#d9f7e8;border-radius:8px;padding:16px;white-space:pre-wrap;overflow:auto;max-height:65vh;line-height:1.35}</style></head><body><h1>Simulador de usuário — menu via Agent</h1><p>O navegador consulta o near-cache local. O menu completo é renderizado abaixo.</p><label>Flag: <input id="key" value="menu_itau_mobile" size="28"></label><button id="load">Carregar menu</button><button id="twice">Carregar duas vezes</button><div id="status">Aguardando...</div><pre id="menu" class="menu">Menu ainda não carregado.</pre><script>
const key=()=>encodeURIComponent(document.getElementById('key').value),status=document.getElementById('status'),menu=document.getElementById('menu');
async function load(){const start=performance.now();const r=await fetch('/v1/evaluate/'+key(),{headers:{'accept-encoding':'gzip'}});const etag=r.headers.get('etag')||'-';if(r.status===204){status.textContent='HTTP 204 — flag desativada';menu.textContent='';return}const body=await r.json();const value=typeof body.value==='string'?body.value:JSON.stringify(body.value,null,2);menu.textContent=value;status.textContent='HTTP '+r.status+' | revision '+body.revision+' | ETag '+etag+' | '+value.length.toLocaleString()+' bytes renderizados | '+(performance.now()-start).toFixed(2)+' ms';}
document.getElementById('load').onclick=()=>load().catch(e=>status.textContent='Erro: '+e);document.getElementById('twice').onclick=async()=>{await load();await load();};
</script></body></html>`))
	})
}
