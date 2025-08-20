// 配置页面JavaScript
document.addEventListener('DOMContentLoaded', function() {
    loadSystemSettings();
    loadEmailConfig();
    loadWills();
    
    // 绑定表单提交事件
    document.getElementById('system-form').addEventListener('submit', saveSystemSettings);
    document.getElementById('email-form').addEventListener('submit', saveEmailConfig);
    document.getElementById('will-form').addEventListener('submit', saveWill);
});

// 标签切换
function showTab(tabName) {
    // 隐藏所有标签内容
    const tabContents = document.querySelectorAll('.tab-content');
    tabContents.forEach(content => {
        content.classList.remove('active');
    });
    
    // 移除所有标签按钮的激活状态
    const tabBtns = document.querySelectorAll('.tab-btn');
    tabBtns.forEach(btn => {
        btn.classList.remove('active');
    });
    
    // 显示选中的标签内容
    document.getElementById(`${tabName}-tab`).classList.add('active');
    
    // 激活对应的标签按钮
    event.target.classList.add('active');
}

// 加载系统设置
async function loadSystemSettings() {
    try {
        const settings = await apiRequest('/settings');
        
        document.getElementById('check_interval').value = settings.data.check_interval;
        document.getElementById('max_inactive_days').value = settings.data.max_inactive_days;
        document.getElementById('enable_notification').checked = settings.data.enable_notification;
        document.getElementById('timezone').value = settings.data.timezone;
        
    } catch (error) {
        console.error('加载系统设置失败:', error);
        showNotification('加载系统设置失败', 'error');
    }
}

// 保存系统设置
async function saveSystemSettings(event) {
    event.preventDefault();
    
    try {
        const formData = new FormData(event.target);
        const settings = {
            check_interval: parseInt(formData.get('check_interval')),
            max_inactive_days: parseInt(formData.get('max_inactive_days')),
            enable_notification: formData.get('enable_notification') === 'on',
            timezone: formData.get('timezone')
        };
        
        await apiRequest('/settings', {
            method: 'PUT',
            body: JSON.stringify(settings)
        });
        
        showNotification('系统设置保存成功', 'success');
        
    } catch (error) {
        console.error('保存系统设置失败:', error);
        showNotification('保存系统设置失败', 'error');
    }
}

// 加载邮件配置
async function loadEmailConfig() {
    try {
        const config = await apiRequest('/email/config');
        
        document.getElementById('smtp_host').value = config.data.smtp_host;
        document.getElementById('smtp_port').value = config.data.smtp_port;
        document.getElementById('username').value = config.data.username;
        document.getElementById('password').value = config.data.password;
        document.getElementById('from_email').value = config.data.from_email;
        document.getElementById('test_email').value = config.data.test_email;
        
    } catch (error) {
        console.error('加载邮件配置失败:', error);
        showNotification('加载邮件配置失败', 'error');
    }
}

// 保存邮件配置
async function saveEmailConfig(event) {
    event.preventDefault();
    
    try {
        const formData = new FormData(event.target);
        const config = {
            smtp_host: formData.get('smtp_host'),
            smtp_port: parseInt(formData.get('smtp_port')),
            username: formData.get('username'),
            password: formData.get('password'),
            from_email: formData.get('from_email'),
            test_email: formData.get('test_email')
        };
        
        await apiRequest('/email/config', {
            method: 'PUT',
            body: JSON.stringify(config)
        });
        
        showNotification('邮件配置保存成功', 'success');
        
    } catch (error) {
        console.error('保存邮件配置失败:', error);
        showNotification('保存邮件配置失败', 'error');
    }
}

// 测试邮件配置
async function testEmailConfig() {
    try {
        const formData = new FormData(document.getElementById('email-form'));
        const config = {
            smtp_host: formData.get('smtp_host'),
            smtp_port: parseInt(formData.get('smtp_port')),
            username: formData.get('username'),
            password: formData.get('password'),
            from_email: formData.get('from_email'),
            test_email: formData.get('test_email')
        };
        
        showNotification('正在测试邮件配置...', 'info');
        
        const result = await apiRequest('/email/config/test', {
            method: 'POST',
            body: JSON.stringify(config)
        });
        
        showNotification(result.message, 'success');
        
    } catch (error) {
        console.error('测试邮件配置失败:', error);
        showNotification('测试邮件配置失败: ' + error.message, 'error');
    }
}

