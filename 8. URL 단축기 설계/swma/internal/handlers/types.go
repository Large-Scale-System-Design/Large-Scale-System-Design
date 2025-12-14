package handlers

import (
	"database/sql"

	"example.com/urlshortener/internal/config"
)

type Handler struct {
	Cfg config.Config
	DB  *sql.DB
}
