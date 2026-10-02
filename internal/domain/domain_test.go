package domain

import "testing"

func TestEnumValidation(t *testing.T) {
	if !MachineTypeTractor.IsValid() || MachineType("barca").IsValid() {
		t.Error("MachineType.IsValid")
	}
	if !FuelTypeDiesel.IsValid() || FuelType("carbune").IsValid() {
		t.Error("FuelType.IsValid")
	}
	if !ImplementTypePlow.IsValid() || ImplementType("x").IsValid() {
		t.Error("ImplementType.IsValid")
	}
	if !AssetStatusMaintenance.IsValid() || AssetStatus("pierdut").IsValid() || !MachineStatusActive.IsValid() {
		t.Error("AssetStatus.IsValid")
	}
	if !ResourceCategoryHarvest.IsValid() || ResourceCategory("gaz").IsValid() {
		t.Error("ResourceCategory.IsValid")
	}
}

func TestFieldCrop_ComputeYield(t *testing.T) {
	fc := &FieldCrop{}
	fc.ComputeYield()
	if fc.YieldPerHa != nil {
		t.Error("fără producție și suprafață, randamentul e nil")
	}
	area, production := 10.0, 45.0
	fc = &FieldCrop{PlantedAreaHa: &area, ProductionTotal: &production}
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
