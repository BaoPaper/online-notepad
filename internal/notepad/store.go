package notepad

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (s *Store) bootstrap() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureNotesDirLocked(); err != nil {
		return err
	}
	if err := os.MkdirAll(s.uploadsDir, 0o755); err != nil {
		return err
	}

	return s.migrateOldNotesLocked()
}

func (s *Store) ensureNotesDirLocked() error {
	return os.MkdirAll(s.notesDir, 0o755)
}

func (s *Store) writeNoteLocked(id string, content []byte) error {
	if err := s.ensureNotesDirLocked(); err != nil {
		return err
	}
	return os.WriteFile(s.notePath(id), content, 0o644)
}

func (s *Store) ensureLandingNote() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, err := s.readMetaLocked()
	if err != nil {
		return "", err
	}
	if len(meta.Notes) > 0 {
		sorted := sortNotesByUpdatedAt(meta.Notes)
		return sorted[0].ID, nil
	}

	now := time.Now().UnixMilli()
	note := Note{
		ID:        uuidString(),
		Title:     "新建笔记",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.writeNoteLocked(note.ID, []byte("")); err != nil {
		return "", err
	}

	meta.Notes = append(meta.Notes, note)
	if err := s.writeMetaLocked(meta); err != nil {
		return "", err
	}

	return note.ID, nil
}

func (s *Store) getNoteContent(id string) (Note, string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, err := s.readMetaLocked()
	if err != nil {
		return Note{}, "", false, err
	}

	index := findNoteIndex(meta.Notes, id)
	if index < 0 {
		return Note{}, "", false, nil
	}

	content, err := s.readNoteContentLocked(id)
	if err != nil {
		return Note{}, "", false, err
	}

	return meta.Notes[index], content, true, nil
}

func (s *Store) listNotes() ([]Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, err := s.readMetaLocked()
	if err != nil {
		return nil, err
	}

	return sortNotesByUpdatedAt(meta.Notes), nil
}

