package utils

import (
	"fmt"
	"net/smtp"
	"strings"
)

// SendMail sends a professional HTML email with organization branding
func SendMail(smtpHost, smtpPort, sender, password, recipient, subject, body string) error {
	auth := smtp.PlainAuth("", sender, password, smtpHost)

	// Professional HTML template
	htmlBody := fmt.Sprintf(`
<html>
<body style="font-family: Arial, sans-serif; background-color: #f6f6f6; padding: 0; margin: 0;">
	<table width="100%%" bgcolor="#f6f6f6" cellpadding="0" cellspacing="0" border="0">
		<tr>
			<td align="center">
				<table width="600" cellpadding="0" cellspacing="0" border="0" style="background: #fff; margin: 40px 0; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.05);">
					<tr>
						<td style="padding: 32px 32px 16px 32px; border-bottom: 1px solid #eee;">
							<img src="https://res.cloudinary.com/drly2lfdz/image/upload/v1766835711/iconptit_gtkanp.png" alt="PTIT Dormitory" width="120" style="display:block; margin-bottom: 16px;">
							<h2 style="margin: 0; color: #2d3e50;">PTIT Dormitory</h2>
						</td>
					</tr>
					<tr>
						<td style="padding: 32px;">
							<p style="font-size: 16px; color: #333; margin-top: 0;">Kính gửi <b>%s</b>,</p>
							<div style="font-size: 15px; color: #333; line-height: 1.7; margin-bottom: 24px;">
								%s
							</div>
							<p style="font-size: 15px; color: #555; margin-bottom: 0;">Trân trọng,<br><b>Ban Quản Lý Ký Túc Xá PTIT</b><br><span style="font-size:13px; color:#888;">Email tự động, vui lòng không trả lời.</span></p>
						</td>
					</tr>
				</table>
			</td>
		</tr>
	</table>
</body>
</html>
`, recipient, body)

	// Compose MIME message for HTML email
	msg := strings.Join([]string{
		fmt.Sprintf("From: PTIT Dormitory <%s>", sender),
		fmt.Sprintf("To: %s", recipient),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
		"",
		htmlBody,
	}, "\r\n")

	return smtp.SendMail(smtpHost+":"+smtpPort, auth, sender, []string{recipient}, []byte(msg))
}
