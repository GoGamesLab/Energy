package energy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type ConverterID string

type EnergyConverter interface {
	FromType() string
	FromBase(e EnergyUnit) EnergyUnit
	ToType() string
	ToBase(value EnergyUnit) (EnergyUnit, error)
	Efficiency() float32
}

type CustomConverter struct {
	IDStr      ConverterID `json:"id"`
	FromEnergy string      `json:"from_type"`
	ToEnergy   string      `json:"to_type"`
	Eff        float32     `json:"efficiency"`
}

var _ EnergyConverter = (*CustomConverter)(nil)

func (c *CustomConverter) ID() ConverterID                  { return c.IDStr }
func (c *CustomConverter) FromType() string                 { return c.FromEnergy }
func (c *CustomConverter) ToType() string                   { return c.ToEnergy }
func (c *CustomConverter) Efficiency() float32              { return c.Eff }
func (c *CustomConverter) FromBase(e EnergyUnit) EnergyUnit { return e * EnergyUnit(c.Eff) }
func (c *CustomConverter) ToBase(value EnergyUnit) (EnergyUnit, error) {
	if c.Eff <= 0 {
		return 0, errors.New("invalid converter efficiency")
	}

	return EnergyUnit(value / EnergyUnit(c.Eff)), nil
}

var EnergyConverters = make(map[ConverterID]*CustomConverter)

func RegisterConverter(c *CustomConverter) error {
	if c.IDStr == "" {
		return errors.New("🧨 Elemento não pode ter ID vazio")
	}
	if c.Eff <= 0 || c.Eff > 1.0 {
		c.Eff = 0.85
	}
	if _, exists := EnergyConverters[c.IDStr]; exists {
		return fmt.Errorf("🧨 Elemento com ID %s já registrado", c.IDStr)
	}

	EnergyConverters[c.IDStr] = c
	return nil
}

func GetConverter(id ConverterID) (*CustomConverter, error) {
	if c, ok := EnergyConverters[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("🧨 Converter %v: not found", id)
}

func LoadConvertersFromJSON(convertersPath string) error {
	cData, err := os.ReadFile(convertersPath)
	if err != nil {
		return err
	}

	var cList []CustomConverter
	if err := json.Unmarshal(cData, &cList); err != nil {
		return err
	}

	for _, c := range cList {
		conv := c
		if err := RegisterConverter(&conv); err != nil {
			return err
		}
	}

	return nil
}
