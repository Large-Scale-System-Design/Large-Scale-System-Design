package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		writeError(w, http.StatusNotFound, "not_found", "not found")
		return
	}

	now := time.Now().UTC()

	var (
		orig    string
		expire  sql.NullTime
	)
	q := `
SELECT original_url, expire_at
FROM urls
WHERE short_code = ?
  AND deleted_at IS NULL
  AND (expire_at IS NULL OR expire_at > ?)
LIMIT 1
`
	err := h.DB.QueryRow(q, code, now).Scan(&orig, &expire)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "not_found", "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "db_error", "query failed")
		return
	}

	w.Header().Set("Location", orig)
	w.WriteHeader(http.StatusFound) // 302
}
