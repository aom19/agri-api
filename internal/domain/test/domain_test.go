package domain_test

import (
	"testing"

	"agri-api/internal/domain"
)

func TestEnumValidation(t *testing.T) {
	if !domain.MachineTypeTractor.IsValid() || domain.MachineType("barca").IsValid() {
		t.Error("MachineType.IsValid")
	}
	if !domain.FuelTypeDiesel.IsValid() || domain.FuelType("carbune").IsValid() {
		t.Error("FuelType.IsValid")
	}
	if !domain.ImplementTypePlow.IsValid() || domain.ImplementType("x").IsValid() {
		t.Error("ImplementType.IsValid")
	}
	if !domain.AssetStatusMaintenance.IsValid() || domain.AssetStatus("pierdut").IsValid() || !domain.MachineStatusActive.IsValid() {
		t.Error("AssetStatus.IsValid")
	}
	if !domain.ResourceCategoryHarvest.IsValid() || domain.ResourceCategory("gaz").IsValid() {
		t.Error("ResourceCategory.IsValid")
	}
	if !domain.OperationTypeSeeding.IsValid() || domain.OperationType("plowing").IsValid() || domain.OperationType("").IsValid() {
		t.Error("OperationType.IsValid")
	}
	for _, ot := range domain.OperationTypes {
		if !ot.IsValid() || ot.Label() == string(ot) {
			t.Errorf("tipul %q nu are etichetă", ot)
		}
	}
	if domain.OperationType("plowing").Label() != "plowing" {
		t.Error("un tip necunoscut rămâne codul lui")
	}
}

func TestFieldCrop_ComputeYield(t *testing.T) {
	fc := &domain.FieldCrop{}
	fc.ComputeYield()
	if fc.YieldPerHa != nil {
		t.Error("fără producție și suprafață, randamentul e nil")
	}
	area, production := 10.0, 45.0
	fc = &domain.FieldCrop{PlantedAreaHa: &area, ProductionTotal: &production}
	fc.ComputeYield()
	if fc.YieldPerHa == nil || *fc.YieldPerHa != 4.5 {
		t.Errorf("randament: %v", fc.YieldPerHa)
	}
	zero := 0.0
	fc.PlantedAreaHa = &zero
	fc.ComputeYield()
	if fc.YieldPerHa != nil {
		t.Error("suprafața zero nu produce randament")
	}
}
