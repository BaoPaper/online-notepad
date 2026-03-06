package notepad

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreNoteLifecycle(t *testing.T) {
	app := newTestApp(t)
	store := app.store

	note, err := store.createNote("第一篇")
	if err != nil {
		t.Fatalf("createNote: %v", err)
	}
	if note.Title != "第一篇" {
		t.Fatalf("title = %q, want 第一篇", note.Title)
	}

	title, ok, err := store.renameNote(note.ID, "")
	if err != nil {
		t.Fatalf("renameNote: %v", err)
	}
	if !ok {
		t.Fatal("expected note to exist during rename")
	}
	if title != "无标题" {
		t.Fatalf("renamed title = %q, want 无标题", title)
	}

	attachment := "123e4567-e89b-12d3-a456-426614174000.png"
	attachmentPath := filepath.Join(store.uploadsDir, attachment)
	if err := os.WriteFile(attachmentPath, []byte("img"), 0o644); err != nil {
		t.Fatalf("write attachment: %v", err)
	}
	if err := store.addAttachment(note.ID, attachment); err != nil {
		t.Fatalf("addAttachment: %v", err)
	}

	if ok, err := store.saveNoteContent(note.ID, "![img](/uploads/"+attachment+")"); err != nil || !ok {
		t.Fatalf("saveNoteContent keep attachment: ok=%v err=%v", ok, err)
	}

	notes, err := store.listNotes()
	if err != nil {
		t.Fatalf("listNotes: %v", err)
	}
	if len(notes) != 1 || len(notes[0].Attachments) != 1 || notes[0].Attachments[0] != attachment {
		t.Fatalf("attachments after save = %#v", notes)
	}

	if ok, err := store.saveNoteContent(note.ID, "plain text"); err != nil || !ok {
		t.Fatalf("saveNoteContent remove attachment: ok=%v err=%v", ok, err)
	}
	if _, err := os.Stat(attachmentPath); !os.IsNotExist(err) {
		t.Fatalf("expected attachment to be removed, stat err = %v", err)
	}

	notes, err = store.listNotes()
	if err != nil {
		t.Fatalf("listNotes second time: %v", err)
	}
	if len(notes[0].Attachments) != 0 {
		t.Fatalf("attachments after cleanup = %#v", notes[0].Attachments)
	}

	deleted, err := store.deleteNote(note.ID)
	if err != nil {
		t.Fatalf("deleteNote: %v", err)
	}
	if !deleted {
		t.Fatal("expected note to be deleted")
	}
	if _, err := os.Stat(store.notePath(note.ID)); !os.IsNotExist(err) {
		t.Fatalf("expected note file to be removed, stat err = %v", err)
	}
}

func TestStoreMigratesLegacyNotes(t *testing.T) {
	baseDir := t.TempDir()
	store := &Store{
		notesDir:   filepath.Join(baseDir, "notes"),
		uploadsDir: filepath.Join(baseDir, "uploads"),
		metaFile:   filepath.Join(baseDir, "notes", "meta.json"),
	}
	if err := os.MkdirAll(store.notesDir, 0o755); err != nil {
		t.Fatalf("mkdir notes: %v", err)
	}
	if err := os.WriteFile(filepath.Join(store.notesDir, "1.txt"), []byte("legacy note"), 0o644); err != nil {
		t.Fatalf("write legacy note: %v", err)
	}

	if err := store.bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	notes, err := store.listNotes()
	if err != nil {
		t.Fatalf("listNotes: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("notes count = %d, want 1", len(notes))
	}
	if notes[0].Title != "笔记 1" {
		t.Fatalf("migrated title = %q, want 笔记 1", notes[0].Title)
	}
	if _, err := os.Stat(filepath.Join(store.notesDir, "1.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected legacy file removed, stat err = %v", err)
	}

	store.mu.Lock()
	meta, err := store.readMetaLocked()
	store.mu.Unlock()
	if err != nil {
		t.Fatalf("readMetaLocked: %v", err)
	}
	if !meta.Migrated {
		t.Fatal("expected migrated flag to be true")
	}
}
