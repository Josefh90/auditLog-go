package internal

import (
	"encoding/json"
	"reflect"
	"time"

	"gorm.io/gorm"
)

func RegisterAuditCallbacks(db *gorm.DB) {
	if err := db.Callback().Create().After("gorm:create").Register("audit:create",
		auditCallback("CREATE")); err != nil {
		panic("audit:create: " + err.Error())
	}
	if err := db.Callback().Update().After("gorm:update").Register("audit:update",
		auditCallback("UPDATE")); err != nil {
		panic("audit:update: " + err.Error())
	}
	if err := db.Callback().Delete().After("gorm:delete").Register("audit:delete",
		auditCallback("DELETE")); err != nil {
		panic("audit:delete: " + err.Error())
	}
}

func auditCallback(action string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if db.Error != nil || db.Statement == nil || db.Statement.Schema == nil {
			return
		}
		if db.Statement.Schema.Name == "AuditLog" {
			return
		}

		var oldData, newData []byte

		if action == "UPDATE" || action == "DELETE" {
			oldRecord := reflect.New(db.Statement.Schema.ModelType).Interface()
			pk := map[string]interface{}{}
			for _, field := range db.Statement.Schema.PrimaryFields {
				if val, zero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue); !zero {
					pk[field.DBName] = val
				}
			}
			if len(pk) > 0 {
				_ = db.Session(&gorm.Session{NewDB: true}).
					Model(db.Statement.Model).
					Where(pk).First(oldRecord).Error
				oldData, _ = json.Marshal(oldRecord)
			}
		}

		if action == "CREATE" || action == "UPDATE" {
			newData, _ = json.Marshal(db.Statement.Dest)
		}

		_ = db.Session(&gorm.Session{NewDB: true}).Create(&AuditLog{
			TableName: db.Statement.Schema.Table,
			Action:    action,
			OldData:   oldData,
			NewData:   newData,
			CreatedAt: time.Now(),
		})
	}
}
