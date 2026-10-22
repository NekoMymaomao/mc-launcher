const translations = {
    zh: {
        launch: '启动', versions: '版本', mods: '模组', accounts: '账号', config: '设置', docs: '文档',
        install: '安装', remove: '删除', verify: '验证', start: '启动', back: '返回',
        version: '版本', username: '用户名', search: '搜索', installMod: '安装模组',
        noVersion: '暂无已安装版本', clickInstall: '点击安装新版本',
        addAccount: '添加账号', offlineLogin: '离线登录', msLogin: 'Microsoft 登录',
        memory: '内存', java: 'Java', theme: '主题', lang: '语言',
        apiDocs: 'API 文档', tutorial: '使用教程', faq: '常见问题'
    },
    en: {
        launch: 'Launch', versions: 'Versions', mods: 'Mods', accounts: 'Accounts', config: 'Config', docs: 'Docs',
        install: 'Install', remove: 'Remove', verify: 'Verify', start: 'Start', back: 'Back',
        version: 'Version', username: 'Username', search: 'Search', installMod: 'Install Mod',
        noVersion: 'No versions installed', clickInstall: 'Click to install new version',
        addAccount: 'Add Account', offlineLogin: 'Offline Login', msLogin: 'Microsoft Login',
        memory: 'Memory', java: 'Java', theme: 'Theme', lang: 'Language',
        apiDocs: 'API Docs', tutorial: 'Tutorial', faq: 'FAQ'
    }
};

let currentLang = localStorage.getItem('lang') || 'zh';

function setLang(lang) {
    currentLang = lang;
    localStorage.setItem('lang', lang);
    render();
}

function t(key) {
    return translations[currentLang]?.[key] || key;
}
