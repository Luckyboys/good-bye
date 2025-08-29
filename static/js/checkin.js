// 签到页面JavaScript
document.addEventListener('DOMContentLoaded', function() {
    loadStatusInfo();
    
    // 设置当前年份
    document.getElementById('current-year').textContent = new Date().getFullYear();
});

// 加载状态信息
async function loadStatusInfo() {
    try {
        const statusElement = document.getElementById('status-info');
        showLoading(statusElement);
        
        const statusData = await apiRequest('/status');
        
        const statusHtml = `
            <div class="status-info">
                <div class="status-item">
                    <span class="status-indicator ${statusData.data.is_inactive ? 'warning' : 'healthy'}"></span>
                    <strong>当前状态:</strong> ${statusData.data.is_inactive ? '不活跃' : '活跃'}
                </div>
                <div class="status-item">
                    <strong>签到时间:</strong> ${formatDateTime(statusData.data.last_seen)}
                </div>
            </div>
        `;
        
        hideLoading(statusElement, statusHtml);
        
    } catch (error) {
        console.error('加载状态信息失败:', error);
        showNotification('加载状态信息失败', 'error');
    }
}

// 执行签到
async function performCheckin() {
    try {
        const button = document.querySelector('button[onclick="performCheckin()"]');
        const originalText = button.textContent;
        button.disabled = true;
        button.textContent = '签到中...';
        
        const result = await apiRequest('/checkin', {
            method: 'POST'
        });
        
        // 显示结果
        const resultElement = document.getElementById('checkin-result');
        resultElement.className = 'result success';
        resultElement.textContent = result.message;
        resultElement.style.display = 'block';
        
        // 刷新状态信息
        await loadStatusInfo();
        
        // 重置按钮
        button.disabled = false;
        button.textContent = originalText;
        
        showNotification('签到成功！', 'success');
        
        // 3秒后隐藏结果
        setTimeout(() => {
            resultElement.style.display = 'none';
        }, 3000);
        
    } catch (error) {
        console.error('签到失败:', error);
        
        const button = document.querySelector('button[onclick="performCheckin()"]');
        button.disabled = false;
        button.textContent = '立即签到';
        
        const resultElement = document.getElementById('checkin-result');
        resultElement.className = 'result error';
        resultElement.textContent = '签到失败: ' + error.message;
        resultElement.style.display = 'block';
        
        showNotification('签到失败', 'error');
    }
}

// 发送测试邮件
async function sendTestEmail() {
    try {
        const button = document.querySelector('button[onclick="sendTestEmail()"]');
        const originalText = button.textContent;
        button.disabled = true;
        button.textContent = '发送中...';
        
        showNotification('正在发送测试邮件...', 'info');
        
        const result = await apiRequest('/email/test', {
            method: 'POST'
        });
        
        // 显示结果
        const resultElement = document.getElementById('email-result');
        resultElement.className = 'result success';
        resultElement.textContent = result.message;
        resultElement.style.display = 'block';
        
        // 重置按钮
        button.disabled = false;
        button.textContent = originalText;
        
        showNotification('测试邮件发送成功！', 'success');
        
        // 3秒后隐藏结果
        setTimeout(() => {
            resultElement.style.display = 'none';
        }, 3000);
        
    } catch (error) {
        console.error('发送测试邮件失败:', error);
        
        const button = document.querySelector('button[onclick="sendTestEmail()"]');
        button.disabled = false;
        button.textContent = '发送测试邮件';
        
        const resultElement = document.getElementById('email-result');
        resultElement.className = 'result error';
        resultElement.textContent = '发送测试邮件失败: ' + error.message;
        resultElement.style.display = 'block';
        
        showNotification('发送测试邮件失败', 'error');
    }
}

// 确认发送测试遗书
function confirmSendTestWill() {
    // 创建确认对话框
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(0, 0, 0, 0.5);
        display: flex;
        justify-content: center;
        align-items: center;
        z-index: 10000;
    `;

    const modalContent = document.createElement('div');
    modalContent.className = 'modal-content';
    modalContent.style.cssText = `
        background: white;
        padding: 2rem;
        border-radius: 8px;
        max-width: 500px;
        width: 90%;
        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    `;

    modalContent.innerHTML = `
        <h3 style="margin-bottom: 1rem; color: #333;">确认发送测试遗书</h3>
        <p style="margin-bottom: 1.5rem; color: #666; line-height: 1.6;">
            您即将发送一封测试遗书到配置中的第一个收件人地址。<br>
            这是一封测试邮件，邮件标题会标记为【测试】，内容也会明确标识为测试邮件。<br>
            <strong>请确认您真的要发送这封邮件吗？</strong>
        </p>
        <div style="text-align: right;">
            <button onclick="closeModal()" class="btn btn-secondary" style="margin-right: 0.5rem;">取消</button>
            <button onclick="sendTestWill()" class="btn btn-warning">确认发送</button>
        </div>
    `;

    modal.appendChild(modalContent);
    document.body.appendChild(modal);

    // 点击背景关闭模态框
    modal.addEventListener('click', function(e) {
        if (e.target === modal) {
            closeModal();
        }
    });

    // 关闭模态框函数
    window.closeModal = function() {
        modal.remove();
        delete window.closeModal;
        delete window.sendTestWill;
    };

    // 发送测试遗书函数
    window.sendTestWill = async function() {
        closeModal();
        await doSendTestWill();
    };
}

// 发送测试遗书
async function doSendTestWill() {
    try {
        const button = document.querySelector('button[onclick="confirmSendTestWill()"]');
        const originalText = button.textContent;
        button.disabled = true;
        button.textContent = '发送中...';
        
        showNotification('正在发送测试遗书...', 'warning');
        
        const result = await apiRequest('/wills/test-send', {
            method: 'POST'
        });
        
        // 显示结果
        const resultElement = document.getElementById('email-result');
        resultElement.className = 'result success';
        resultElement.textContent = result.message;
        resultElement.style.display = 'block';
        
        // 重置按钮
        button.disabled = false;
        button.textContent = originalText;
        
        showNotification('测试遗书发送成功！', 'success');
        
        // 3秒后隐藏结果
        setTimeout(() => {
            resultElement.style.display = 'none';
        }, 3000);
        
    } catch (error) {
        console.error('发送测试遗书失败:', error);
        
        const button = document.querySelector('button[onclick="confirmSendTestWill()"]');
        button.disabled = false;
        button.textContent = '发送测试遗书';
        
        const resultElement = document.getElementById('email-result');
        resultElement.className = 'result error';
        resultElement.textContent = '发送测试遗书失败: ' + error.message;
        resultElement.style.display = 'block';
        
        showNotification('发送测试遗书失败', 'error');
    }
}