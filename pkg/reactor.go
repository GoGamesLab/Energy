package energy

import (
	"math"
	"sync"

	container "github.com/GoGamesLab/Inventory/pkg"
	materials "github.com/GoGamesLab/Materials/pkg"
)

// UraniumReactor químico/nuclear: consome combustível continuamente para gerar energia,
// armazenando-a em uma bateria interna até a capacidade máxima.
type UraniumReactor struct {
	mu          sync.Mutex
	efficiency  float64  // converte combustível para energia base
	fuel        *Supply  // combustível bruto armazenado (material)
	battery     *Battery // buffer interno que representa a energia armazenada
	burnRate    float64  // quantidade máxima de combustível que queima por tick
	lossPerTick EnergyUnit
}

// Combustível exclusivo do reator
const ReactorFuelID = container.ItemID(materials.UraniumID)

func NewReactor(eff float64, internalCapacity EnergyUnit, burnRate float64) *UraniumReactor {
	if eff <= 0 {
		eff = 0.5
	}
	if burnRate <= 0 {
		burnRate = 10.0
	}

	fuel := &Supply{
		Fuel: *container.NewStorage(),
	}
	fuel.Fuel.AddItem(ReactorFuelID, 0)
	return &UraniumReactor{
		efficiency: eff,
		fuel:       fuel,
		burnRate:   burnRate,
		battery:    NewBattery(internalCapacity),
	}
}

func (r *UraniumReactor) GetFuel() Supply {
	return Supply{
		Fuel: r.fuel.Fuel,
	}
}

// AddFuel adiciona combustível bruto (material) ao reator
func (r *UraniumReactor) AddFuel(amount float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if amount <= 0 {
		return
	}
	r.fuel.Fuel.AddItem(ReactorFuelID, float32(amount))
}

// Fuel retorna a quantidade atual de combustível bruto dentro do reator
func (r *UraniumReactor) Fuel() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return float64(r.fuel.Fuel.Items[ReactorFuelID])
}

// Update simula a passagem de tempo (tick). Transforma combustível em energia interna.
// Deve ser chamado em cada iteração do loop principal do jogo para que o reator funcione de forma passiva.
func (r *UraniumReactor) Update() []EnergyByproduct {
	r.mu.Lock()
	defer r.mu.Unlock()

	var byproducts []EnergyByproduct

	if r.fuel.Fuel.Items[ReactorFuelID] <= 0 {
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
	estoque := float64(r.fuel.Fuel.Items[ReactorFuelID])
	if fuelToBurn > estoque {
		fuelToBurn = estoque
		energyToProduce = fuelToBurn * r.efficiency
	}

	// Consumir o combustível e injetar a energia na bateria interna
	r.fuel.Fuel.RemoveItem(ReactorFuelID, float32(fuelToBurn))
	_, _ = r.battery.Produce(energyToProduce)

	// O calor gerado como subproduto é proporcional à ineficiência do combustível consumido
	loss := (1.0 - r.efficiency) * fuelToBurn
	if loss > 0 {
		byproducts = append(byproducts, EnergyByproduct{Type: "heat", Value: loss})
	}

	return byproducts
}

// --- Implementação da Interface EnergyNode (Delegações para a Bateria Interna) ---

func (r *UraniumReactor) Peek() EnergyUnit {
	return r.battery.Peek()
}

func (r *UraniumReactor) Consume(amount EnergyUnit) (EnergyUnit, error) {
	return r.battery.Consume(amount)
}

func (r *UraniumReactor) Produce(amount EnergyUnit) (EnergyUnit, error) {
	// Permite que sistemas externos injetem energia de volta no reator, caso necessário
	return r.battery.Produce(amount)
}

func (r *UraniumReactor) Capacity() EnergyUnit {
	return r.battery.Capacity()
}

func (r *UraniumReactor) Stored() EnergyUnit {
	return r.battery.Stored()
}

func (r *UraniumReactor) Demand() EnergyUnit { return r.lossPerTick }

func (r *UraniumReactor) ApplyTickLoss() {
	r.battery.ApplyTickLoss()
}
