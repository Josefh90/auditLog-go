package internal

import (
	"encoding/json"
	"reflect"
	"time"

	"gorm.io/gorm"
)

// RegisterAuditCallbacks registers the audit logging callbacks for the GORM DB instance.
// It sets up hooks to capture CREATE, UPDATE, and DELETE operations and logs them via
// the UnifiedAuditCallback function, passing the corresponding action string.
//
// Parameters:
// - db: the *gorm.DB instance to register the callbacks on.
//
// This centralizes audit callback registration, making it easy to enable auditing with a single call.
func RegisterAuditCallbacks(db *gorm.DB) {

	// Register a callback to run after every CREATE operation.
	// It uses the UnifiedAuditCallback with the action "CREATE".
	if err := db.Callback().Create().After("gorm:create").Register("audit:create",
		AuditCallback("CREATE")); err != nil {
		panic("audit:create: " + err.Error())
	}

	// Register a callback to run after every UPDATE operation.
	// It uses the UnifiedAuditCallback with the action "UPDATE"
	if err := db.Callback().Update().After("gorm:update").Register("audit:update",
		AuditCallback("UPDATE")); err != nil {
		panic("audit:update: " + err.Error())
	}

	// Register a callback to run after every DELETE operation.
	// It uses the UnifiedAuditCallback with the action "DELETE".
	if err := db.Callback().Delete().After("gorm:delete").Register("audit:delete",
		AuditCallback("DELETE")); err != nil {
		panic("audit:delete: " + err.Error())
	}
}

// Deprecated: use AuditCallback instead.
// UnifiedAuditCallback is kept for backward compatibility and will be removed in a future release.
func UnifiedAuditCallback(action string) func(db *gorm.DB) {
	return AuditCallback(action)
}

// UnifiedAuditCallback returns a GORM callback function that logs audit entries for
// the specified database operation (CREATE, UPDATE, DELETE). It captures both the old
// and new state of the record as JSON, enabling full change tracking.
//
// Parameters:
// - action: a string indicating the operation type ("CREATE", "UPDATE", or "DELETE").
//
// This design allows a single callback function to be reused for all three operations by
// passing the appropriate action during registration.
func AuditCallback(action string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		// Return early if there is an error in the DB context or if the statement/schema is missing
		if db.Error != nil || db.Statement == nil || db.Statement.Schema == nil {
			return
		}

		// Avoid logging audit entries for the AuditLog table itself to prevent infinite recursion
		if db.Statement.Schema.Name == "AuditLog" {
			return
		}

		var oldData, newData []byte

		// For UPDATE and DELETE operations, fetch the old state of the record
		if action == "UPDATE" || action == "DELETE" {
			// Create a new instance of the model type to hold the old record's data
			oldRecord := reflect.New(db.Statement.Schema.ModelType).Interface()

			// Build a map of primary key fields and their values from the current record
			pk := map[string]interface{}{}
			for _, field := range db.Statement.Schema.PrimaryFields {
				if val, zero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue); !zero {
					pk[field.DBName] = val
				}
			}

			// If primary keys exist, query the old record from the DB using a fresh session
			if len(pk) > 0 {
				_ = db.Session(&gorm.Session{NewDB: true}).
					Model(db.Statement.Model).
					Where(pk).
					First(oldRecord).Error

				// Marshal the old record into JSON bytes for logging
				oldData, _ = json.Marshal(oldRecord)
			}
		}

		// For CREATE and UPDATE operations, marshal the new state of the record
		if action == "CREATE" || action == "UPDATE" {
			newData, _ = json.Marshal(db.Statement.Dest)
		}

		// Create a new AuditLog entry with the captured information and insert it using a fresh DB session
		_ = db.Session(&gorm.Session{NewDB: true}).Create(&AuditLog{
			TableName: db.Statement.Schema.Table, // Name of the affected table
			Action:    action,                    // The action type: CREATE, UPDATE, DELETE
			OldData:   oldData,                   // JSON of old record state (may be nil)
			NewData:   newData,                   // JSON of new record state (may be nil)
			CreatedAt: time.Now(),                // Timestamp of when this audit entry was created
		})
	}
}
