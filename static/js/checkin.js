// 签到页面JavaScript
document.addEventListener('DOMContentLoaded', function() {
    loadStatusInfo();
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
                    <strong>最后签到时间:</strong> ${formatDateTime(statusData.data.last_seen)}
                </div>
                <div class="status-item">
                    <strong>不活跃时长:</strong> ${statusData.data.inactive_duration}
                </div>
                <div class="status-item">
                    <strong>最大不活跃天数:</strong> ${statusData.data.max_inactive_days} 天
                </div>
                <div class="status-item">
                    <strong>检查间隔:</strong> ${statusData.data.check_interval} 小时
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