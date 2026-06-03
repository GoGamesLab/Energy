package energy

import "math"

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

	// 1. Puxa a energia disponível da fonte
	got, err := from.(EnergySink).Consume(want) // Cast temporário se for um EnergyNode, ou ajuste a interface
	if err != nil || got <= 0 {
		return 0, nil, err
	}

	// 2. Aplica perda por distância (Resistência do cabo)
	lossFrac := math.Max(0.0, p.LossPerUnitDistance*p.Distance)
	if lossFrac > 1.0 {
		lossFrac = 1.0
	}

	lossAmount := float64(got) * lossFrac
	effective := EnergyUnit(float64(got) - lossAmount)

	// 3. Aplica o teto de vazão (Throughput) do cabo
	if p.ThroughputPerTick > 0 && effective > p.ThroughputPerTick {
		// Se estourar o limite, o excesso é descartado ou acumula (aqui limitamos o que é entregue)
		// Em jogos como Factorio, o cabo simplesmente gargala a transferência.
		effective = p.ThroughputPerTick
	}

	// 4. Entrega a energia convertida para o consumidor
	// Como a interface EnergySink usa Consume(), para fins de rede elétrica pura,
	// assumimos que injetar energia em um Sink/Node é feito via Produce se for Node,
	// ou se o 'to' for um acumulador.
	// Nota: Se 'to' for uma máquina consumidora pura, ela precisará de um buffer interno (Battery)
	produced, perr := to.(EnergySource).Produce(effective)
	if perr != nil {
		return 0, nil, perr
	}

	// Gera calor proporcional à perda de transmissão nos fios
	if lossAmount > 0 {
		byproducts = append(byproducts, EnergyByproduct{Type: "heat", Value: lossAmount})
	}

	return produced, byproducts, nil
}
