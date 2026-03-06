package notepad

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (a *App) handleNotePage(c *gin.Context) {
	id := c.Param("id")
	note, content, found, err := a.store.getNoteContent(id)
	if err != nil {
		a.renderError(c, http.StatusInternalServerError, "服务器错误", "读取笔记失败，请稍后再试。", nil)
		return
	}
	if !found {
		a.renderError(c, http.StatusNotFound, "笔记不存在", "该笔记可能已被删除或不存在。", &Action{Href: "/", Text: "返回首页"})
		return
	}

	notes, err := a.store.listNotes()
	if err != nil {
		a.renderError(c, http.StatusInternalServerError, "服务器错误", "读取笔记列表失败，请稍后再试。", nil)
		return
	}

	c.HTML(http.StatusOK, "note.ejs", gin.H{
		"AppName":      appName,
		"Note":         content,
		"Notes":        notes,
		"CurrentID":    id,
		"CurrentTitle": note.Title,
	})
}

func (a *App) handleNoteContent(c *gin.Context) {
	id := c.Param("id")
	note, content, found, err := a.store.getNoteContent(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal Server Error"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"id":        note.ID,
		"title":     note.Title,
		"content":   content,
		"createdAt": note.CreatedAt,
		"updatedAt": effectiveUpdatedAt(note),
	})
}

func (a *App) handleNotesList(c *gin.Context) {
	notes, err := a.store.listNotes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, notes)
}

func (a *App) handleCreateNote(c *gin.Context) {
	var payload struct {
		Title string `json:"title" form:"title"`
	}
	_ = c.ShouldBind(&payload)

	note, err := a.store.createNote(payload.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, note)
}

func (a *App) handleDeleteNote(c *gin.Context) {
	ok, err := a.store.deleteNote(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal Server Error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (a *App) handleRenameNote(c *gin.Context) {
	var payload struct {
		Title string `json:"title" form:"title"`
	}
	_ = c.ShouldBind(&payload)

	title, ok, err := a.store.renameNote(c.Param("id"), payload.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal Server Error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "title": title})
}

func (a *App) handleSaveNote(c *gin.Context) {
	ok, err := a.store.saveNoteContent(c.Param("id"), c.PostForm("note"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal Server Error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (a *App) handleUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "No file"})
		return
	}
	if file.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "File too large"})
		return
	}

	if err := os.MkdirAll(a.store.uploadsDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal Server Error"})
		return
	}

	originalName := filepath.Base(file.Filename)
	storedName := uuid.NewString() + strings.ToLower(filepath.Ext(originalName))
	destination := filepath.Join(a.store.uploadsDir, storedName)
	if err := c.SaveUploadedFile(file, destination); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Upload failed"})
		return
	}

	noteID := c.PostForm("noteId")
	if noteID != "" {
		if err := a.store.addAttachment(noteID, storedName); err != nil {
			_ = os.Remove(destination)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal Server Error"})
			return
		}
	}

	url := "/uploads/" + storedName
	markdown := fmt.Sprintf("[%s](%s)", originalName, url)
	if imagePattern.MatchString(originalName) {
		markdown = fmt.Sprintf("![%s](%s)", originalName, url)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"url":          url,
		"filename":     storedName,
		"originalname": originalName,
		"markdown":     markdown,
	})
}
