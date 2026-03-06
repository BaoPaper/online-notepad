package notepad

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotesAPIRequiresAuth(t *testing.T) {
	app := newTestApp(t)
	router := newTestRouter(t, app)

	req, err := http.NewRequest(http.MethodGet, "/api/notes", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	w := performRequest(router, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(w.Body.String(), "Unauthorized") {
		t.Fatalf("body = %q, want Unauthorized", w.Body.String())
	}
}

func TestLoginSetsCookieAndRedirects(t *testing.T) {
	app := newTestApp(t)
	router := newTestRouter(t, app)

	body := url.Values{
		"password":   {"password"},
		"rememberMe": {"on"},
	}.Encode()
	req, err := http.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := performRequest(router, req)
	if w.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusFound)
	}
	if location := w.Header().Get("Location"); location != "/" {
		t.Fatalf("location = %q, want /", location)
	}
	if !strings.Contains(w.Header().Get("Set-Cookie"), authCookieName+"=") {
		t.Fatalf("Set-Cookie = %q, want auth cookie", w.Header().Get("Set-Cookie"))
	}
}

func TestLoginPageUsesChineseDisplayName(t *testing.T) {
	app := newTestApp(t)
	router := newTestRouter(t, app)

	req, err := http.NewRequest(http.MethodGet, "/login", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Accept", "text/html")

	w := performRequest(router, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "<title>简单笔记 - 登录</title>") {
		t.Fatalf("body = %q, want Chinese display name in title", w.Body.String())
	}
}

func TestAuthenticatedNoteWorkflow(t *testing.T) {
	app := newTestApp(t)
	router := newTestRouter(t, app)
	cookie := authCookie(t, app)

	createReq, err := http.NewRequest(http.MethodPost, "/api/notes", strings.NewReader(`{"title":"测试笔记"}`))
	if err != nil {
		t.Fatalf("new create request: %v", err)
	}
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(cookie)

	createResp := performRequest(router, createReq)
	if createResp.Code != http.StatusOK {
		t.Fatalf("create status = %d, want %d", createResp.Code, http.StatusOK)
	}

	var created Note
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created note: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected created note id")
	}

	saveReq, err := http.NewRequest(http.MethodPost, "/save/"+created.ID, strings.NewReader(url.Values{"note": {"hello markdown"}}.Encode()))
	if err != nil {
		t.Fatalf("new save request: %v", err)
	}
	saveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	saveReq.AddCookie(cookie)

	saveResp := performRequest(router, saveReq)
	if saveResp.Code != http.StatusOK {
		t.Fatalf("save status = %d, want %d", saveResp.Code, http.StatusOK)
	}

	contentReq, err := http.NewRequest(http.MethodGet, "/api/notes/"+created.ID+"/content", nil)
	if err != nil {
		t.Fatalf("new content request: %v", err)
	}
	contentReq.Header.Set("Accept", "application/json")
	contentReq.AddCookie(cookie)

	contentResp := performRequest(router, contentReq)
	if contentResp.Code != http.StatusOK {
		t.Fatalf("content status = %d, want %d", contentResp.Code, http.StatusOK)
	}

	var payload struct {
		Success bool   `json:"success"`
		Content string `json:"content"`
		Title   string `json:"title"`
	}
	if err := json.Unmarshal(contentResp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal content payload: %v", err)
	}
	if !payload.Success || payload.Content != "hello markdown" || payload.Title != "测试笔记" {
		t.Fatalf("unexpected content payload: %+v", payload)
	}

	pageReq, err := http.NewRequest(http.MethodGet, "/note/"+created.ID, nil)
	if err != nil {
		t.Fatalf("new note page request: %v", err)
	}
	pageReq.Header.Set("Accept", "text/html")
	pageReq.AddCookie(cookie)

	pageResp := performRequest(router, pageReq)
	if pageResp.Code != http.StatusOK {
		t.Fatalf("note page status = %d, want %d", pageResp.Code, http.StatusOK)
	}
	if !strings.Contains(pageResp.Body.String(), "<title>测试笔记 - 简单笔记</title>") {
		t.Fatalf("body = %q, want Chinese display name in note title", pageResp.Body.String())
	}
}

func TestUploadAPIStoresFileAndReturnsMarkdown(t *testing.T) {
	app := newTestApp(t)
	router := newTestRouter(t, app)
	note, err := app.store.createNote("上传测试")
	if err != nil {
		t.Fatalf("createNote: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("noteId", note.ID); err != nil {
		t.Fatalf("write noteId: %v", err)
	}
	part, err := writer.CreateFormFile("file", "demo.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(part, strings.NewReader("demo content")); err != nil {
		t.Fatalf("copy file content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/api/upload", &body)
	if err != nil {
		t.Fatalf("new upload request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(authCookie(t, app))

	resp := performRequest(router, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want %d", resp.Code, http.StatusOK)
	}

	var payload struct {
		Success      bool   `json:"success"`
		Filename     string `json:"filename"`
		OriginalName string `json:"originalname"`
		Markdown     string `json:"markdown"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal upload payload: %v", err)
	}
	if !payload.Success || payload.OriginalName != "demo.txt" {
		t.Fatalf("unexpected upload payload: %+v", payload)
	}
	if !strings.HasPrefix(payload.Markdown, "[demo.txt](/uploads/") {
		t.Fatalf("markdown = %q, want upload link", payload.Markdown)
	}
	if _, err := os.Stat(filepath.Join(app.store.uploadsDir, payload.Filename)); err != nil {
		t.Fatalf("uploaded file missing: %v", err)
	}
}
