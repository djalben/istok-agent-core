package http

import (
	"archive/zip"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/istok/agent-core/internal/application"
)

// ExportHandler serves generated project files as a ZIP archive.
type ExportHandler struct {
	orchestrator *application.Orchestrator
}

// NewExportHandler creates a handler for project export.
func NewExportHandler(orchestrator *application.Orchestrator) *ExportHandler {
	return &ExportHandler{orchestrator: orchestrator}
}

// HandleExport writes all generated files as a ZIP to the response.
// GET /api/v1/project/export
func (h *ExportHandler) HandleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	result := h.orchestrator.GetLastResult()
	if result == nil || len(result.Code) == 0 {
		writeError(w, http.StatusNotFound, "No generated project available. Run generation first.")
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
	for name, content := range result.Code {
		f, err := zw.Create(name)
		if err != nil {
			log.Printf("⚠️ Export ZIP: failed to create entry %q: %v", name, err)
			continue
		}
		if _, err := f.Write([]byte(content)); err != nil {
			log.Printf("⚠️ Export ZIP: failed to write %q: %v", name, err)
			continue
		}
		fileCount++
	}

	log.Printf("📦 Export: %d files packed into %s", fileCount, filename)
}
