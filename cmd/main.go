package main

import (
	"log/slog"
	"os"

	energy "github.com/GoGamesLab/Energy/pkg"
)

var Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

func main() {
	Logger.Info("🧳 Energy control start")

	energyManager := energy.NewEnergyManager()

	energyManager.RegisterConverter(energy.NewHeatConverter(0.9))
	energyManager.RegisterConverter(energy.NewElectricConverter(0.85))
}
