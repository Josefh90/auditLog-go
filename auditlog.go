package auditlog

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func CreateCallback(db *gorm.DB) {
	if db.Error != nil || db.Statement == nil || db.Statement.Schema == nil {
		return
	}

	if db.Statement.Schema.Name == "AuditLog" {
		return
	}

	newData, err := json.Marshal(db.Statement.Dest)
	if err != nil {
		return
	}

	log := AuditLog{
		TableName: db.Statement.Schema.Table,
		Action:    "CREATE",
		NewData:   datatypes.JSON(newData),
		CreatedAt: time.Now(),
	}

	db.Session(&gorm.Session{
		NewDB: true,
	}).Model(&AuditLog{}).Create(&log)
}
