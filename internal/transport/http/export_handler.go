package http

import (
	"archive/zip"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/libs-artifex/wrapper/v2"
)

// ExportHandler serves generated project files as a ZIP archive.
type ExportHandler struct{}

// NewExportHandler creates a handler for project export.
func NewExportHandler() *ExportHandler {
	return &ExportHandler{}
}

// HandleExport writes a session's generated files as a ZIP to the response.
// Files are read from the per-session fileStore (keyed by session_id), so
// concurrent generations stay isolated — no shared "last result" singleton.
// GET /api/v1/project/export?session_id=...
func (h *ExportHandler) HandleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET only")

		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "session_id required")

		return
	}

	files := globalFileStore.Get(sessionID)
	if len(files) == 0 {
		writeError(w, http.StatusNotFound, "No generated project available for this session. Run generation first.")

		return
	}

	// Set headers for ZIP download
	filename := fmt.Sprintf("istok-project-%s.zip", time.Now().Format("2006-01-02_150405"))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	// Write ZIP directly to response
	zw := zip.NewWriter(w)
	defer zw.Close()

	fileCount := 0
	ctx := r.Context()
	for name, content := range files {
		f, err := zw.Create(name)
		if err != nil {
			logFrom(ctx).WarnContext(ctx, "export zip create entry failed",
				"name", name,
				"error", wrapper.Wrap(err),
			)

			continue
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			logFrom(ctx).WarnContext(ctx, "export zip write entry failed",
				"name", name,
				"error", wrapper.Wrap(err),
			)

			continue
		}
		fileCount++
	}
	logFrom(ctx).InfoContext(ctx, "export zip complete",
		"fileCount", fileCount,
		"filename", filename,
		"sessionId", sessionID,
	)
}
