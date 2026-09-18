package filemanager

import (
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Manager struct {
	root   string
	logger *slog.Logger
}

type FileEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	IsDir   bool   `json:"is_dir"`
	ModTime string `json:"mod_time"`
}

func New(root string, logger *slog.Logger) *Manager {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}
	os.MkdirAll(absRoot, 0755)
	return &Manager{root: absRoot, logger: logger}
}

func (m *Manager) safePath(relPath string) (string, error) {
	cleaned := filepath.Clean("/" + relPath)
	cleaned = strings.TrimPrefix(cleaned, "/")

	full := filepath.Join(m.root, cleaned)
	abs, err := filepath.Abs(full)
	if err != nil {
		return "", fmt.Errorf("invalid path")
	}

	rel, err := filepath.Rel(m.root, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path traversal detected")
	}

	return abs, nil
}

func (m *Manager) List(dirPath string) ([]FileEntry, error) {
	safePath, err := m.safePath(dirPath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(safePath)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	files := make([]FileEntry, 0)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		relPath, _ := filepath.Rel(m.root, filepath.Join(safePath, entry.Name()))
		files = append(files, FileEntry{
			Name:    entry.Name(),
			Path:    "/" + relPath,
			Size:    info.Size(),
			Mode:    info.Mode().String(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}
	return files, nil
}

func (m *Manager) Read(relPath string) ([]byte, error) {
	safePath, err := m.safePath(relPath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(safePath)
}

func (m *Manager) Write(relPath string, data []byte) error {
	safePath, err := m.safePath(relPath)
	if err != nil {
		return err
	}
	dir := filepath.Dir(safePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(safePath, data, 0644)
}

func (m *Manager) Mkdir(relPath string) error {
	safePath, err := m.safePath(relPath)
	if err != nil {
		return err
	}
	return os.MkdirAll(safePath, 0755)
}

func (m *Manager) Delete(relPath string) error {
	safePath, err := m.safePath(relPath)
	if err != nil {
		return err
	}
	return os.RemoveAll(safePath)
}

func (m *Manager) Rename(oldPath, newPath string) error {
	oldSafe, err := m.safePath(oldPath)
	if err != nil {
		return err
	}
	newSafe, err := m.safePath(newPath)
	if err != nil {
		return err
	}
	return os.Rename(oldSafe, newSafe)
}

func (m *Manager) ServeDownload(w http.ResponseWriter, r *http.Request, relPath string) error {
	safePath, err := m.safePath(relPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(safePath)
	if err != nil {
		return fmt.Errorf("file not found")
	}
	if info.IsDir() {
		return fmt.Errorf("cannot download directory")
	}

	ext := filepath.Ext(safePath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(safePath)))

	f, err := os.Open(safePath)
	if err != nil {
		return err
	}
	defer f.Close()

	io.Copy(w, f)
	return nil
}

func (m *Manager) ServeRead(w http.ResponseWriter, r *http.Request, relPath string) error {
	safePath, err := m.safePath(relPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(safePath)
	if err != nil {
		return fmt.Errorf("file not found")
	}
	if info.IsDir() {
		return fmt.Errorf("cannot read directory as file")
	}
	if info.Size() > 10*1024*1024 {
		return fmt.Errorf("file too large")
	}

	ext := filepath.Ext(safePath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", mimeType)
	f, err := os.Open(safePath)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(w, f)
	return nil
}

func (m *Manager) Upload(r *http.Request, relDir string) ([]string, error) {
	safeDir, err := m.safePath(relDir)
	if err != nil {
		return nil, err
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		return nil, fmt.Errorf("parse form: %w", err)
	}

	var uploaded []string
	for _, fileHeaders := range r.MultipartForm.File {
		for _, fh := range fileHeaders {
			src, err := fh.Open()
			if err != nil {
				continue
			}

			name := filepath.Base(fh.Filename)
			if strings.Contains(name, "/") || strings.Contains(name, "\\") {
				src.Close()
				continue
			}

			destPath := filepath.Join(safeDir, name)

			rel, err := filepath.Rel(m.root, destPath)
			if err != nil || strings.HasPrefix(rel, "..") {
				src.Close()
				continue
			}

			dst, err := os.Create(destPath)
			if err != nil {
				src.Close()
				continue
			}
			io.Copy(dst, src)
			dst.Close()
			src.Close()
			uploaded = append(uploaded, name)
		}
	}

	if len(uploaded) == 0 {
		return nil, fmt.Errorf("no files uploaded")
	}
	return uploaded, nil
}

func ParseUploadDir(r *http.Request) string {
	_ = multipart.NewReader
	dir := r.FormValue("path")
	if dir == "" {
		dir = "/"
	}
	return dir
}
