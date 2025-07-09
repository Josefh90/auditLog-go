package auditlog

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CreateCallback is a GORM callback that automatically logs
// CREATE operations into the AuditLog table.
//
// It serializes the newly created object to JSON and stores it
// along with metadata such as the table name and timestamp.
func CreateCallback(db *gorm.DB) {
	// Abort early if there was an error or if the DB context is not properly set up
	if db.Error != nil || db.Statement == nil || db.Statement.Schema == nil {
		return
	}

	// Prevent recursion: do not log operations on the AuditLog table itself
	if db.Statement.Schema.Name == "AuditLog" {
		return
	}

	// Serialize the newly inserted record (db.Statement.Dest) to JSON
	newData, err := json.Marshal(db.Statement.Dest)
	if err != nil {
		return // Silently fail on serialization errors
	}

	// Create a new AuditLog entry
	log := AuditLog{
		TableName: db.Statement.Schema.Table,  // Name of the table where the operation occurred
		Action:    "CREATE",                   // Action type (CREATE in this case)
		NewData:   datatypes.JSON(newData),   // Serialized new record
		CreatedAt: time.Now(),                // Timestamp of the operation
	}

	// Save the audit log entry using a fresh DB session to avoid interfering with the current transaction
	db.Session(&gorm.Session{
		NewDB: true, // Important to avoid recursive callback triggering
	}).Model(&AuditLog{}).Create(&log)
}
