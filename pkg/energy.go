package energy

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sync"

	container "github.com/GoGamesLab/Inventory/pkg"
)

// Unidade de energia base (joules-equivalente)
type EnergyUnit = float64

// Fonte de combustível para fontes que consomem materiais ou substâncias
type Supply struct {
	Fuel container.Storage
}

// Fonte/sumidouro genérico de energia
type EnergySource interface {
	Produce(amount EnergyUnit) (produced EnergyUnit, err error) // tenta produzir até amount
	Peek() EnergyUnit                                           // quantidade disponível (instantânea)
}

// Consome energia
type EnergySink interface {
	Consume(amount EnergyUnit) (consumed EnergyUnit, err error) // tenta consumir até amount
	Demand() EnergyUnit                                         // demanda desejada (p.ex. por tick)
}

// Fonte e sumidouro combinados (baterias, nodes)
type EnergyNode interface {
	EnergySource
	EnergySink
	Capacity() EnergyUnit
	Stored() EnergyUnit
}

// Hint para processos que requerem energia
type EnergyRequirement struct {
	Amount   EnergyUnit // em unidades base
	TypeHint string     // "heat", "electric", "kinetic", "nuclear" -- usado para escolher conversor
	MinTemp  float64    // opcional (para heat)
	MaxTemp  float64
}

// Byproduct (p.ex. calor gerado ao gerar elétrico)
type EnergyByproduct struct {
	Type  string  // "heat", "radiation", ...
	Value float64 // em representação semântica (não em EnergyUnit)
}

// Converters registry and nodes
type EnergyManager struct {
	mu         sync.Mutex
	converters map[string]EnergyConverter
	nodes      map[string]EnergyNode
}

func NewEnergyManager() *EnergyManager {
	return &EnergyManager{
		converters: make(map[string]EnergyConverter),
		nodes:      make(map[string]EnergyNode),
	}
}

func (m *EnergyManager) RegisterConverter(c EnergyConverter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.converters[c.TypeName()] = c
}

func (m *EnergyManager) RegisterNode(id string, n EnergyNode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodes[id] = n
}

// Satisfaz um requisito: tenta prover Amount do TypeHint
// Retorna quantidade efetivamente fornecida (base units) e byproducts gerados
func (m *EnergyManager) SatisfyRequirement(req EnergyRequirement, preferNodeIDs []string) (provided EnergyUnit, byproducts []EnergyByproduct, err error) {
	m.mu.Lock()
	conv, ok := m.converters[req.TypeHint]
	m.mu.Unlock()
	if !ok {
		return 0, nil, fmt.Errorf("no converter for type %q", req.TypeHint)
	}

	log.Printf("Using converter %s (eff=%.3f)", conv.TypeName(), conv.Efficiency())

	need := req.Amount // in base units
	if need <= 0 {
		return 0, nil, nil
	}

	// build preferred set for O(1) checks
	preferred := make(map[string]struct{}, len(preferNodeIDs))
	for _, p := range preferNodeIDs {
		preferred[p] = struct{}{}
	}

	// strategy: iterate preferred nodes first, try to consume stored base units
	var totalProvided EnergyUnit
	for _, id := range preferNodeIDs {
		m.mu.Lock()
		node, exists := m.nodes[id]
		m.mu.Unlock()
		if !exists {
			continue
		}

		avail := node.Peek()
		if avail <= 0 {
			continue
		}

		want := math.Min(need-totalProvided, avail)
		consumed, err := node.Consume(want)
		if err != nil {
			continue
		}
		totalProvided += consumed

		// Byproduct example: conversion loss becomes heat
		loss := (1.0 - conv.Efficiency()) * float64(consumed)
		if loss > 0 {
			byproducts = append(byproducts, EnergyByproduct{Type: "heat", Value: loss})
		}

		if totalProvided >= need {
			break
		}
	}

	// If still lacking, attempt to produce from nodes that can Produce (reactors),
	// skipping the preferred nodes already tried.
	if totalProvided < need {
		m.mu.Lock()
		for id, node := range m.nodes {
			if _, isPref := preferred[id]; isPref {
				continue
			}

			// try produce
			toProduce := need - totalProvided
			prod, pErr := node.Produce(toProduce)
			if pErr != nil || prod <= 0 {
				continue
			}
			totalProvided += prod
			loss := (1.0 - conv.Efficiency()) * float64(prod)
			if loss > 0 {
				byproducts = append(byproducts, EnergyByproduct{Type: "heat", Value: loss})
			}
			if totalProvided >= need {
				break
			}
		}
		m.mu.Unlock()
	}

	if totalProvided == 0 {
		return 0, byproducts, errors.New("no energy could be provided")
	}

	return totalProvided, byproducts, nil
}
