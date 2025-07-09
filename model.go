package auditlog

import (
	"time"

	"gorm.io/datatypes"
)

// AuditLog represents a single entry in the audit log.
// It captures the changes made to database records for auditing purposes.
type AuditLog struct {
	ID uint `gorm:"primaryKey"` // Unique identifier for the audit log entry

	TableName string `gorm:"index"` // Name of the database table where the change occurred

	Action string // Type of action performed (e.g., CREATE, UPDATE, DELETE)

	NewData datatypes.JSON // JSON representation of the new data (after the change)

	OldData datatypes.JSON // JSON representation of the old data (before the change)

	CreatedAt time.Time // Timestamp when the audit log entry was created
}
