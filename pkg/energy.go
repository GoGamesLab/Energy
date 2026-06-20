package energy

import (
	container "github.com/GoGamesLab/Inventory/pkg"
)

type EnergyUnit = float64

type Supply struct {
	Fuel container.Storage
}

type EnergySource interface {
	Capacity() EnergyUnit
	Produce(amount EnergyUnit) EnergyUnit
	Type() string
}

type EnergySink interface {
	Demand() EnergyUnit
	Consume() EnergyUnit
	Type() string
}

type EnergyManager struct {
	EnergyConverters map[string]EnergyConverter
	EnergySources    map[string]EnergySource
	EnergySinks      map[string]EnergySink
}

func NewEnergyManager() *EnergyManager {
	return &EnergyManager{
		EnergyConverters: make(map[string]EnergyConverter),
		EnergySources:    make(map[string]EnergySource),
		EnergySinks:      make(map[string]EnergySink),
	}
}

func (m *EnergyManager) RegisterConverter(c EnergyConverter)      { m.EnergyConverters[c.ToType()] = c }
func (m *EnergyManager) RegisterSource(id string, n EnergySource) { m.EnergySources[id] = n }
func (m *EnergyManager) RegisterSink(id string, n EnergySink)     { m.EnergySinks[id] = n }

func (m *EnergyManager) Update() {
	for _, sink := range m.EnergySinks {
		need := sink.Demand()
		if need == 0 {
			continue
		}

		got := EnergyUnit(0)

		for _, src := range m.EnergySources {
			if got >= need {
				break
			}

			// TODO: tem de ver qual o tipo de energia da fonte
			// e converter para o tipo desejado pelo destino

			remaining := need - got
			produced := src.Produce(remaining)
			got += produced
		}

		if got+1e-9 >= need {
			sink.Consume()
		}
	}
}
