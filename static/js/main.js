// 主页面JavaScript
document.addEventListener('DOMContentLoaded', function() {
    loadDashboard();
});

// 加载仪表板数据
async function loadDashboard() {
    try {
        // 加载服务状态
        const statusElement = document.getElementById('service-status');
        showLoading(statusElement);
        
        const statusData = await apiRequest('/status');
        const healthData = await apiRequest('/health');
        
        const statusHtml = `
            <div class="status-item">
                <span class="status-indicator ${healthData.data.status === 'healthy' ? 'healthy' : 'error'}"></span>
                <span>服务状态: ${healthData.data.status === 'healthy' ? '正常' : '异常'}</span>
            </div>
            <div class="status-item">
                <span>最后签到: ${formatRelativeTime(statusData.data.last_seen)}</span>
            </div>
            <div class="status-item">
                <span>不活跃时长: ${statusData.data.inactive_duration}</span>
            </div>
            <div class="status-item">
                <span>状态: ${statusData.data.is_inactive ? '不活跃' : '活跃'}</span>
            </div>
        `;
        
        hideLoading(statusElement, statusHtml);
        
        // 加载系统信息
        const infoElement = document.getElementById('system-info');
        showLoading(infoElement);
        
        const statsData = await apiRequest('/stats');
        
        const infoHtml = `
            <div class="info-item">
                <strong>检查间隔:</strong> ${statsData.data.system_settings.check_interval} 小时
            </div>
            <div class="info-item">
                <strong>最大不活跃天数:</strong> ${statsData.data.system_settings.max_inactive_days} 天
            </div>
            <div class="info-item">
                <strong>通知状态:</strong> ${statsData.data.system_settings.enable_notification ? '已启用' : '已禁用'}
            </div>
            <div class="info-item">
                <strong>遗书数量:</strong> ${statsData.data.will_stats.total} 个
                (已发送: ${statsData.data.will_stats.sent}, 未发送: ${statsData.data.will_stats.unsent})
            </div>
            <div class="info-item">
                <strong>邮件统计:</strong> ${statsData.data.email_stats.total} 封
                (成功: ${statsData.data.email_stats.sent}, 失败: ${statsData.data.email_stats.failed})
            </div>
        `;
        
        hideLoading(infoElement, infoHtml);
        
    } catch (error) {
        console.error('加载仪表板失败:', error);
        showNotification('加载仪表板数据失败', 'error');
    }
}

// 发送测试邮件
async function sendTestEmail() {
    try {
        showNotification('正在发送测试邮件...', 'info');
        
        const result = await apiRequest('/email/test', {
            method: 'POST'
        });
        
        showNotification(result.message, 'success');
        
    } catch (error) {
        console.error('发送测试邮件失败:', error);
        showNotification('发送测试邮件失败: ' + error.message, 'error');
    }
}

// 检查遗书文件
async function checkWillFile() {
    try {
        showNotification('正在检查遗书文件...', 'info');
        
        const result = await apiRequest('/wills');
        
        if (result.success) {
            const message = `遗书文件存在，内容长度: ${result.data.content.length} 字符`;
            showNotification(message, 'success');
        } else {
            showNotification(result.message, 'error');
        }
        
    } catch (error) {
        console.error('检查遗书文件失败:', error);
        showNotification('检查遗书文件失败: ' + error.message, 'error');
    }
}