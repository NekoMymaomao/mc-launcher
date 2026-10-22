const API_BASE = window.location.hostname === 'localhost' ? 'http://localhost:9876' : '';

async function api(endpoint, options = {}) {
    const url = `${API_BASE}/api${endpoint}`;
    const defaultOptions = {
        headers: { 'Content-Type': 'application/json' },
        ...options
    };
    if (defaultOptions.body && typeof defaultOptions.body === 'object') {
        defaultOptions.body = JSON.stringify(defaultOptions.body);
    }
    const resp = await fetch(url, defaultOptions);
    return await resp.json();
}

const API = {
    ping: () => api('/ping'),
    info: () => api('/info'),
    versions: {
        list: () => api('/versions'),
        install: (v) => api('/versions/install', { method: 'POST', body: { version: v } }),
        remove: (v) => api('/versions/remove', { method: 'POST', body: { version: v } }),
        verify: (v) => api('/versions/verify', { method: 'POST', body: { version: v } })
    },
    launch: (v, u) => api('/launch', { method: 'POST', body: { version: v, username: u } }),
    accounts: {
        list: () => api('/accounts'),
        offline: (n) => api('/accounts/offline', { method: 'POST', body: { name: n } }),
        microsoft: () => api('/accounts/microsoft', { method: 'POST' })
    },
    java: {
        list: () => api('/java'),
        install: (m) => api('/java/install', { method: 'POST', body: { major: m } }),
        set: (p) => api('/java/set', { method: 'PUT', body: { path: p } })
    },
    modrinth: {
        search: (q, t, p) => api(`/modrinth/search?q=${encodeURIComponent(q)}&type=${t || ''}&page=${p || 1}`),
        info: (s) => api(`/modrinth/info/${s}`),
        install: (data) => api('/modrinth/install', { method: 'POST', body: data })
    },
    config: {
        get: () => api('/config'),
        set: (c) => api('/config', { method: 'PUT', body: c })
    },
    servers: {
        list: () => api('/servers'),
        add: (a, n) => api('/servers', { method: 'POST', body: { address: a, name: n } }),
        remove: (a) => api(`/servers/${a}`, { method: 'DELETE' })
    },
    docs: {
        api: () => fetch(`${API_BASE}/docs/api.html`).then(r => r.text()),
        tutorial: () => fetch(`${API_BASE}/docs/tutorial.html`).then(r => r.text()),
        faq: () => fetch(`${API_BASE}/docs/faq.html`).then(r => r.text())
    }
};
