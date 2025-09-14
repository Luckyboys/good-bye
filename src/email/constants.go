package email

import (
	"crypto/tls"
	"time"
)

// 邮件服务常量定义

// SMTP 相关常量
const (
	DefaultSMTPPort      = 587
	DefaultTLSServerName = "smtp.gmail.com"
	DefaultTimeout       = 30 * time.Second
)

// TLS 相关常量
const (
	TLSMinVersion = tls.VersionTLS12
)

// 重试相关常量
const (
	DefaultMaxRetryCount    = 20
	DefaultMaxRetryDuration = 7 * 24 * time.Hour // 1周
	InitialRetryDelay       = 1 * time.Second
	MaxRetryDelay           = 24 * time.Hour
	RetryPauseDelay         = 100 * time.Millisecond
)

// 邮件内容类型常量
const (
	ContentTypeHTML = "text/html; charset=UTF-8"
	ContentTypeText = "text/plain; charset=UTF-8"
	MIMEVersion     = "1.0"
)

// 邮件主题常量
const (
	SubjectTestEmail     = "生存确认服务测试邮件"
	SubjectReminderEmail = "生存确认服务 - 签到提醒"
	SubjectWillEmail     = "遗书通知 - 重要信息"
	SubjectTestWillEmail = "【测试】遗书通知 - 重要信息"
	SubjectConfigTest    = "邮件配置测试"
)

// Markdown 检测常量
const (
	HeaderH1       = "# "
	HeaderH2       = "## "
	HeaderH3       = "### "
	UnorderedList1 = "- "
	UnorderedList2 = "* "
	OrderedList    = "1. "
	LinkStart      = "["
	ImageStart     = "!["
	CodeBlock      = "```"
	InlineCode     = "`"
	Quote          = "> "
	Bold           = "**"
	Italic         = "*"
	HorizontalRule = "---"
)

// 邮件状态消息常量
const (
	MessageTestEmailNotConfigured      = "测试邮箱地址未配置"
	MessageWillRecipientsNotConfigured = "未配置收件人邮箱"
	MessageReadPosthumousFailed        = "读取遗书文件失败"
	MessageAllEmailsFailed             = "所有邮件发送失败"
	MessageEmailConfigValidationFailed = "邮件配置验证失败"
	MessageConnectSMTPFailed           = "连接SMTP服务器失败"
	MessageEnableTLSFailed             = "启用TLS失败"
	MessageSMTPAuthFailed              = "SMTP认证失败"
	MessageSetFromFailed               = "设置发件人失败"
	MessageSetToFailed                 = "设置收件人失败"
	MessageGetDataFailed               = "获取邮件写入器失败"
	MessageWriteContentFailed          = "写入邮件内容失败"
	MessageRetryManagerNotInit         = "Retry manager is not initialized"
	MessageReminderFunctionDisabled    = "提醒功能未启用"
)

// 邮件错误消息常量
const (
	ErrorTestEmailNotConfigured   = "test email not configured"
	ErrorNoRecipientsConfigured   = "no recipients configured"
	ErrorRetryManagerNotInit      = "retry manager is not initialized"
	ErrorReminderFunctionDisabled = "reminder function is disabled"
	ErrorSMTPHostRequired         = "SMTP host is required"
	ErrorInvalidSMTPPort          = "invalid SMTP port"
	ErrorUsernameRequired         = "username is required"
	ErrorPasswordRequired         = "password is required"
	ErrorFromEmailRequired        = "from email is required"
	ErrorMaxRetryCountReached     = "max retry count (%d) reached"
	ErrorMaxRetryDurationReached  = "max retry duration (%v) reached"
)

// 时间格式常量
const (
	TimeFormatRFC3339 = "2006-01-02 15:04:05"
)

// AllowedHTMLTags HTML允许的标签常量
var AllowedHTMLTags = []string{
	"h1", "h2", "h3", "h4", "h5", "h6", "p", "br", "hr",
	"ul", "ol", "li", "blockquote", "pre", "code",
	"strong", "em", "i", "b", "a", "img",
	"table", "thead", "tbody", "tr", "th", "td",
}

// MarkdownIndicators Markdown检测指示器常量
var MarkdownIndicators = []string{
	HeaderH1, HeaderH2, HeaderH3,
	UnorderedList1, UnorderedList2, OrderedList,
	LinkStart, ImageStart, CodeBlock, InlineCode, Quote,
	Bold, Italic, HorizontalRule,
}

