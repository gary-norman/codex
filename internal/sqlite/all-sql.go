package sqlite

import (
	"database/sql"

	"github.com/gary-norman/forum/internal/colors"
)

type AllModel struct {
	DB *sql.DB
}

var (
	Colors, _ = colors.UseFlavor("Mocha")
)
