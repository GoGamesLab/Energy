package energy

import (
	"errors"
	"fmt"
	"log"
	"math"

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
}

// Byproduct (p.ex. calor gerado ao gerar elétricidade)
type EnergyByproduct struct {
	Type  string // "heat", "radiation", ...
	Value EnergyUnit
}

// Converters registry and nodes
type EnergyManager struct {
	EnergyConverters map[string]EnergyConverter
	EnergyNodes      map[string]EnergyNode
}

func NewEnergyManager() *EnergyManager {
	return &EnergyManager{
		EnergyConverters: make(map[string]EnergyConverter),
		EnergyNodes:      make(map[string]EnergyNode),
	}
}

func (m *EnergyManager) RegisterConverter(c EnergyConverter) { m.EnergyConverters[c.ToType()] = c }

func (m *EnergyManager) RegisterNode(id string, n EnergyNode) { m.EnergyNodes[id] = n }

// SatisfyRequirement satisfaz um requisito: tenta prover req.Amount convertendo a energia necessária.
// Retorna a quantidade efetivamente fornecida (em unidades do requisito) e os subprodutos gerados.
func (m *EnergyManager) SatisfyRequirement(req EnergyRequirement, preferNodeIDs []string) (provided EnergyUnit, byproducts []EnergyByproduct, err error) {
	conv, ok := m.EnergyConverters[req.TypeHint]
	if !ok {
		return 0, nil, fmt.Errorf("no converter for type %q", req.TypeHint)
	}

	log.Printf("Using converter %s (eff=%.3f)", conv.ToType(), conv.Efficiency())

	if req.Amount <= 0 {
		return 0, nil, nil
	}

	eff := float64(conv.Efficiency())
	if eff <= 0 {
		return 0, nil, fmt.Errorf("invalid converter efficiency: %.3f", eff)
	}

	// Como o conversor tem uma perda, para entregar `req.Amount` de energia útil,
	// nós precisamos extrair um valor maior (`neededFromSource`) da fonte primária.
	neededFromSource := req.Amount / eff
	var totalSourceConsumed float64

	// Build preferred set para buscas O(1)
	preferred := make(map[string]struct{}, len(preferNodeIDs))
	for _, p := range preferNodeIDs {
		preferred[p] = struct{}{}
	}

	// ESTRATÉGIA 1: Iterar pelos nós preferenciais primeiro (Baterias/Buffers locais)
	for _, id := range preferNodeIDs {
		node, exists := m.EnergyNodes[id]
		if !exists {
			continue
		}

		avail := node.Peek()
		if avail <= 0 {
			continue
		}

		// Quanto ainda precisamos extrair da fonte primária neste passo
		remainingSourceNeeded := neededFromSource - totalSourceConsumed
		wantFromSource := math.Min(remainingSourceNeeded, avail)

		consumed, err := node.Consume(wantFromSource)
		if err != nil {
			continue
		}
		totalSourceConsumed += consumed

		// Subproduto: A perda da conversão vira calor
		loss := (1.0 - eff) * consumed
		if loss > 0 {
			byproducts = append(byproducts, EnergyByproduct{Type: "heat", Value: loss})
		}

		if totalSourceConsumed >= neededFromSource {
			break
		}
	}

	// ESTRATÉGIA 2: Se ainda faltar energia, tenta produzir de nós geradores (ex: Reatores),
	// pulando os nós preferenciais que já foram testados.
	if totalSourceConsumed < neededFromSource {
		for id, node := range m.EnergyNodes {
			if _, isPref := preferred[id]; isPref {
				continue
			}

			remainingSourceNeeded := neededFromSource - totalSourceConsumed
			prod, pErr := node.Produce(remainingSourceNeeded)
			if pErr != nil || prod <= 0 {
				continue
			}

			totalSourceConsumed += prod

			loss := (1.0 - eff) * prod
			if loss > 0 {
				byproducts = append(byproducts, EnergyByproduct{Type: "heat", Value: loss})
			}

			if totalSourceConsumed >= neededFromSource {
				break
			}
		}
	}

	if totalSourceConsumed == 0 {
		return 0, byproducts, errors.New("no energy could be provided")
	}

	// A energia útil efetivamente entregue para a máquina é o que consumimos da fonte
	// multiplicado pela eficiência do conversor.
	provided = totalSourceConsumed * eff

	return provided, byproducts, nil
}
