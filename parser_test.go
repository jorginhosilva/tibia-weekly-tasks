package main

import (
	"errors"
	"reflect"
	"testing"
)

// Teste "table-driven": uma tabela de casos (entrada -> saída esperada).
// Rode com:  go test ./...
func TestExtrairItens(t *testing.T) {
	casos := []struct {
		nome     string
		log      string
		esperado []Item
	}{
		{
			nome:     "item com quantidade e artigo",
			log:      "Looted Items:\n  1x a war hammer",
			esperado: []Item{{"war hammer", 1}},
		},
		{
			nome:     "x dentro do nome nao e confundido com quantidade (onyx)",
			log:      "Looted Items:\n  2x an onyx chip",
			esperado: []Item{{"onyx chip", 2}},
		},
		{
			nome:     "sem quantidade e com x no nome (fox)",
			log:      "Looted Items:\n  a fox tail",
			esperado: []Item{{"fox tail", 1}},
		},
		{
			nome:     "x no nome depois da quantidade (wax)",
			log:      "Looted Items:\n  1x a wax candle",
			esperado: []Item{{"wax candle", 1}},
		},
		{
			nome:     "nome comecando com a mas sem artigo",
			log:      "Looted Items:\n  3x amber staff",
			esperado: []Item{{"amber staff", 3}},
		},
		{
			nome:     "quebras de linha do Windows (CRLF)",
			log:      "Looted Items:\r\n  1x a war hammer\r\n  4x a bashmu tongue\r\n",
			esperado: []Item{{"war hammer", 1}, {"bashmu tongue", 4}},
		},
		{
			nome:     "para ao encontrar outra secao",
			log:      "Looted Items:\n  1x a war hammer\nKilled Monsters:\n  37x bashmu",
			esperado: []Item{{"war hammer", 1}},
		},
		{
			nome:     "ignora o que vem antes da secao (incluindo Loot: com numero)",
			log:      "Loot: 251,888\nKilled Monsters:\n  37x bashmu\nLooted Items:\n  1x a war hammer",
			esperado: []Item{{"war hammer", 1}},
		},
		{
			nome:     "cabecalho alternativo Loot:",
			log:      "Loot:\n  2x a red gem",
			esperado: []Item{{"red gem", 2}},
		},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			obtido, err := extrairItens(c.log)
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if !reflect.DeepEqual(obtido, c.esperado) {
				t.Errorf("obtido %v, esperado %v", obtido, c.esperado)
			}
		})
	}
}

func TestExtrairItensSemSecaoLoot(t *testing.T) {
	logs := []string{
		"",
		"texto qualquer sem nada de loot",
		"Loot: 251,888\nSupplies: 45,464", // só o total de gold, não é a lista de itens
	}
	for _, l := range logs {
		_, err := extrairItens(l)
		if !errors.Is(err, ErrSemSecaoLoot) {
			t.Errorf("log %q: esperava ErrSemSecaoLoot, veio %v", l, err)
		}
	}
}

func TestIdentificarWeekly(t *testing.T) {
	weekly := map[string]string{
		"bashmu fang":   "Bashmu Fang",
		"bashmu tongue": "Bashmu Tongue",
	}
	itens := []Item{
		{"platinum coin", 707},
		{"bashmu fang", 1},
		{"bashmu tongue", 4},
		{"bashmu fang", 2}, // repetido: as quantidades devem somar
	}

	obtido := identificarWeekly(itens, weekly)
	esperado := []Item{{"Bashmu Fang", 3}, {"Bashmu Tongue", 4}}

	if !reflect.DeepEqual(obtido, esperado) {
		t.Errorf("obtido %v, esperado %v", obtido, esperado)
	}
}