// 加载遗书列表
async function loadWills() {
    try {
        const willsElement = document.getElementById('wills-list');
        showLoading(willsElement);
        
        const wills = await apiRequest('/wills');
        
        if (wills.data.length === 0) {
            willsElement.innerHTML = '<p>暂无遗书，点击上方按钮添加</p>';
            return;
        }
        
        const willsHtml = wills.data.map(will => `
            <div class="will-item">
                <h3>${will.title}</h3>
                <p>${will.content.substring(0, 100)}${will.content.length > 100 ? '...' : ''}</p>
                <div class="meta">
                    <span>创建时间: ${formatDateTime(will.created_at)}</span>
                    ${will.is_sent ? `<span style="color: #28a745;">已发送: ${formatDateTime(will.sent_at)}</span>` : '<span style="color: #ffc107;">未发送</span>'}
                </div>
                <div class="actions">
                    <button onclick="editWill(${will.id}, '${will.title.replace(/'/g, "\\'")}', \`${will.content.replace(/`/g, '\\`').replace(/\$/g, '\\$')}\`)" class="btn btn-secondary">编辑</button>
                    <button onclick="deleteWill(${will.id})" class="btn btn-danger">删除</button>
                </div>
            </div>
        `).join('');
        
        hideLoading(willsElement, willsHtml);
        
    } catch (error) {
        console.error('加载遗书列表失败:', error);
        document.getElementById('wills-list').innerHTML = '<p>加载遗书列表失败</p>';
        showNotification('加载遗书列表失败', 'error');
    }
}

// 显示添加遗书表单
function showAddWillForm() {
    document.getElementById('will-modal-title').textContent = '添加遗书';
    document.getElementById('will_id').value = '';
    document.getElementById('will_title').value = '';
    document.getElementById('will_content').value = '';
    document.getElementById('will-modal').style.display = 'block';
}

// 编辑遗书
function editWill(id, title, content) {
    document.getElementById('will-modal-title').textContent = '编辑遗书';
    document.getElementById('will_id').value = id;
    document.getElementById('will_title').value = title;
    document.getElementById('will_content').value = content;
    document.getElementById('will-modal').style.display = 'block';
}

// 关闭遗书模态框
function closeWillModal() {
    document.getElementById('will-modal').style.display = 'none';
}

// 保存遗书
async function saveWill(event) {
    event.preventDefault();
    
    try {
        const formData = new FormData(event.target);
        const willId = formData.get('will_id');
        
        const willData = {
            title: formData.get('will_title'),
            content: formData.get('will_content')
        };
        
        if (willId) {
            // 更新遗书
            await apiRequest(`/wills/${willId}`, {
                method: 'PUT',
                body: JSON.stringify(willData)
            });
            showNotification('遗书更新成功', 'success');
        } else {
            // 创建遗书
            await apiRequest('/wills', {
                method: 'POST',
                body: JSON.stringify(willData)
            });
            showNotification('遗书创建成功', 'success');
        }
        
        closeWillModal();
        loadWills();
        
    } catch (error) {
        console.error('保存遗书失败:', error);
        showNotification('保存遗书失败: ' + error.message, 'error');
    }
}

// 删除遗书
async function deleteWill(id) {
    if (!confirm('确定要删除这个遗书吗？此操作不可恢复。')) {
        return;
    }
    
    try {
        await apiRequest(`/wills/${id}`, {
            method: 'DELETE'
        });
        
        showNotification('遗书删除成功', 'success');
        loadWills();
        
    } catch (error) {
        console.error('删除遗书失败:', error);
        showNotification('删除遗书失败: ' + error.message, 'error');
    }
}

// 点击模态框外部关闭
window.onclick = function(event) {
    const modal = document.getElementById('will-modal');
    if (event.target === modal) {
        closeWillModal();
    }
}