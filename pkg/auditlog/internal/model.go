package internal

import (
	"time"
)

// AuditLog represents a single entry in the audit log.
// It captures the changes made to database records for auditing purposes.
type AuditLog struct {
	ID uint `gorm:"primaryKey"` // Unique identifier for the audit log entry

	TableName string `gorm:"index"` // Name of the database table where the change occurred

	Action string // Type of action performed (e.g., CREATE, UPDATE, DELETE)

	NewData []byte `gorm:"type:jsonb"`

	OldData []byte `gorm:"type:jsonb"`

	CreatedAt time.Time // Timestamp when the audit log entry was created
}
