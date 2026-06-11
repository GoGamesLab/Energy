package energy

import "errors"

type EnergyConverter interface {
	FromType() string
	FromBase(e EnergyUnit) EnergyUnit
	ToType() string
	ToBase(value EnergyUnit) (EnergyUnit, error)
	Efficiency() float32
}

type baseConverter struct{ efficiency float32 }

func newBaseConverter(efficiency float32) baseConverter {
	if efficiency <= 0 {
		efficiency = 0.85
	}
	return baseConverter{efficiency: efficiency}
}

func (b baseConverter) FromBase(e EnergyUnit) EnergyUnit { return e * EnergyUnit(b.efficiency) }

func (b baseConverter) ToBase(value EnergyUnit) (EnergyUnit, error) {
	if b.efficiency <= 0 {
		return 0, errors.New("invalid converter efficiency")
	}
	return EnergyUnit(value / EnergyUnit(b.efficiency)), nil
}

func (b baseConverter) Efficiency() float32 { return b.efficiency }

// Boiler
type Boiler struct{ base baseConverter }

func NewBoiler(eff float32) *Boiler { return &Boiler{base: newBaseConverter(eff)} }

func (c *Boiler) FromType() string                        { return "heat" }
func (c *Boiler) FromBase(e EnergyUnit) EnergyUnit        { return c.base.FromBase(e) }
func (c *Boiler) ToType() string                          { return "kinetic" }
func (c *Boiler) ToBase(v EnergyUnit) (EnergyUnit, error) { return c.base.ToBase(v) }
func (c *Boiler) Efficiency() float32                     { return c.base.Efficiency() }

// Dynamo
type Dynamo struct{ base baseConverter }

func NewDynamo(eff float32) *Dynamo { return &Dynamo{base: newBaseConverter(eff)} }

func (c *Dynamo) FromType() string                        { return "kinetic" }
func (c *Dynamo) FromBase(e EnergyUnit) EnergyUnit        { return c.base.FromBase(e) }
func (c *Dynamo) ToType() string                          { return "electric" }
func (c *Dynamo) ToBase(v EnergyUnit) (EnergyUnit, error) { return c.base.ToBase(v) }
func (c *Dynamo) Efficiency() float32                     { return c.base.Efficiency() }
