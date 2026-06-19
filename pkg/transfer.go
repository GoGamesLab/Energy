package energy

import (
	"fmt"
	"math"
)

// Rede/Flow: canal com throughput e perda por distância
type EnergyPipe struct {
	ThroughputPerTick   EnergyUnit
	LossPerUnitDistance float64 // frac lost per unit distance
	Distance            float64
}

// Transfer tira energia de uma FONTE (Source), aplica perdas e entrega a um CONSUMIDOR (Sink)
func (p *EnergyPipe) Transfer(from EnergySource, to EnergySink, want EnergyUnit) (transferred EnergyUnit, byproducts []EnergyByproduct, err error) {
	if want <= 0 {
		return 0, nil, nil
	}

	// 1. Puxa energia com segurança se a fonte permitir consumo direto (como uma bateria/node)
	var got EnergyUnit
	if consumerFrom, ok := from.(interface {
		Consume(EnergyUnit) (EnergyUnit, error)
	}); ok {
		got, err = consumerFrom.Consume(want)
	} else {
		// Se for uma fonte pura (ex: painel solar), nós apenas "olhamos" ou assumimos a produção
		got = math.Min(from.Peek(), want)
		// Aqui você precisaria de um método na fonte para confirmar a extração, ex: from.Extract(got)
	}
	if err != nil || got <= 0 {
		return 0, nil, err
	}

	// 2. Aplica perda por distância
	lossFrac := math.Max(0.0, p.LossPerUnitDistance*p.Distance)
	if lossFrac > 1.0 {
		lossFrac = 1.0
	}
	lossAmount := float64(got) * lossFrac
	effective := EnergyUnit(float64(got) - lossAmount)

	if p.ThroughputPerTick > 0 && effective > p.ThroughputPerTick {
		effective = p.ThroughputPerTick
	}

	// 3. Entrega a energia convertida para o consumidor checando se ele aceita recarga (EnergyNode/Battery)
	var produced EnergyUnit
	if nodeTo, ok := to.(interface {
		Produce(EnergyUnit) (EnergyUnit, error)
	}); ok {
		produced, err = nodeTo.Produce(effective)
		if err != nil {
			return 0, nil, err
		}
	} else {
		return 0, nil, fmt.Errorf("target sink cannot receive/store energy directly from pipe")
	}

	if lossAmount > 0 {
		byproducts = append(byproducts, EnergyByproduct{Type: "heat", Value: lossAmount})
	}

	return produced, byproducts, nil
}
