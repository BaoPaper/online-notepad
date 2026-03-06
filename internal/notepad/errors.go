package notepad

import "github.com/gin-gonic/gin"

func (a *App) renderError(c *gin.Context, statusCode int, title, message string, action *Action) {
	if wantsJSON(c) {
		c.JSON(statusCode, gin.H{"success": false, "message": defaultString(message, title)})
		return
	}

	if wantsHTML(c) {
		fallback := action
		if fallback == nil {
			next := a.defaultErrorAction(c)
			fallback = &next
		}

		c.HTML(statusCode, "error.ejs", gin.H{
			"StatusCode": statusCode,
			"Title":      title,
			"Message":    message,
			"ActionHref": fallback.Href,
			"ActionText": fallback.Text,
			"AppName":    appName,
		})
		return
	}

	c.String(statusCode, defaultString(message, title))
}

func (a *App) defaultErrorAction(c *gin.Context) Action {
	if a.claimsFromContext(c) != nil {
		return Action{Href: "/", Text: "返回首页"}
	}
	return Action{Href: "/login", Text: "返回登录页"}
}
