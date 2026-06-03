package energy

import "errors"

// Converter entre EnergyUnit (base) e representações semânticas
type EnergyConverter interface {
	TypeName() string
	FromBase(e EnergyUnit) float64
	ToBase(value float64) (EnergyUnit, error)
	Efficiency() float64
}

// HeatConverter: converte J <-> degrees-equivalent (design simplificado)
type HeatConverter struct {
	efficiency float64
}

func NewHeatConverter(efficiency float64) *HeatConverter {
	if efficiency <= 0 {
		efficiency = 0.9
	}
	return &HeatConverter{efficiency: efficiency}
}

func (c *HeatConverter) TypeName() string { return "heat" }

func (c *HeatConverter) FromBase(e EnergyUnit) float64 { return float64(e) * c.efficiency }

func (c *HeatConverter) ToBase(value float64) (EnergyUnit, error) {
	if c.efficiency <= 0 {
		return 0, errors.New("invalid converter efficiency")
	}
	return EnergyUnit(value / c.efficiency), nil
}

func (c *HeatConverter) Efficiency() float64 { return c.efficiency }

// ElectricConverter: lida com throughput/loss (simplificado)
type ElectricConverter struct {
	efficiency float64
}

func NewElectricConverter(efficiency float64) *ElectricConverter {
	if efficiency <= 0 {
		efficiency = 0.85
	}
	return &ElectricConverter{efficiency: efficiency}
}

func (c *ElectricConverter) TypeName() string { return "electric" }

func (c *ElectricConverter) FromBase(e EnergyUnit) float64 { return float64(e) * c.efficiency }

func (c *ElectricConverter) ToBase(value float64) (EnergyUnit, error) {
	if c.efficiency <= 0 {
		return 0, errors.New("invalid converter efficiency")
	}
	return EnergyUnit(value / c.efficiency), nil
}

func (c *ElectricConverter) Efficiency() float64 { return c.efficiency }