func (s *Store) createNote(title string) (Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, err := s.readMetaLocked()
	if err != nil {
		return Note{}, err
	}

	now := time.Now().UnixMilli()
	note := Note{
		ID:        uuidString(),
		Title:     defaultString(title, "新建笔记"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.writeNoteLocked(note.ID, []byte("")); err != nil {
		return Note{}, err
	}

	meta.Notes = append(meta.Notes, note)
	if err := s.writeMetaLocked(meta); err != nil {
		return Note{}, err
	}

	return note, nil
}

func (s *Store) deleteNote(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, err := s.readMetaLocked()
	if err != nil {
		return false, err
	}

	index := findNoteIndex(meta.Notes, id)
	if index < 0 {
		return false, nil
	}

	note := meta.Notes[index]
	for _, filename := range note.Attachments {
		if err := s.deleteAttachmentLocked(filename); err != nil {
			return false, err
		}
	}

	meta.Notes = append(meta.Notes[:index], meta.Notes[index+1:]...)
	if err := s.writeMetaLocked(meta); err != nil {
		return false, err
	}

	if err := os.Remove(s.notePath(id)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	return true, nil
}

func (s *Store) renameNote(id, title string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, err := s.readMetaLocked()
	if err != nil {
		return "", false, err
	}

	index := findNoteIndex(meta.Notes, id)
	if index < 0 {
		return "", false, nil
	}

	meta.Notes[index].Title = defaultString(title, "无标题")
	markNoteAsEdited(&meta.Notes[index])
	if err := s.writeMetaLocked(meta); err != nil {
		return "", false, err
	}

	return meta.Notes[index].Title, true, nil
}

func (s *Store) saveNoteContent(id, content string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, err := s.readMetaLocked()
	if err != nil {
		return false, err
	}

	index := findNoteIndex(meta.Notes, id)
	if index < 0 {
		return false, nil
	}

	if err := s.writeNoteLocked(id, []byte(content)); err != nil {
		return false, err
	}

	currentAttachments := extractAttachments(content)
	oldAttachments := meta.Notes[index].Attachments
	for _, filename := range oldAttachments {
		if !containsString(currentAttachments, filename) {
			if err := s.deleteAttachmentLocked(filename); err != nil {
				return false, err
			}
		}
	}

	meta.Notes[index].Attachments = currentAttachments
	markNoteAsEdited(&meta.Notes[index])
	if err := s.writeMetaLocked(meta); err != nil {
		return false, err
	}

	return true, nil
}

func (s *Store) addAttachment(noteID, filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, err := s.readMetaLocked()
	if err != nil {
		return err
	}

	index := findNoteIndex(meta.Notes, noteID)
	if index < 0 {
		return nil
	}

	meta.Notes[index].Attachments = append(meta.Notes[index].Attachments, filename)
	return s.writeMetaLocked(meta)
}

func (s *Store) migrateOldNotesLocked() error {
	meta, err := s.readMetaLocked()
	if err != nil {
		return err
	}
	if meta.Migrated {
		return nil
	}

	baseTime := time.Now().UnixMilli()
	for i := 1; i <= 8; i++ {
		oldPath := filepath.Join(s.notesDir, fmt.Sprintf("%d.txt", i))
		content, err := os.ReadFile(oldPath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}

		if strings.TrimSpace(string(content)) != "" {
			id := uuidString()
			if err := s.writeNoteLocked(id, content); err != nil {
				return err
			}

			timestamp := baseTime - int64(i*1000)
			meta.Notes = append(meta.Notes, Note{
				ID:        id,
				Title:     fmt.Sprintf("笔记 %d", i),
				CreatedAt: timestamp,
				UpdatedAt: timestamp,
			})
		}

		if err := os.Remove(oldPath); err != nil {
			return err
		}
	}

	meta.Migrated = true
	return s.writeMetaLocked(meta)
}

func (s *Store) readMetaLocked() (Meta, error) {
	data, err := os.ReadFile(s.metaFile)
	if errors.Is(err, os.ErrNotExist) {
		return Meta{Notes: []Note{}}, nil
	}
	if err != nil {
		return Meta{}, err
	}

	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return Meta{}, err
	}
	if meta.Notes == nil {
		meta.Notes = []Note{}
	}

	return meta, nil
}

func (s *Store) writeMetaLocked(meta Meta) error {
	if meta.Notes == nil {
		meta.Notes = []Note{}
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := s.ensureNotesDirLocked(); err != nil {
		return err
	}
	return os.WriteFile(s.metaFile, data, 0o644)
}

func (s *Store) readNoteContentLocked(id string) (string, error) {
	data, err := os.ReadFile(s.notePath(id))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *Store) deleteAttachmentLocked(filename string) error {
	name := filepath.Base(filename)
	if name == "." || name == "" {
		return nil
	}

	err := os.Remove(filepath.Join(s.uploadsDir, name))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *Store) notePath(id string) string {
	return filepath.Join(s.notesDir, id+".txt")
}

func effectiveUpdatedAt(note Note) int64 {
	if note.UpdatedAt != 0 {
		return note.UpdatedAt
	}
	return note.CreatedAt
}

func sortNotesByUpdatedAt(notes []Note) []Note {
	sorted := append([]Note(nil), notes...)
	sort.Slice(sorted, func(i, j int) bool {
		leftUpdated := effectiveUpdatedAt(sorted[i])
		rightUpdated := effectiveUpdatedAt(sorted[j])
		if leftUpdated != rightUpdated {
			return leftUpdated > rightUpdated
		}
		return sorted[i].CreatedAt > sorted[j].CreatedAt
	})
	return sorted
}

func markNoteAsEdited(note *Note) {
	note.UpdatedAt = time.Now().UnixMilli()
}

func extractAttachments(content string) []string {
	matches := attachmentPattern.FindAllStringSubmatch(content, -1)
	attachments := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			attachments = append(attachments, match[1])
		}
	}
	return attachments
}

func findNoteIndex(notes []Note, id string) int {
	for index, note := range notes {
		if note.ID == id {
			return index
		}
	}
	return -1
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
