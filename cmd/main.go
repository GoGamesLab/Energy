package main

import (
	"fmt"
	"log/slog"
	"os"

	energy "github.com/GoGamesLab/Energy/pkg"
)

var Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

func main() {
	Logger.Info("🧳 Energy control start")

	energyManager := energy.EnergyManagerInstance()

	Logger.Info(fmt.Sprintf("Energy manager %s created", energyManager))
}
