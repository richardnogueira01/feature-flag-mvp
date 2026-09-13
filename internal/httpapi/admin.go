package httpapi

import "net/http"

// AdminUIHandler serves the /admin page (task 015): a browser-only, no-build admin console for
// the application/environment/configuration/flag hierarchy plus a small operational dashboard.
// It only talks to /v1/applications/**, /internal/status and /metrics on this same origin, and
// optionally to a user-supplied Agent URL for its /readyz status (a different origin/port in
// most deployments, hence the relaxed connect-src below).
func AdminUIHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; connect-src *; img-src 'self'; base-uri 'none'; frame-ancestors 'none'")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(adminHTML))
	})
}

const adminHTML = `<!doctype html>
<html lang="pt-BR">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Feature Flag MVP — Admin</title>
<style>
  :root { color-scheme: light; }
  * { box-sizing: border-box; }
  body { font: 15px/1.4 system-ui, sans-serif; max-width: 1200px; margin: 0 auto; padding: 16px; background: #f7f8fa; color: #17202a; }
  h1 { font-size: 20px; margin: 0 0 4px; }
  h2 { font-size: 16px; margin: 0 0 12px; }
  h3 { font-size: 13px; text-transform: uppercase; letter-spacing: .03em; color: #57606a; margin: 0 0 8px; }
  a { color: #0b5fff; }
  header { display: flex; justify-content: space-between; align-items: baseline; flex-wrap: wrap; gap: 8px; margin-bottom: 16px; }
  header nav a { margin-left: 12px; font-size: 13px; }
  main { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; align-items: start; }
  @media (max-width: 900px) { main { grid-template-columns: 1fr; } }
  section.panel { background: #fff; border: 1px solid #e2e5e9; border-radius: 8px; padding: 12px; min-height: 120px; }
  ul.list { list-style: none; margin: 0 0 10px; padding: 0; max-height: 260px; overflow: auto; }
  ul.list li { padding: 0; margin: 0 0 4px; }
  ul.list button.item { width: 100%; text-align: left; padding: 7px 8px; border: 1px solid transparent; background: #f1f3f5; border-radius: 6px; cursor: pointer; font: inherit; }
  ul.list button.item[aria-current="true"] { border-color: #0b5fff; background: #eaf1ff; font-weight: 600; }
  form.create { display: flex; gap: 6px; }
  input, textarea, select { font: inherit; padding: 7px 8px; border: 1px solid #ccd1d6; border-radius: 6px; width: 100%; }
  textarea { min-height: 64px; font-family: ui-monospace, monospace; font-size: 13px; }
  button { font: inherit; padding: 7px 12px; border: 1px solid #0b5fff; background: #0b5fff; color: #fff; border-radius: 6px; cursor: pointer; }
  button.secondary { background: #fff; color: #0b5fff; }
  button.danger { background: #fff; color: #c0392b; border-color: #c0392b; }
  button:disabled { opacity: .5; cursor: default; }
  table { width: 100%; border-collapse: collapse; font-size: 13px; }
  th, td { text-align: left; padding: 6px 4px; border-bottom: 1px solid #eee; vertical-align: top; }
  .muted { color: #57606a; font-size: 12px; }
  .row { display: flex; gap: 6px; align-items: center; }
  .grow { flex: 1; }
  #toast { position: fixed; bottom: 12px; right: 12px; max-width: 360px; }
  #toast div { background: #17202a; color: #fff; padding: 8px 12px; border-radius: 6px; margin-top: 6px; font-size: 13px; }
  #toast div.error { background: #c0392b; }
  .dash-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-top: 16px; }
  @media (max-width: 700px) { .dash-grid { grid-template-columns: 1fr; } }
  dl.metrics { display: grid; grid-template-columns: auto 1fr; gap: 4px 10px; margin: 0; font-size: 13px; }
  dl.metrics dt { color: #57606a; }
  dl.metrics dd { margin: 0; font-weight: 600; }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 999px; font-size: 12px; font-weight: 600; }
  .badge.up { background: #e3f9e5; color: #1a7f37; }
  .badge.down { background: #ffe9e6; color: #c0392b; }
  .visually-hidden { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0,0,0,0); white-space: nowrap; }
</style>
</head>
<body>
<header>
  <div>
    <h1>Feature Flag MVP</h1>
    <div class="muted">Administração de Applications, Environments, Configurations e Flags.</div>
  </div>
  <nav>
    <a href="/swagger">API (Swagger)</a>
    <a href="/test/large-payload">Teste de payload 10 MiB ↗</a>
  </nav>
</header>

<main>
  <section class="panel" aria-labelledby="h-apps">
    <h3 id="h-apps">Applications</h3>
    <ul class="list" id="list-apps"></ul>
    <form class="create" id="form-app">
      <label class="grow"><span class="visually-hidden">Nome da application</span>
        <input id="input-app" name="name" placeholder="nova application" maxlength="128" required>
      </label>
      <button type="submit">+</button>
    </form>
  </section>

  <section class="panel" aria-labelledby="h-envs">
    <h3 id="h-envs">Environments</h3>
    <p class="muted" id="empty-envs">Selecione uma application.</p>
    <ul class="list" id="list-envs" hidden></ul>
    <form class="create" id="form-env" hidden>
      <label class="grow"><span class="visually-hidden">Nome do environment</span>
        <input id="input-env" name="name" placeholder="novo environment" maxlength="128" required>
      </label>
      <button type="submit">+</button>
    </form>
  </section>

  <section class="panel" aria-labelledby="h-cfgs">
    <h3 id="h-cfgs">Configurations</h3>
    <p class="muted" id="empty-cfgs">Selecione um environment.</p>
    <ul class="list" id="list-cfgs" hidden></ul>
    <form class="create" id="form-cfg" hidden>
      <label class="grow"><span class="visually-hidden">Nome da configuration</span>
        <input id="input-cfg" name="name" placeholder="nova configuration" maxlength="128" required>
      </label>
      <button type="submit">+</button>
    </form>
  </section>

  <section class="panel" aria-labelledby="h-flags">
    <h3 id="h-flags">Flags</h3>
    <p class="muted" id="empty-flags">Selecione uma configuration.</p>
    <div id="flags-body" hidden>
      <dl class="metrics" style="margin-bottom:10px">
        <dt>Revision</dt><dd id="cfg-revision">—</dd>
        <dt>Flags</dt><dd id="cfg-count">—</dd>
        <dt>Última atualização</dt><dd id="cfg-updated">—</dd>
      </dl>
      <table>
        <thead><tr><th>Key</th><th>Enabled</th><th>Value</th><th>Revision</th><th></th></tr></thead>
        <tbody id="flags-table"></tbody>
      </table>
      <h3 style="margin-top:14px">Nova flag</h3>
      <form id="form-flag">
        <div class="row" style="margin-bottom:6px">
          <label class="grow"><span class="visually-hidden">Key</span>
            <input id="new-flag-key" placeholder="key" maxlength="128" required>
          </label>
          <label class="row"><input type="checkbox" id="new-flag-enabled" checked> enabled</label>
        </div>
        <label>Value (JSON, opcional — se vazio usa só "enabled")
          <textarea id="new-flag-value" placeholder='ex: {"limit": 10} ou "texto" ou 42'></textarea>
        </label>
        <p class="muted" id="new-flag-error" role="alert"></p>
        <button type="submit">Criar flag</button>
      </form>
    </div>
  </section>
</main>

<section class="dash-grid">
  <section class="panel" aria-labelledby="h-cp">
    <h3 id="h-cp">Control Plane — dados locais, atualizados a cada 5s</h3>
    <dl class="metrics" id="cp-metrics">
      <dt>Status</dt><dd id="cp-status">—</dd>
      <dt>Revision</dt><dd id="cp-revision">—</dd>
      <dt>Gaps de revisão</dt><dd id="cp-gaps">—</dd>
      <dt>Resyncs (ok/falha)</dt><dd id="cp-resyncs">—</dd>
      <dt>Outbox (ok/falha)</dt><dd id="cp-outbox">—</dd>
      <dt>Avaliações</dt><dd id="cp-evals">—</dd>
      <dt>Latência média de avaliação</dt><dd id="cp-latency">—</dd>
    </dl>
  </section>
  <section class="panel" aria-labelledby="h-agent">
    <h3 id="h-agent">Agent</h3>
    <form class="row" id="form-agent-url" style="margin-bottom:10px">
      <label class="grow">URL do Agent
        <input id="agent-url" placeholder="http://localhost:2772">
      </label>
      <button type="submit" class="secondary">Salvar</button>
    </form>
    <dl class="metrics">
      <dt>Status</dt><dd id="agent-status">—</dd>
      <dt>Revision replicada</dt><dd id="agent-revision">—</dd>
    </dl>
    <p class="muted">Métricas detalhadas de polling, erros e idade do snapshot por Agent chegam
      com a task 014 (near-cache escopado) — ainda não implementada. Este painel mostra apenas
      o resultado de <code>/readyz</code> do Agent informado acima.</p>
  </section>
</section>

<div id="toast" aria-live="polite"></div>

<script>
'use strict';

function esc(s) {
  return String(s).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
}
function toast(message, isError) {
  const box = document.createElement('div');
  if (isError) box.className = 'error';
  box.textContent = message;
  document.getElementById('toast').appendChild(box);
  setTimeout(() => box.remove(), 5000);
}
async function api(path, options) {
  const response = await fetch(path, Object.assign({ headers: { 'content-type': 'application/json' } }, options));
  let body = null;
  const text = await response.text();
  if (text) { try { body = JSON.parse(text); } catch (e) { body = text; } }
  if (!response.ok) {
    const message = (body && body.error) ? body.error : ('HTTP ' + response.status);
    throw new Error(message);
  }
  return body;
}

const state = { app: null, env: null, cfg: null };

async function loadApplications() {
  const list = document.getElementById('list-apps');
  list.innerHTML = '<li class="muted">Carregando...</li>';
  try {
    const data = await api('/v1/applications');
    const apps = data.applications || [];
    list.innerHTML = '';
    for (const a of apps) {
      const li = document.createElement('li');
      const btn = document.createElement('button');
      btn.className = 'item';
      btn.textContent = a.name;
      btn.type = 'button';
      btn.setAttribute('aria-current', String(state.app === a.name));
      btn.onclick = () => selectApplication(a.name);
      li.appendChild(btn);
      list.appendChild(li);
    }
    if (apps.length === 0) list.innerHTML = '<li class="muted">Nenhuma application ainda.</li>';
  } catch (e) {
    list.innerHTML = '';
    toast('Erro ao listar applications: ' + e.message, true);
  }
}

function selectApplication(name) {
  state.app = name; state.env = null; state.cfg = null;
  document.getElementById('form-env').hidden = false;
  document.getElementById('list-envs').hidden = false;
  document.getElementById('empty-envs').hidden = true;
  document.getElementById('form-cfg').hidden = true;
  document.getElementById('list-cfgs').hidden = true;
  document.getElementById('empty-cfgs').hidden = false;
  document.getElementById('flags-body').hidden = true;
  document.getElementById('empty-flags').hidden = false;
  loadApplications();
  loadEnvironments();
}

async function loadEnvironments() {
  if (!state.app) return;
  const list = document.getElementById('list-envs');
  list.innerHTML = '<li class="muted">Carregando...</li>';
  try {
    const data = await api('/v1/applications/' + encodeURIComponent(state.app) + '/environments');
    const envs = data.environments || [];
    list.innerHTML = '';
    for (const e of envs) {
      const li = document.createElement('li');
      const btn = document.createElement('button');
      btn.className = 'item';
      btn.textContent = e.name;
      btn.type = 'button';
      btn.setAttribute('aria-current', String(state.env === e.name));
      btn.onclick = () => selectEnvironment(e.name);
      li.appendChild(btn);
      list.appendChild(li);
    }
    if (envs.length === 0) list.innerHTML = '<li class="muted">Nenhum environment ainda.</li>';
  } catch (e) {
    list.innerHTML = '';
    toast('Erro ao listar environments: ' + e.message, true);
  }
}

function selectEnvironment(name) {
  state.env = name; state.cfg = null;
  document.getElementById('form-cfg').hidden = false;
  document.getElementById('list-cfgs').hidden = false;
  document.getElementById('empty-cfgs').hidden = true;
  document.getElementById('flags-body').hidden = true;
  document.getElementById('empty-flags').hidden = false;
  loadEnvironments();
  loadConfigurations();
}

async function loadConfigurations() {
  if (!state.app || !state.env) return;
  const list = document.getElementById('list-cfgs');
  list.innerHTML = '<li class="muted">Carregando...</li>';
  try {
    const path = '/v1/applications/' + encodeURIComponent(state.app) + '/environments/' + encodeURIComponent(state.env) + '/configurations';
    const data = await api(path);
    const cfgs = data.configurations || [];
    list.innerHTML = '';
    for (const c of cfgs) {
      const li = document.createElement('li');
      const btn = document.createElement('button');
      btn.className = 'item';
      btn.textContent = c.name;
      btn.type = 'button';
      btn.setAttribute('aria-current', String(state.cfg === c.name));
      btn.onclick = () => selectConfiguration(c.name);
      li.appendChild(btn);
      list.appendChild(li);
    }
    if (cfgs.length === 0) list.innerHTML = '<li class="muted">Nenhuma configuration ainda.</li>';
  } catch (e) {
    list.innerHTML = '';
    toast('Erro ao listar configurations: ' + e.message, true);
  }
}

function selectConfiguration(name) {
  state.cfg = name;
  document.getElementById('flags-body').hidden = false;
  document.getElementById('empty-flags').hidden = true;
  loadConfigurations();
  loadFlags();
}

function configurationPath() {
  return '/v1/applications/' + encodeURIComponent(state.app) + '/environments/' + encodeURIComponent(state.env) + '/configurations/' + encodeURIComponent(state.cfg);
}

async function loadFlags() {
  if (!state.app || !state.env || !state.cfg) return;
  const tbody = document.getElementById('flags-table');
  tbody.innerHTML = '<tr><td colspan="5" class="muted">Carregando...</td></tr>';
  try {
    const snapshot = await api(configurationPath());
    document.getElementById('cfg-revision').textContent = snapshot.revision;
    document.getElementById('cfg-count').textContent = (snapshot.flags || []).length;
    document.getElementById('cfg-updated').textContent = snapshot.updated_at ? new Date(snapshot.updated_at).toLocaleString('pt-BR') : '—';
    tbody.innerHTML = '';
    for (const f of (snapshot.flags || [])) {
      tbody.appendChild(renderFlagRow(f));
    }
    if ((snapshot.flags || []).length === 0) {
      tbody.innerHTML = '<tr><td colspan="5" class="muted">Nenhuma flag ainda.</td></tr>';
    }
  } catch (e) {
    tbody.innerHTML = '';
    toast('Erro ao carregar configuration: ' + e.message, true);
  }
}

function renderFlagRow(flag) {
  const tr = document.createElement('tr');

  const keyTd = document.createElement('td');
  keyTd.textContent = flag.key;
  tr.appendChild(keyTd);

  const enabledTd = document.createElement('td');
  const checkbox = document.createElement('input');
  checkbox.type = 'checkbox';
  checkbox.checked = !!flag.enabled;
  checkbox.setAttribute('aria-label', 'Enabled para ' + flag.key);
  checkbox.onchange = async () => {
    checkbox.disabled = true;
    try {
      await api(configurationPath() + '/flags/' + encodeURIComponent(flag.key), {
        method: 'PATCH', body: JSON.stringify({ enabled: checkbox.checked }),
      });
      toast('Flag "' + flag.key + '" atualizada.');
      loadFlags();
    } catch (e) {
      toast('Erro ao atualizar enabled: ' + e.message, true);
      checkbox.checked = !checkbox.checked;
    } finally {
      checkbox.disabled = false;
    }
  };
  enabledTd.appendChild(checkbox);
  tr.appendChild(enabledTd);

  const valueTd = document.createElement('td');
  const valueInput = document.createElement('textarea');
  valueInput.style.minHeight = '32px';
  valueInput.value = flag.value === undefined ? '' : JSON.stringify(flag.value);
  valueTd.appendChild(valueInput);
  const valueError = document.createElement('div');
  valueError.className = 'muted';
  valueError.setAttribute('role', 'alert');
  valueTd.appendChild(valueError);
  const saveBtn = document.createElement('button');
  saveBtn.type = 'button';
  saveBtn.textContent = 'Salvar value';
  saveBtn.className = 'secondary';
  saveBtn.style.marginTop = '4px';
  saveBtn.onclick = async () => {
    valueError.textContent = '';
    let parsed;
    try {
      parsed = JSON.parse(valueInput.value);
    } catch (e) {
      valueError.textContent = 'JSON inválido: ' + e.message;
      return;
    }
    saveBtn.disabled = true;
    try {
      await api(configurationPath() + '/flags/' + encodeURIComponent(flag.key), {
        method: 'PUT', body: JSON.stringify({ value: parsed }),
      });
      toast('Value de "' + flag.key + '" atualizado.');
      loadFlags();
    } catch (e) {
      toast('Erro ao salvar value: ' + e.message, true);
    } finally {
      saveBtn.disabled = false;
    }
  };
  valueTd.appendChild(saveBtn);
  tr.appendChild(valueTd);

  const revTd = document.createElement('td');
  revTd.textContent = flag.revision;
  tr.appendChild(revTd);

  const actionsTd = document.createElement('td');
  const delBtn = document.createElement('button');
  delBtn.type = 'button';
  delBtn.className = 'danger';
  delBtn.textContent = 'Remover';
  delBtn.onclick = async () => {
    if (!confirm('Remover a flag "' + flag.key + '"?')) return;
    delBtn.disabled = true;
    try {
      await api(configurationPath() + '/flags/' + encodeURIComponent(flag.key), { method: 'DELETE' });
      toast('Flag "' + flag.key + '" removida.');
      loadFlags();
    } catch (e) {
      toast('Erro ao remover flag: ' + e.message, true);
      delBtn.disabled = false;
    }
  };
  actionsTd.appendChild(delBtn);
  tr.appendChild(actionsTd);

  return tr;
}

document.getElementById('form-app').addEventListener('submit', async (ev) => {
  ev.preventDefault();
  const input = document.getElementById('input-app');
  try {
    await api('/v1/applications', { method: 'POST', body: JSON.stringify({ name: input.value }) });
    input.value = '';
    toast('Application criada.');
    loadApplications();
  } catch (e) { toast('Erro ao criar application: ' + e.message, true); }
});

document.getElementById('form-env').addEventListener('submit', async (ev) => {
  ev.preventDefault();
  const input = document.getElementById('input-env');
  try {
    await api('/v1/applications/' + encodeURIComponent(state.app) + '/environments', {
      method: 'POST', body: JSON.stringify({ name: input.value }),
    });
    input.value = '';
    toast('Environment criado.');
    loadEnvironments();
  } catch (e) { toast('Erro ao criar environment: ' + e.message, true); }
});

document.getElementById('form-cfg').addEventListener('submit', async (ev) => {
  ev.preventDefault();
  const input = document.getElementById('input-cfg');
  try {
    await api('/v1/applications/' + encodeURIComponent(state.app) + '/environments/' + encodeURIComponent(state.env) + '/configurations', {
      method: 'POST', body: JSON.stringify({ name: input.value }),
    });
    input.value = '';
    toast('Configuration criada.');
    loadConfigurations();
  } catch (e) { toast('Erro ao criar configuration: ' + e.message, true); }
});

document.getElementById('form-flag').addEventListener('submit', async (ev) => {
  ev.preventDefault();
  const keyInput = document.getElementById('new-flag-key');
  const enabledInput = document.getElementById('new-flag-enabled');
  const valueInput = document.getElementById('new-flag-value');
  const errorBox = document.getElementById('new-flag-error');
  errorBox.textContent = '';
  let body;
  if (valueInput.value.trim()) {
    let parsed;
    try {
      parsed = JSON.parse(valueInput.value);
    } catch (e) {
      errorBox.textContent = 'JSON inválido: ' + e.message;
      return;
    }
    body = { key: keyInput.value, value: parsed };
  } else {
    body = { key: keyInput.value, enabled: enabledInput.checked };
  }
  try {
    await api(configurationPath() + '/flags', { method: 'POST', body: JSON.stringify(body) });
    keyInput.value = ''; valueInput.value = ''; enabledInput.checked = true;
    toast('Flag "' + body.key + '" criada.');
    loadFlags();
  } catch (e) { toast('Erro ao criar flag: ' + e.message, true); }
});

// --- Dashboard -------------------------------------------------------------

function parsePrometheusText(text) {
  const metrics = [];
  for (const line of text.split('\n')) {
    if (!line || line.startsWith('#')) continue;
    const match = line.match(/^([a-zA-Z_:][a-zA-Z0-9_:]*)(\{[^}]*\})?\s+(\S+)$/);
    if (!match) continue;
    const labels = {};
    if (match[2]) {
      for (const pair of match[2].slice(1, -1).split(',')) {
        const kv = pair.split('=');
        if (kv.length === 2) labels[kv[0].trim()] = kv[1].trim().replace(/^"|"$/g, '');
      }
    }
    metrics.push({ name: match[1], labels, value: parseFloat(match[3]) });
  }
  return metrics;
}
function sumWhere(metrics, name, predicate) {
  return metrics.filter(m => m.name === name && (!predicate || predicate(m.labels))).reduce((acc, m) => acc + (isNaN(m.value) ? 0 : m.value), 0);
}

async function refreshControlPlaneDashboard() {
  try {
    const status = await api('/internal/status');
    document.getElementById('cp-status').innerHTML = '<span class="badge ' + (status.synchronized ? 'up' : 'down') + '">' + esc(status.status) + '</span>';
    document.getElementById('cp-revision').textContent = status.revision;
  } catch (e) {
    document.getElementById('cp-status').innerHTML = '<span class="badge down">indisponível</span>';
  }
  try {
    const text = await fetch('/metrics').then(r => r.text());
    const metrics = parsePrometheusText(text);
    document.getElementById('cp-gaps').textContent = sumWhere(metrics, 'feature_flag_revision_gaps_total');
    document.getElementById('cp-resyncs').textContent =
      sumWhere(metrics, 'feature_flag_resyncs_total', l => l.result === 'success') + ' / ' +
      sumWhere(metrics, 'feature_flag_resyncs_total', l => l.result === 'failure');
    document.getElementById('cp-outbox').textContent =
      sumWhere(metrics, 'feature_flag_outbox_events_total', l => l.result === 'success') + ' / ' +
      sumWhere(metrics, 'feature_flag_outbox_events_total', l => l.result === 'failure');
    document.getElementById('cp-evals').textContent =
      ['enabled', 'disabled', 'missing'].map(r => r + ':' + sumWhere(metrics, 'feature_flag_evaluations_total', l => l.result === r)).join('  ');
    const sum = sumWhere(metrics, 'feature_flag_evaluation_duration_seconds_sum');
    const count = sumWhere(metrics, 'feature_flag_evaluation_duration_seconds_count');
    document.getElementById('cp-latency').textContent = count > 0 ? ((sum / count) * 1000).toFixed(3) + ' ms' : '—';
  } catch (e) {
    // /metrics is best-effort for this dashboard; leave previous values on transient failure.
  }
}

function agentUrl() {
  return (localStorage.getItem('ff_agent_url') || '').replace(/\/$/, '');
}
document.getElementById('agent-url').value = agentUrl();
document.getElementById('form-agent-url').addEventListener('submit', (ev) => {
  ev.preventDefault();
  localStorage.setItem('ff_agent_url', document.getElementById('agent-url').value.trim());
  toast('URL do Agent salva.');
  refreshAgentDashboard();
});

async function refreshAgentDashboard() {
  const url = agentUrl();
  if (!url) {
    document.getElementById('agent-status').innerHTML = '<span class="muted">nenhuma URL configurada</span>';
    document.getElementById('agent-revision').textContent = '—';
    return;
  }
  try {
    const response = await fetch(url + '/readyz');
    const body = await response.json().catch(() => null);
    const ready = response.ok;
    document.getElementById('agent-status').innerHTML = '<span class="badge ' + (ready ? 'up' : 'down') + '">' + (body && body.status ? esc(body.status) : (ready ? 'READY' : 'DEGRADED')) + '</span>';
    document.getElementById('agent-revision').textContent = body && body.revision !== undefined ? body.revision : '—';
  } catch (e) {
    document.getElementById('agent-status').innerHTML = '<span class="badge down">indisponível</span>';
    document.getElementById('agent-revision').textContent = '—';
  }
}

function refreshDashboard() {
  refreshControlPlaneDashboard();
  refreshAgentDashboard();
}

loadApplications();
refreshDashboard();
setInterval(refreshDashboard, 5000);
</script>
</body>
</html>`
