package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		writeError(w, http.StatusNotFound, "not_found", "not found")
		return
	}

	now := time.Now().UTC()
	q := `
UPDATE urls
SET deleted_at = ?
WHERE short_code = ?
  AND deleted_at IS NULL
`
	res, err := h.DB.Exec(q, now, code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "update failed")
		return
	}
	n, err := res.RowsAffected()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "update failed")
		return
	}
	if n == 0 {
		// if it exists but expired, still treat as not found per your rule
		var dummy uint64
		err2 := h.DB.QueryRow("SELECT id FROM urls WHERE short_code = ? LIMIT 1", code).Scan(&dummy)
		if err2 == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "not_found", "not found")
			return
		}
		writeError(w, http.StatusNotFound, "not_found", "not found")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204
}
