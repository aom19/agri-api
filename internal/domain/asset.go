package domain

type AssetStatus string

const (
	AssetStatusActive      AssetStatus = "active"
	AssetStatusMaintenance AssetStatus = "maintenance"
	AssetStatusInactive    AssetStatus = "inactive"
)

var validAssetStatuses = map[AssetStatus]struct{}{
	AssetStatusActive:      {},
	AssetStatusMaintenance: {},
	AssetStatusInactive:    {},
}

func (assetStatus AssetStatus) IsValid() bool {
	_, ok := validAssetStatuses[assetStatus]
	return ok
}

// Backward-compatible alias for existing references outside domain.
type MachineStatus = AssetStatus

const (
	MachineStatusActive      = AssetStatusActive
	MachineStatusMaintenance = AssetStatusMaintenance
	MachineStatusInactive    = AssetStatusInactive
)
