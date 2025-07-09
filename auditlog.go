package auditlog

import (
	"encoding/json"
	"reflect"
	"time"

	"gorm.io/datatypes"
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
    db.Callback().Create().
        After("gorm:create").
        Register("audit:create", UnifiedAuditCallback("CREATE"))

    // Register a callback to run after every UPDATE operation.
    // It uses the UnifiedAuditCallback with the action "UPDATE".
    db.Callback().Update().
        After("gorm:update").
        Register("audit:update", UnifiedAuditCallback("UPDATE"))

    // Register a callback to run after every DELETE operation.
    // It uses the UnifiedAuditCallback with the action "DELETE".
    db.Callback().Delete().
        After("gorm:delete").
        Register("audit:delete", UnifiedAuditCallback("DELETE"))
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
		// Abort early if the DB context is invalid or already errored
		if db.Error != nil || db.Statement == nil || db.Statement.Schema == nil {
			return
		}

		// Prevent logging audit entries on the AuditLog table itself to avoid infinite recursion
		if db.Statement.Schema.Name == "AuditLog" {
			return
		}

		var oldData, newData []byte
		var err error

		// -----------------------------------
		// Step 1: Fetch old data for UPDATE and DELETE operations
		// -----------------------------------
		if action == "UPDATE" || action == "DELETE" {
			// Create a new empty instance of the model type to hold the existing record state
			oldRecord := reflect.New(db.Statement.Schema.ModelType).Interface()

			// Build a query condition map using primary key fields and their values
			primaryKeys := map[string]interface{}{}
			for _, field := range db.Statement.Schema.PrimaryFields {
				// Extract the primary key value from the current statement's model instance
				if value, isZero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue); !isZero {
					primaryKeys[field.DBName] = value
				}
			}

			if len(primaryKeys) > 0 {
				// Use a new DB session to avoid interfering with the current transaction or callbacks
				_ = db.Session(&gorm.Session{NewDB: true}).
					Model(db.Statement.Model).
					Where(primaryKeys).
					First(oldRecord).Error

				// Serialize the old record state into JSON for logging
				oldData, _ = json.Marshal(oldRecord)
			}
		}

		// -----------------------------------
		// Step 2: Serialize new data for CREATE and UPDATE operations
		// -----------------------------------
		if action == "CREATE" || action == "UPDATE" {
			newData, err = json.Marshal(db.Statement.Dest)
			if err != nil {
				// Silently ignore serialization errors to avoid affecting main DB operation
				newData = nil
			}
		}

		// -----------------------------------
		// Step 3: Construct the audit log entry struct
		// -----------------------------------
		audit := AuditLog{
			TableName: db.Statement.Schema.Table,   // Name of the affected table
			Action:    action,                      // "CREATE", "UPDATE", or "DELETE"
			OldData:   datatypes.JSON(oldData),    // JSON of old record state (if any)
			NewData:   datatypes.JSON(newData),    // JSON of new record state (if any)
			CreatedAt: time.Now(),                  // Timestamp of the operation
		}

		// -----------------------------------
		// Step 4: Insert the audit log entry into the audit_logs table
		// Use a fresh DB session to prevent recursive triggers of this callback
		// -----------------------------------
		_ = db.Session(&gorm.Session{NewDB: true}).Model(&AuditLog{}).Create(&audit)
	}
}
