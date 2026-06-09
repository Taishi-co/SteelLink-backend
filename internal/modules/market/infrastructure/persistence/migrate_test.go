package persistence

import (
	"reflect"
	"testing"

	marketentity "github.com/Taishi-co/SteelLink-backend/internal/modules/market/domain/entity"
)

func TestMigrationModelsCreateProcurementProjectBeforeLogs(t *testing.T) {
	models := migrationModels()

	projectIndex := migrationModelIndex(models, &marketentity.ProcurementProject{})
	logIndex := migrationModelIndex(models, &marketentity.ProcurementLog{})
	if projectIndex == -1 {
		t.Fatal("ProcurementProject is missing from migration models")
	}
	if logIndex == -1 {
		t.Fatal("ProcurementLog is missing from migration models")
	}
	if projectIndex > logIndex {
		t.Fatalf("ProcurementProject migration index = %d, want before ProcurementLog index %d", projectIndex, logIndex)
	}
}

func migrationModelIndex(models []interface{}, target interface{}) int {
	targetType := reflect.TypeOf(target)
	for i, model := range models {
		if reflect.TypeOf(model) == targetType {
			return i
		}
	}
	return -1
}
