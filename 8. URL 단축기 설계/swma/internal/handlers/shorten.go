package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"example.com/urlshortener/internal/base62"
	"example.com/urlshortener/internal/idgen"
)

type shortenReq struct {
	OriginalURL string  `json:"originalUrl"`
	ExpireAt    *string `json:"expireAt,omitempty"` // RFC3339
}

type shortenResp struct {
	ID          uint64     `json:"id"`
	Code        string     `json:"code"`
	ShortURL    string     `json:"shortUrl"`
	OriginalURL string     `json:"originalUrl"`
	ExpireAt    *time.Time `json:"expireAt,omitempty"`
	Reused      bool       `json:"reused"`
}

var generator = idgen.NewGenerator(0, 0)

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	orig := strings.TrimSpace(req.OriginalURL)
	if orig == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "originalUrl is required")
		return
	}
	if len(orig) > 2048 {
		writeError(w, http.StatusBadRequest, "bad_request", "originalUrl too long (max 2048)")
		return
	}

	u, err := url.Parse(orig)
	if err != nil || u.Scheme == "" || u.Host == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "originalUrl must be an absolute URL")
		return
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		writeError(w, http.StatusBadRequest, "bad_request", "only http/https URLs are allowed")
		return
	}

	var expireAt *time.Time
	if req.ExpireAt != nil && strings.TrimSpace(*req.ExpireAt) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ExpireAt))
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "expireAt must be RFC3339 (e.g. 2030-01-01T00:00:00Z)")
			return
		}
		tt := t.UTC()
		expireAt = &tt
	}

	sha := sha256Bytes(orig)
	now := nowUTC()

	// Reuse if active row exists (not deleted, not expired)
	var (
		id       uint64
		code     string
		dbExpire sql.NullTime
		dbOrig   string
	)
	q := `
SELECT id, short_code, original_url, expire_at
FROM urls
WHERE original_url_sha256 = ?
  AND deleted_at IS NULL
  AND (expire_at IS NULL OR expire_at > ?)
ORDER BY id DESC
LIMIT 1
`
	err = h.DB.QueryRow(q, sha[:], now).Scan(&id, &code, &dbOrig, &dbExpire)
	if err == nil {
		// prevent rare hash collision returning wrong url
		if dbOrig == orig {
			resp := shortenResp{
				ID:          id,
				Code:        code,
				ShortURL:    h.Cfg.BaseURL + "/s/" + code,
				OriginalURL: dbOrig,
				Reused:      true,
			}
			if dbExpire.Valid {
				t := dbExpire.Time.UTC()
				resp.ExpireAt = &t
			}
			writeJSON(w, http.StatusOK, resp)
			return
		}
	}
	if err != nil && err != sql.ErrNoRows {
		writeError(w, http.StatusInternalServerError, "db_error", "query failed")
		return
	}

	// Create a new record
	newID := uint64(generator.Next())
	newCode := base62.Encode(newID)

	ins := `
INSERT INTO urls (id, short_code, original_url, original_url_sha256, expire_at)
VALUES (?, ?, ?, ?, ?)
`
	_, err = h.DB.Exec(ins, newID, newCode, orig, sha[:], expireAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "insert failed")
		return
	}

	resp := shortenResp{
		ID:          newID,
		Code:        newCode,
		ShortURL:    h.Cfg.BaseURL + "/s/" + newCode,
		OriginalURL: orig,
		ExpireAt:    expireAt,
		Reused:      false,
	}
	writeJSON(w, http.StatusCreated, resp)
}
