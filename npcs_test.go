package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLerCsvNpcAceitaArquivoBom(t *testing.T) {
	bom := string(rune(0xFEFF)) // caractere invisível que o Excel grava no começo
	texto := bom + "item;preco\r\nChaos Mace;9000\r\n\r\n Crystal Mace ; 12000\r\nCat's Paw;2000"

	precos, problemas := lerCsvNpc(strings.NewReader(texto), "teste.csv")
	if len(problemas) > 0 {
		t.Fatalf("não esperava problemas, veio: %v", problemas)
	}
	esperado := []precoNpc{{"Chaos Mace", 9000}, {"Crystal Mace", 12000}, {"Cat's Paw", 2000}}
	if !reflect.DeepEqual(precos, esperado) {
		t.Errorf("preços = %v, esperava %v", precos, esperado)
	}
}

func TestLerCsvNpcApontaOsErrosDeFormato(t *testing.T) {
	texto := strings.Join([]string{
		"item;preco",
		"Chaos Mace;9.000",     // linha 2: preço com ponto
		"Crystal Mace",         // linha 3: falta a coluna do preço
		";500",                 // linha 4: nome vazio
		"Glacier  Robe;11000",  // linha 5: dois espaços
		"Buckle;0",             // linha 6: preço zero
		"War Axe;999999999999", // linha 7: preço alto demais
		"Buckle;7000",          // linha 8: válido
		"buckle;7000",          // linha 9: repetido (maiúsculas não importam)
		"Halberd;-400",         // linha 10: preço negativo
	}, "\n")

	_, problemas := lerCsvNpc(strings.NewReader(texto), "rashid.csv")
	esperados := []string{
		"rashid.csv, linha 2: ", "9.000",
		"rashid.csv, linha 3: ", "2 colunas",
		"rashid.csv, linha 4: ", "vazio",
		"rashid.csv, linha 5: ", "dois espaços",
		"rashid.csv, linha 6: ", "zero",
		"rashid.csv, linha 7: ", "alto demais",
		"rashid.csv, linha 9: ", "repetido",
		"rashid.csv, linha 10: ", "-400",
	}
	tudo := strings.Join(problemas, "\n")
	for _, trecho := range esperados {
		if !strings.Contains(tudo, trecho) {
			t.Errorf("faltou um problema com %q; problemas encontrados:\n%s", trecho, tudo)
		}
	}
	if len(problemas) != 8 {
		t.Errorf("esperava 8 problemas, veio %d:\n%s", len(problemas), tudo)
	}
}

func TestLerCsvNpcExigeOCabecalho(t *testing.T) {
	_, problemas := lerCsvNpc(strings.NewReader("Chaos Mace;9000\n"), "x.csv")
	if len(problemas) != 1 || !strings.Contains(problemas[0], "primeira linha") {
		t.Errorf("esperava um problema de cabeçalho, veio %v", problemas)
	}
}

func TestLerCsvNpcArquivoVazioNaoEErro(t *testing.T) {
	for _, texto := range []string{"", "item;preco\n"} {
		precos, problemas := lerCsvNpc(strings.NewReader(texto), "x.csv")
		if len(precos) != 0 || len(problemas) != 0 {
			t.Errorf("%q: esperava nada, veio preços %v e problemas %v", texto, precos, problemas)
		}
	}
}

func TestNomeDoNpc(t *testing.T) {
	casos := map[string]string{
		"rashid.csv":       "Rashid",
		"yasir.csv":        "Yasir",
		"esrik_pompan.csv": "Esrik Pompan",
		"black-bert.csv":   "Black Bert",
		"nah'bob.csv":      "Nah'bob", // tudo em minúsculas: só a primeira letra sobe
		"Nah'Bob.csv":      "Nah'Bob", // maiúscula escrita no arquivo é respeitada
		"Esrik_Pompan.csv": "Esrik Pompan",
	}
	for arquivo, esperado := range casos {
		if got := nomeDoNpc(arquivo); got != esperado {
			t.Errorf("nomeDoNpc(%q) = %q, esperava %q", arquivo, got, esperado)
		}
	}
}