// HTML 邮件模板常量
const (
	// 测试邮件模板
	TestEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>生存确认服务测试邮件</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f4f4f4; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #fff; }
        .footer { background-color: #f4f4f4; padding: 10px; text-align: center; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>生存确认服务测试邮件</h1>
        </div>
        <div class="content">
            <p>您好！</p>
            <p>这是一封来自生存确认服务的测试邮件。</p>
            <p>如果您收到这封邮件，说明邮件服务配置正常。</p>
            <p>发送时间：%s</p>
        </div>
        <div class="footer">
            <p>生存确认服务 - 自动化邮件通知系统</p>
        </div>
    </div>
</body>
</html>`

	// 提醒邮件模板
	ReminderEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>生存确认服务 - 签到提醒</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #fff3cd; padding: 20px; text-align: center; border: 1px solid #ffeaa7; }
        .content { padding: 20px; background-color: #fff; border: 1px solid #ddd; border-top: none; }
        .warning { background-color: #f8d7da; padding: 15px; border: 1px solid #f5c6cb; border-radius: 4px; margin: 20px 0; }
        .footer { background-color: #f8f9fa; padding: 15px; text-align: center; font-size: 12px; border: 1px solid #ddd; border-top: none; }
        .btn { display: inline-block; padding: 10px 20px; background-color: #007bff; color: white; text-decoration: none; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="color: #856404;">🔔 生存确认服务 - 签到提醒</h1>
        </div>
        <div class="content">
            <p>您好！</p>
            <p>这是一封来自生存确认服务的<strong>签到提醒邮件</strong>。</p>
            
            <div class="warning">
                <h3>⚠️ 重要提醒</h3>
                <p>您已经超过 <strong>%s</strong> 没有进行签到操作。</p>
                <p>如果您在 <strong>%s</strong> 之前仍然没有签到，系统将自动发送您的遗书邮件。</p>
            </div>
            
            <h4>请立即采取以下操作：</h4>
            <ol>
                <li><a href="https://good-bye.eachol.ren:12443">访问生存确认服务</a></li>
                <li>进行签到操作以确认您的安全状态</li>
                <li>确保定期签到以避免误触发遗书发送</li>
            </ol>
            
            <p><strong>提醒：</strong>请勿忽视此提醒，一旦遗书邮件发送，将无法撤销。</p>
        </div>
        <div class="footer">
            <p>生存确认服务 - 自动化邮件通知系统</p>
            <p>发送时间：%s</p>
        </div>
    </div>
</body>
</html>`

	// 遗书邮件模板样式
	WillEmailStyles = `        body { 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6; 
            color: #333; 
            max-width: 800px; 
            margin: 0 auto; 
            padding: 20px;
            background-color: #fff;
        }
        h1, h2, h3, h4, h5, h6 { color: #2c3e50; margin-top: 2em; margin-bottom: 1em; }
        h1 { border-bottom: 2px solid #3498db; padding-bottom: 0.3em; }
        h2 { border-bottom: 1px solid #bdc3c7; padding-bottom: 0.2em; }
        p { margin-bottom: 1em; }
        pre { 
            background-color: #f8f9fa; 
            border: 1px solid #e9ecef; 
            border-radius: 4px; 
            padding: 16px; 
            overflow-x: auto;
            font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
        }
        code { 
            background-color: #f8f9fa; 
            border: 1px solid #e9ecef; 
            border-radius: 3px; 
            padding: 2px 4px; 
            font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
        }
        blockquote { 
            border-left: 4px solid #3498db; 
            margin: 1em 0; 
            padding-left: 1em; 
            color: #7f8c8d; 
        }
        table { 
            border-collapse: collapse; 
            width: 100%%; 
            margin: 1em 0; 
        }
        th, td { 
            border: 1px solid #ddd; 
            padding: 8px; 
            text-align: left; 
        }
        th { 
            background-color: #f8f9fa; 
            font-weight: bold; 
        }
        img { max-width: 100%%; height: auto; }
        a { color: #3498db; text-decoration: none; }
        a:hover { text-decoration: underline; }`

	// 遗书邮件模板包装器
	WillEmailTemplateWrapper = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>遗书</title>
    <style>
%s
    </style>
</head>
<body>
    %s
</body>
</html>`
)
