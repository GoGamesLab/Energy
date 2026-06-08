package main

import (
	"fmt"
	"log/slog"
	"os"

	energy "github.com/GoGamesLab/Energy/pkg"
)

var Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

func runReactorA(em *energy.EnergyManager) {
	battery := energy.NewBattery(10000) // capacity
	em.RegisterNode("battery-1", battery)

	// processo requer 500 "heat" units
	battery.Produce(1000)
	batteryHeatProvided, battteryByProducts, _ := em.SatisfyRequirement(
		energy.EnergyRequirement{Amount: 500, TypeHint: "heat"},
		[]string{"reactor-1"})

	Logger.Info(fmt.Sprintf("Battery provided: %f heat", batteryHeatProvided))
	for _, b := range battteryByProducts {
		Logger.Info(fmt.Sprintf("- byproduct %s = %f", b.Type, b.Value))
	}

	reactorA := energy.NewUraniumReactor(0.75, 1000, 0.9)
	em.RegisterNode("reactor-1", reactorA)

	reactorA.AddFuel(2)
	for reactorA.Fuel() > 0 {
		updateByProducts := reactorA.Update()
		Logger.Info(fmt.Sprintf("Combustivel %f, Energia %f", reactorA.Fuel(), reactorA.Stored()))
		for _, b := range updateByProducts {
			Logger.Info(fmt.Sprintf("- passive byproduct %s = %f", b.Type, b.Value))
		}
		reactorElectricityProvided, reactorByProducts, _ := em.SatisfyRequirement(
			energy.EnergyRequirement{Amount: 1000, TypeHint: "electric"},
			[]string{"reactor-1"})
		Logger.Info(fmt.Sprintf("Reactor provided: %f electricity", reactorElectricityProvided))
		for _, b := range reactorByProducts {
			Logger.Info(fmt.Sprintf("- byproduct %s = %f", b.Type, b.Value))
		}
		if reactorElectricityProvided < 1000 {
			break
		}
	}
}

func runReactorB() {
	// 1. Configura o cenário
	// Reator: 75% eficiência, cabe 100k de energia interna, queima 1 combustíveis por tick
	reactorB := energy.NewUraniumReactor(0.75, 1000, 1.0)

	// Destino: Uma subestação de baterias na sua fábrica para acumular o que vem do reator
	factoryBattery := energy.NewBattery(50000)

	// Cabo de Transmissão: Throughput máximo de 500 unidades/tick, 2% de perda por bloco, distância de 10 blocos (20% de perda total)
	cable := &energy.EnergyPipe{
		ThroughputPerTick:   500.0,
		LossPerUnitDistance: 0.02,
		Distance:            10.0,
	}

	// 2. Abastece o reator com combustível bruto e processa um tick de queima
	reactorB.AddFuel(30.0)
	fmt.Printf("Novo reator B com %f elementos de combustível\n", reactorB.Fuel())
	reactorByproducts := reactorB.Update()

	fmt.Println("--- TICK 1: Geração no Reator ---")
	fmt.Printf("Combustível restante no Reator: %.2f\n", reactorB.Fuel())
	fmt.Printf("Energia gerada e acumulada no buffer do Reator: %.2f J\n", reactorB.Stored())
	for _, bp := range reactorByproducts {
		fmt.Printf("  [Subproduto Reator] %s gerado: %.2f\n", bp.Type, bp.Value)
	}
	fmt.Println()

	// 3. Tenta transferir 400 unidades de energia do reator para a bateria da fábrica através do cabo
	fmt.Println("--- TICK 1: Transmissão via Cabo ---")
	desejado := 400.0
	enviado, cableByproducts, err := cable.Transfer(reactorB, factoryBattery, desejado)
	if err != nil {
		fmt.Printf("Erro na transmissão: %v\n", err)
		return
	}

	// 4. Resultados da transferência
	fmt.Printf("Tentou puxar: %.2f J\n", desejado)
	fmt.Printf("Efetivamente estocado na Fábrica (após perdas): %.2f J\n", enviado)
	fmt.Printf("Energia restante no buffer do Reator: %.2f J\n", reactorB.Stored())

	for _, bp := range cableByproducts {
		// Esse é o calor que vai dissipar nos contêineres dos cabos/fiação!
		fmt.Printf("  [Subproduto Cabo] %s dissipado na linha: %.2f\n", bp.Type, bp.Value)
	}
}

func main() {
	Logger.Info("🧳 Energy control start")

	energyManager := energy.NewEnergyManager()

	energyManager.RegisterConverter(energy.NewHeatConverter(0.9))
	energyManager.RegisterConverter(energy.NewElectricConverter(0.85))

	// processo contínuo
	runReactorA(energyManager)

	// processo discreto
	runReactorB()
}
