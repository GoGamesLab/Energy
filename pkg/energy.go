package energy

import (
	"log"

	container "github.com/GoGamesLab/Inventory/pkg"
)

type EnergyUnit = float64

type Supply struct {
	Fuel container.Storage
}

type EnergySource interface {
	EnergyCapacity() EnergyUnit
	GenerateEnergy(amount EnergyUnit) EnergyUnit
	Type() string
}

type EnergySink interface {
	EnergyDemand() EnergyUnit
	ConsumeEnergy() EnergyUnit
	Type() string
}

type EnergyManager struct {
	Produced      EnergyUnit
	Consumed      EnergyUnit
	EnergySources map[string]EnergySource
	EnergySinks   map[string]EnergySink
}

func NewEnergyManager() *EnergyManager {
	return &EnergyManager{
		Produced:      0.0,
		Consumed:      0.0,
		EnergySources: make(map[string]EnergySource),
		EnergySinks:   make(map[string]EnergySink),
	}
}

func (m *EnergyManager) RegisterSource(id string, n EnergySource) { m.EnergySources[id] = n }
func (m *EnergyManager) RegisterSink(id string, n EnergySink)     { m.EnergySinks[id] = n }

func (m *EnergyManager) GetEnergy() EnergyUnit { return m.Produced - m.Consumed }

func (m *EnergyManager) Update(converters []Converter, start EnergyUnit) {
	for _, sink := range m.EnergySinks {
		need := sink.EnergyDemand()
		if need == 0 {
			continue
		}

		got := EnergyUnit(0)

		for _, src := range m.EnergySources {
			if got >= need {
				break
			}

			var to EnergyUnit
			var base EnergyUnit = start
			for _, mc := range converters {
				c, err := GetConverter(mc.ID)
				if err != nil {
					log.Fatalf("🧨 Conversor desconhecido: %s", mc.ID)
				}

				to = c.FromBase(base) * EnergyUnit(mc.Quantity)
				if _, err := c.ToBase(to); err != nil {
					log.Fatalf("🧨 converter: %s", err)
				}

				base = to
			}

			remaining := need - got + base
			produced := src.GenerateEnergy(remaining)
			m.Produced += produced
			got += produced
		}

		if got+1e-9 >= need {
			m.Consumed += need
			sink.ConsumeEnergy()
		}
	}
}