func TestLerPreco(t *testing.T) {
	validos := map[string]int{"9000": 9000, "1": 1, "0": 0}
	for texto, esperado := range validos {
		if n, ok := lerPreco(texto); !ok || n != esperado {
			t.Errorf("lerPreco(%q) = %d, %v; esperava %d, true", texto, n, ok, esperado)
		}
	}
	for _, texto := range []string{"", "9.000", "9,000", "-5", "+5", "9000gp", " 9000", "1e3"} {
		if _, ok := lerPreco(texto); ok {
			t.Errorf("lerPreco(%q) deveria recusar", texto)
		}
	}
}

func TestAvisosNpc(t *testing.T) {
	listas := []listaNpc{
		{Npc: "Rashid", Precos: []precoNpc{{"Chaos Mace", 9000}, {"Terra Hood", 2500}}},
		{Npc: "Yasir"}, // sem itens
	}
	avisos := avisosNpc(listas, []string{"Chaos Mace", "Terra Hod"})

	esperado := []string{
		`Rashid: "Terra Hood" não está na lista da weekly`,
		"Yasir: o arquivo não tem itens (o NPC não aparece no site)",
	}
	if !reflect.DeepEqual(avisos, esperado) {
		t.Errorf("avisos = %q, esperava %q", avisos, esperado)
	}
}

func TestLerListasNpcLeAPastaEmOrdem(t *testing.T) {
	pasta := t.TempDir()
	escrever(t, filepath.Join(pasta, "yasir.csv"), "item;preco\nBashmu Fang;500\n")
	escrever(t, filepath.Join(pasta, "rashid.csv"), "item;preco\nChaos Mace;9000\n")
	escrever(t, filepath.Join(pasta, "leia-me.txt"), "não é csv: deve ser ignorado")

	listas, err := lerListasNpc(pasta)
	if err != nil {
		t.Fatal(err)
	}
	esperado := []listaNpc{
		{Npc: "Rashid", Precos: []precoNpc{{"Chaos Mace", 9000}}},
		{Npc: "Yasir", Precos: []precoNpc{{"Bashmu Fang", 500}}},
	}
	if !reflect.DeepEqual(listas, esperado) {
		t.Errorf("listas = %v, esperava %v", listas, esperado)
	}
}

func TestLerListasNpcJuntaOsProblemasDeTodosOsArquivos(t *testing.T) {
	pasta := t.TempDir()
	escrever(t, filepath.Join(pasta, "rashid.csv"), "item;preco\nChaos Mace;9.000\n")
	escrever(t, filepath.Join(pasta, "yasir.csv"), "item;preco\nBashmu Fang\n")

	_, err := lerListasNpc(pasta)
	if err == nil {
		t.Fatal("esperava erro")
	}
	for _, trecho := range []string{"2 problema(s)", "NADA foi gravado", "rashid.csv, linha 2", "yasir.csv, linha 2"} {
		if !strings.Contains(err.Error(), trecho) {
			t.Errorf("o erro não cita %q:\n%v", trecho, err)
		}
	}
}

func TestLerListasNpcPastaInexistenteNaoEErro(t *testing.T) {
	listas, err := lerListasNpc(filepath.Join(t.TempDir(), "nao-existe"))
	if err != nil || len(listas) != 0 {
		t.Errorf("esperava nada e sem erro, veio %v, %v", listas, err)
	}
}

// Confere os arquivos REAIS da pasta dados/npcs: um erro de formato digitado
// à mão aparece aqui, antes de chegar ao banco.
func TestArquivosDosNpcsSemErroDeFormato(t *testing.T) {
	if _, err := lerListasNpc(pastaNpcs); err != nil {
		t.Error(err)
	}
}
