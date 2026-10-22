const app = document.getElementById('app');
let currentTab = 'launch';

document.querySelectorAll('.tab').forEach(btn => {
    btn.addEventListener('click', () => {
        document.querySelectorAll('.tab').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        currentTab = btn.dataset.tab;
        render();
    });
});

async function render() {
    switch(currentTab) {
        case 'launch': await renderLaunch(); break;
        case 'versions': await renderVersions(); break;
        case 'mods': renderMods(); break;
        case 'accounts': await renderAccounts(); break;
        case 'config': await renderConfig(); break;
        case 'docs': await renderDocs(); break;
    }
}

async function renderLaunch() {
    const versions = await API.versions.list();
    const data = versions.data || [];
    app.innerHTML = `
        <h2>${t('launch')}</h2>
        <div class="form-group">
            <label>${t('version')}</label>
            <select id="versionSelect">${data.map(v => `<option value="${v}">${v}</option>`).join('')}</select>
        </div>
        <div class="form-group">
            <label>${t('username')}</label>
            <input type="text" id="username" value="Player">
        </div>
        <button class="btn" onclick="doLaunch()" style="width:100%;padding:20px;font-size:20px">${t('start')}</button>
    `;
}

async function doLaunch() {
    const v = document.getElementById('versionSelect').value;
    const u = document.getElementById('username').value;
    await API.launch(v, u);
}

async function renderVersions() {
    const versions = await API.versions.list();
    const data = versions.data || [];
    app.innerHTML = `
        <h2>${t('versions')}</h2>
        ${data.length === 0 ? `<p>${t('noVersion')}</p>` : 
            data.map(v => `<div class="list-item">
                <span>${v}</span>
                <div>
                    <button onclick="doVerify('${v}')">${t('verify')}</button>
                    <button class="btn-danger" onclick="doRemove('${v}')">${t('remove')}</button>
                </div>
            </div>`).join('')}
        <div class="modal" id="installModal">
            <div class="modal-content">
                <span class="close" onclick="closeModal()">&times;</span>
                <h3>${t('install')}</h3>
                <input type="text" id="newVersion" placeholder="1.20.1">
                <button class="btn" onclick="doInstall()" style="width:100%;margin-top:10px">${t('install')}</button>
            </div>
        </div>
        <button class="btn" onclick="document.getElementById('installModal').classList.add('show')" style="margin-top:20px;width:100%">${t('install')}</button>
    `;
}

async function doInstall() {
    const v = document.getElementById('newVersion').value;
    if (!v) return;
    await API.versions.install(v);
    closeModal();
    renderVersions();
}

function doRemove(v) { if (confirm(`删除 ${v}?`)) API.versions.remove(v).then(() => renderVersions()); }
function doVerify(v) { API.versions.verify(v).then(r => alert(r.data?.message || '验证完成')); }
function closeModal() { document.querySelectorAll('.modal').forEach(m => m.classList.remove('show')); }

function renderMods() {
    app.innerHTML = `
        <h2>${t('mods')}</h2>
        <div class="form-group">
            <input type="text" id="searchQuery" placeholder="${t('search')}...">
            <button class="btn" onclick="searchMods()" style="margin-top:10px">${t('search')}</button>
        </div>
        <div id="modResults"></div>
    `;
}

async function searchMods() {
    const q = document.getElementById('searchQuery').value;
    const r = await API.modrinth.search(q, 'mod');
    const results = r.data?.hits || [];
    document.getElementById('modResults').innerHTML = results.map(m => `
        <div class="list-item">
            <div><strong>${m.title}</strong><br><small>${m.description?.substring(0,50)}...</small></div>
            <button class="btn" onclick="installMod('${m.slug}')">${t('install')}</button>
        </div>
    `).join('');
}

async function renderAccounts() {
    const accs = await API.accounts.list();
    app.innerHTML = `
        <h2>${t('accounts')}</h2>
        ${(accs.data || []).map(a => `<div class="list-item"><span>${a.name} (${a.type})</span></div>`).join('') || '<p>暂无账号</p>'}
        <button class="btn" onclick="doOffline()" style="margin-top:10px">${t('offlineLogin')}</button>
        <button class="btn" onclick="doMS()" style="margin-top:10px;margin-left:10px">${t('msLogin')}</button>
    `;
}

async function doOffline() {
    const name = prompt('输入用户名:');
    if (name) { await API.accounts.offline(name); renderAccounts(); }
}

async function renderConfig() {
    const cfg = await API.config.get();
    app.innerHTML = `
        <h2>${t('config')}</h2>
        <div class="form-group"><label>最大内存 (MB)</label><input type="number" id="maxMem" value="${cfg.data?.maxMemory || 4096}"></div>
        <div class="form-group"><label>最小内存 (MB)</label><input type="number" id="minMem" value="${cfg.data?.minMemory || 2048}"></div>
        <div class="form-group"><label>宽度</label><input type="number" id="width" value="${cfg.data?.width || 854}"></div>
        <div class="form-group"><label>高度</label><input type="number" id="height" value="${cfg.data?.height || 480}"></div>
        <button class="btn" onclick="saveConfig()">保存</button>
    `;
}

async function saveConfig() {
    await API.config.set({
        maxMemory: parseInt(document.getElementById('maxMem').value),
        minMemory: parseInt(document.getElementById('minMem').value),
        width: parseInt(document.getElementById('width').value),
        height: parseInt(document.getElementById('height').value)
    });
    alert('配置已保存');
}

async function renderDocs() {
    const apiDoc = await API.docs.api();
    app.innerHTML = `
        <h2>${t('apiDocs')}</h2>
        <div class="api-docs">${apiDoc}</div>
    `;
}

// Init
render();
