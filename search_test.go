package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestEscaparLike(t *testing.T) {
	casos := map[string]string{
		"bashmu":  "bashmu",
		"100%":    `100\%`,
		"a_b":     `a\_b`,
		`a\b`:     `a\\b`,
		"%_%":     `\%\_\%`,
		"war ham": "war ham",
	}
	for entrada, esperado := range casos {
		if obtido := escaparLike(entrada); obtido != esperado {
			t.Errorf("escaparLike(%q) = %q, esperado %q", entrada, obtido, esperado)
		}
	}
}

func TestPrepararTermo(t *testing.T) {
	padrao, err := prepararTermo("  Bashmu  ")
	if err != nil || padrao != "%Bashmu%" {
		t.Errorf("obtido (%q, %v), esperado (%q, nil)", padrao, err, "%Bashmu%")
	}

	// Quem digita só "%" não pode conseguir listar o banco inteiro.
	padrao, err = prepararTermo("%")
	if err != nil || padrao != `%\%%` {
		t.Errorf("obtido (%q, %v), esperado (%q, nil)", padrao, err, `%\%%`)
	}

	for _, vazio := range []string{"", "   ", "\t\n"} {
		if _, err := prepararTermo(vazio); !errors.Is(err, ErrBuscaVazia) {
			t.Errorf("termo %q: esperava ErrBuscaVazia, veio %v", vazio, err)
		}
	}

	if _, err := prepararTermo(strings.Repeat("a", tamanhoMaxBusca+1)); !errors.Is(err, ErrBuscaLonga) {
		t.Errorf("esperava ErrBuscaLonga, veio %v", err)
	}
	if _, err := prepararTermo(strings.Repeat("a", tamanhoMaxBusca)); err != nil {
		t.Errorf("exatamente %d caracteres deveria ser aceito, veio %v", tamanhoMaxBusca, err)
	}
}

func TestAgruparPorCriatura(t *testing.T) {
	drops := []TaskDrop{
		{"Bashmu", "Bashmu Fang"},
		{"Bashmu", "Bashmu Tongue"},
		{"Juvenile Bashmu", "Bashmu Fang"},
	}

	obtido := agruparPorCriatura(drops)
	esperado := []CriaturaComItens{
		{"Bashmu", []string{"Bashmu Fang", "Bashmu Tongue"}},
		{"Juvenile Bashmu", []string{"Bashmu Fang"}},
	}
	if !reflect.DeepEqual(obtido, esperado) {
		t.Errorf("obtido %v, esperado %v", obtido, esperado)
	}

	if len(agruparPorCriatura(nil)) != 0 {
		t.Error("sem linhas, o resultado deveria ser vazio")
	}
}

func TestAgruparPorItem(t *testing.T) {
	drops := []TaskDrop{
		{"Bashmu", "Bashmu Fang"},
		{"Juvenile Bashmu", "Bashmu Fang"},
		{"Bashmu", "Bashmu Tongue"},
	}

	obtido := agruparPorItem(drops)
	esperado := []ItemComCriaturas{
		{"Bashmu Fang", []string{"Bashmu", "Juvenile Bashmu"}},
		{"Bashmu Tongue", []string{"Bashmu"}},
	}
	if !reflect.DeepEqual(obtido, esperado) {
		t.Errorf("obtido %v, esperado %v", obtido, esperado)
	}
}

func TestAgruparPorNpc(t *testing.T) {
	linhas := []linhaPreco{
		{"Rashid", "Chaos Mace", 9000},
		{"Rashid", "Crystal Mace", 12000},
		{"Yasir", "Bashmu Fang", 500},
	}

	obtido := agruparPorNpc(linhas)
	esperado := []NpcComPrecos{
		{"Rashid", []PrecoNoNpc{{"Chaos Mace", 9000}, {"Crystal Mace", 12000}}},
		{"Yasir", []PrecoNoNpc{{"Bashmu Fang", 500}}},
	}
	if !reflect.DeepEqual(obtido, esperado) {
		t.Errorf("obtido %v, esperado %v", obtido, esperado)
	}
}

func TestPrecosPorNomeDoItemIgnoraMaiusculas(t *testing.T) {
	linhas := []linhaPreco{
		{"Chaos Mace", "Rashid", 9000},
		{"chaos mace", "Yasir", 8000},
	}

	obtido := precosPorNomeDoItem(linhas)
	esperado := map[string][]PrecoNoNpc{
		"chaos mace": {{"Rashid", 9000}, {"Yasir", 8000}},
	}
	if !reflect.DeepEqual(obtido, esperado) {
		t.Errorf("obtido %v, esperado %v", obtido, esperado)
	}
}
