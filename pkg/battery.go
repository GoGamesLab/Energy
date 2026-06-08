package energy

import (
	"math"
)

// Bateria simples / reservatório
type Battery struct {
	capacity    EnergyUnit
	stored      EnergyUnit
	lossPerTick EnergyUnit
}

func NewBattery(cap EnergyUnit) *Battery {
	if cap < 0 {
		cap = 0
	}
	return &Battery{
		capacity: cap,
		stored:   0,
	}
}

func (b *Battery) Produce(amount EnergyUnit) (EnergyUnit, error) {
	if amount <= 0 {
		return 0, nil
	}
	space := b.capacity - b.stored
	if space <= 0 {
		return 0, nil
	}
	add := math.Min(space, amount)
	b.stored += add
	return add, nil
}

func (b *Battery) Consume(amount EnergyUnit) (EnergyUnit, error) {
	if amount <= 0 {
		return 0, nil
	}
	avail := b.stored
	if avail <= 0 {
		return 0, nil
	}
	use := math.Min(avail, amount)
	b.stored -= use
	return use, nil
}

func (b *Battery) Demand() EnergyUnit { return b.lossPerTick }

func (b *Battery) Peek() EnergyUnit { return b.stored }

func (b *Battery) Capacity() EnergyUnit { return b.capacity }

func (b *Battery) Stored() EnergyUnit { return b.stored }

// ApplyTickLoss aplica perda por tick; chamada externamente pelo simulador/tick loop.
func (b *Battery) ApplyTickLoss() {
	if b.lossPerTick <= 0 || b.stored <= 0 {
		return
	}
	l := math.Min(b.lossPerTick, b.stored)
	b.stored -= l
}
