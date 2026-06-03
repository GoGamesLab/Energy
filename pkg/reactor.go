package energy

import (
	"math"
	"sync"
)

// Reactor químico/nuclear: consome combustível continuamente para gerar energia,
// armazenando-a em uma bateria interna até a capacidade máxima.
type Reactor struct {
	mu          sync.Mutex
	efficiency  float64  // converte combustível para energia base
	fuel        float64  // combustível bruto armazenado (material)
	battery     *Battery // buffer interno que representa a energia armazenada
	burnRate    float64  // quantidade máxima de combustível que queima por tick
	lossPerTick EnergyUnit
}

func NewReactor(eff float64, internalCapacity EnergyUnit, burnRate float64) *Reactor {
	if eff <= 0 {
		eff = 0.5
	}
	if burnRate <= 0 {
		burnRate = 10.0
	}
	return &Reactor{
		efficiency: eff,
		burnRate:   burnRate,
		battery:    NewBattery(internalCapacity),
	}
}

// AddFuel adiciona combustível bruto (material) ao reator
func (r *Reactor) AddFuel(amount float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if amount <= 0 {
		return
	}
	r.fuel += amount
}

// Fuel retorna a quantidade atual de combustível bruto dentro do reator
func (r *Reactor) Fuel() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.fuel
}

// Update simula a passagem de tempo (tick). Transforma combustível em energia interna.
// Deve ser chamado em cada iteração do loop principal do jogo para que o reator funcione de forma passiva.
func (r *Reactor) Update() []EnergyByproduct {
	r.mu.Lock()
	defer r.mu.Unlock()

	var byproducts []EnergyByproduct

	if r.fuel <= 0 {
		return nil
	}

	// Verificar espaço no buffer interno de energia
	currentStored := r.battery.Stored()
	capacity := r.battery.Capacity()
	availableSpace := capacity - currentStored

	// Se a bateria interna estiver cheia, o reator suspende a queima temporariamente (comportamento Factorio)
	if availableSpace <= 0 {
		return nil
	}

	// Calcular quanta energia tentaremos gerar baseando-se no limite de queima (burnRate)
	maxEnergyFromBurn := r.burnRate * r.efficiency
	energyToProduce := math.Min(maxEnergyFromBurn, availableSpace)
	fuelToBurn := energyToProduce / r.efficiency

	// Limitar o consumo à quantidade real de combustível em estoque
	if fuelToBurn > r.fuel {
		fuelToBurn = r.fuel
		energyToProduce = fuelToBurn * r.efficiency
	}

	// Consumir o combustível e injetar a energia na bateria interna
	r.fuel -= fuelToBurn
	_, _ = r.battery.Produce(energyToProduce)

	// O calor gerado como subproduto é proporcional à ineficiência do combustível consumido
	loss := (1.0 - r.efficiency) * fuelToBurn
	if loss > 0 {
		byproducts = append(byproducts, EnergyByproduct{Type: "heat", Value: loss})
	}

	return byproducts
}

// --- Implementação da Interface EnergyNode (Delegações para a Bateria Interna) ---

func (r *Reactor) Peek() EnergyUnit {
	return r.battery.Peek()
}

func (r *Reactor) Consume(amount EnergyUnit) (EnergyUnit, error) {
	return r.battery.Consume(amount)
}

func (r *Reactor) Produce(amount EnergyUnit) (EnergyUnit, error) {
	// Permite que sistemas externos injetem energia de volta no reator, caso necessário
	return r.battery.Produce(amount)
}

func (r *Reactor) Capacity() EnergyUnit {
	return r.battery.Capacity()
}

func (r *Reactor) Stored() EnergyUnit {
	return r.battery.Stored()
}

func (r *Reactor) Demand() EnergyUnit { return r.lossPerTick }

func (r *Reactor) ApplyTickLoss() {
	r.battery.ApplyTickLoss()
}
