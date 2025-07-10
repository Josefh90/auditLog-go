package auditlog

import (
	"github.com/Josefh90/auditLog-go/pkg/auditlog/internal"

	"gorm.io/gorm"
)

// RegisterAuditCallbacks is the only exported function.
func RegisterAuditCallbacks(db *gorm.DB) {
	internal.RegisterAuditCallbacks(db)
}
