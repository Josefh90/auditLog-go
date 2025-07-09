package auditlog

import (
	"time"

	"gorm.io/datatypes"
)

type AuditLog struct {
	ID        uint           `gorm:"primaryKey"`
	TableName string         `gorm:"index"`
	Action    string
	NewData   datatypes.JSON
	OldData   datatypes.JSON
	CreatedAt time.Time
}
