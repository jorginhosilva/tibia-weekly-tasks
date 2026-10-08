package main

import (
	"context"
	"strings"
	"testing"
)

// Este é o teste mais importante do projeto: ele roda sobre os dados REAIS.
// Se alguém digitar errado um nome no populate.go, o teste falha dizendo qual.
func TestDropsIniciaisSemErros(t *testing.T) {
	if problemas := ValidarDrops(dropsIniciais); len(problemas) > 0 {
		t.Fatalf("%d problema(s) no populate.go:\n  - %s", len(problemas), strings.Join(problemas, "\n  - "))
	}
}

func TestValidarDropsPegaOsErrosQueJaAconteceram(t *testing.T) {
	casos := []struct {
		nome     string
		drops    []dropInicial
		esperado string // trecho que deve aparecer em algum problema
	}{
		{"nome vazio", []dropInicial{{"", "Item"}}, "nome vazio"},
		{"espaço no fim", []dropInicial{{"Streaked ", "Item"}}, "espaço no começo ou no fim"},
		{"dois espaços", []dropInicial{{"Tentugly's  Head", "Item"}}, "dois espaços"},
		{"maiúsculas diferentes", []dropInicial{{"Two-Headed Turtle", "A"}, {"Two-headed Turtle", "B"}}, "formas diferentes"},
		{"ligação repetida", []dropInicial{{"Wyrm", "Wyrm Scale"}, {"wyrm", "wyrm scale"}}, "ligação repetida"},
		{"criatura e item trocados", []dropInicial{{"Two-Headed Turtle", "Glacier Shoes"}, {"Two-Headed Turtle Heads", "Two-Headed Turtle"}}, "criatura e item trocados"},
		{"erro de digitação em criatura", []dropInicial{{"Hellspawn", "Battle Shield"}, {"Hellspaw", "Knight Legs"}}, "quase iguais"},
		{"erro de digitação em item", []dropInicial{{"Armadile", "Bonebreaker"}, {"Darkfang", "Bonebreakr"}}, "quase iguais"},
	}
	for _, c := range casos {
		problemas := ValidarDrops(c.drops)
		if !strings.Contains(strings.Join(problemas, "\n"), c.esperado) {
			t.Errorf("%s: esperava um problema com %q, veio %q", c.nome, c.esperado, problemas)
		}
	}
}

// A garantia mais importante: dados com erro NUNCA chegam ao banco. O banco
// aqui é nil de propósito: se a validação não barrasse antes, o teste cairia.
func TestCargaRecusaDadosComErroSemTocarNoBanco(t *testing.T) {
	original := dropsIniciais
	defer func() { dropsIniciais = original }()
	dropsIniciais = []dropInicial{{"Hellspawn", "Battle Shield"}, {"Hellspaw", "Knight Legs"}}

	for nome, carga := range map[string]func() error{
		"InserirDadosTibia": func() error { return InserirDadosTibia(context.Background(), nil) },
		"RecriarDadosTibia": func() error { return RecriarDadosTibia(context.Background(), nil) },
	} {
		err := carga()
		if err == nil || !strings.Contains(err.Error(), "NADA foi gravado") {
			t.Errorf("%s: erro = %v, esperava recusar com 'NADA foi gravado'", nome, err)
		}
	}
}

func TestValidarDropsAceitaDadosBons(t *testing.T) {
	bons := []dropInicial{
		{"Bashmu", "Bashmu Fang"},
		{"Bashmu", "Bashmu Tongue"},
		{"Wyrm", "Wyrm Scale"},
	}
	if problemas := ValidarDrops(bons); len(problemas) != 0 {
		t.Errorf("dados bons foram recusados: %q", problemas)
	}
}

func TestUmaLetraDeDiferenca(t *testing.T) {
	casos := []struct {
		a, b string
		want bool
	}{
		{"hellspaw", "hellspawn", true},     // letra a menos
		{"hellflayeer", "hellflayer", true}, // letra a mais
		{"inferniach", "inferniarch", true}, // letra a menos no meio
		{"fermbras", "ferumbras", true},     // letra a menos
		{"hellhound", "hellgorak", false},   // nomes diferentes
		{"wyrm", "wyrm", false},             // iguais não contam
		{"abc", "axc", true},                // letra trocada
		{"abc", "xyc", false},               // duas trocas
		{"abc", "abcde", false},             // duas letras a mais
	}
	for _, c := range casos {
		if got := umaLetraDeDiferenca(c.a, c.b); got != c.want {
			t.Errorf("umaLetraDeDiferenca(%q, %q) = %v, esperava %v", c.a, c.b, got, c.want)
		}
	}
}
