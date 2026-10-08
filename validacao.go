package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// Este arquivo protege o banco contra dados errados no populate.go.
//
// Os erros que já aconteceram de verdade foram: letra a mais ou a menos
// ("Hellspaw" em vez de "Hellspawn"), criatura e item trocados de lugar
// ("Two-Headed Turtle Heads" como criatura), espaços sobrando e o mesmo nome
// escrito com maiúsculas diferentes. Como o banco aceita qualquer texto
// diferente como um nome NOVO, esses erros viravam duplicatas no site.
// Agora a lista é conferida ANTES de qualquer coisa ir para o banco.

// nomesEmAmbosPapeis lista nomes que são, DE VERDADE, criatura e item ao mesmo
// tempo (em minúsculas). Normalmente é um erro de inversão, então só entra aqui
// o que foi conferido no jogo.
var nomesEmAmbosPapeis = map[string]bool{
	// A CONFERIR NO JOGO: a criatura e o item têm o mesmo nome. Se for erro,
	// corrija no populate.go e apague a linha daqui.
	"gore horn":     true,
	"lizard zaogun": true,
}

// paresParecidosPermitidos lista pares de nomes que diferem em UMA letra mas
// são realmente coisas diferentes (em minúsculas, em ordem alfabética: "a|b").
// Só entra aqui o que foi conferido no jogo.
var paresParecidosPermitidos = map[string]bool{
	// Criaturas diferentes de verdade:
	"gorga|gorgo":             true,
	"mutated bat|mutated rat": true,
	"werebear|wereboar":       true,
	// A CONFERIR NO JOGO: podem ser erro de digitação. Se forem, corrija no
	// populate.go e apague a linha daqui.
	"defile|defiler":          true,
	"goggle cake|google cake": true,
}

// ValidarDrops confere a lista inteira e devolve TODOS os problemas
// encontrados (uma frase para cada). Lista vazia = dados bons.
func ValidarDrops(drops []dropInicial) []string {
	var problemas []string
	add := func(formato string, args ...any) { problemas = append(problemas, fmt.Sprintf(formato, args...)) }

	// 1) Cada nome, individualmente.
	for i, d := range drops {
		for _, campo := range []struct{ rotulo, nome string }{{"criatura", d.Criatura}, {"item", d.Item}} {
			switch {
			case campo.nome == "":
				add("linha %d da lista: %s com nome vazio", i+1, campo.rotulo)
			case campo.nome != strings.TrimSpace(campo.nome):
				add("%s %q tem espaço no começo ou no fim", campo.rotulo, campo.nome)
			case strings.Contains(campo.nome, "  "):
				add("%s %q tem dois espaços seguidos", campo.rotulo, campo.nome)
			case utf8.RuneCountInString(campo.nome) > tamanhoMaxNome:
				add("%s %q passa de %d caracteres", campo.rotulo, campo.nome, tamanhoMaxNome)
			}
		}
	}

	// 2) Ligações repetidas e o mesmo nome escrito de formas diferentes.
	criaturas := map[string]map[string]bool{} // minúsculo -> grafias usadas
	itens := map[string]map[string]bool{}
	pares := map[string]bool{}
	for _, d := range drops {
		registrar(criaturas, d.Criatura)
		registrar(itens, d.Item)

		chave := strings.ToLower(d.Criatura) + " -> " + strings.ToLower(d.Item)
		if pares[chave] {
			add("ligação repetida: %q -> %q", d.Criatura, d.Item)
		}
		pares[chave] = true
	}
	for _, grupo := range []struct {
		rotulo string
		nomes  map[string]map[string]bool
	}{{"criatura", criaturas}, {"item", itens}} {
		for minusculo, grafias := range grupo.nomes {
			if len(grafias) > 1 {
				add("%s escrito de formas diferentes: %s", grupo.rotulo, listar(grafias))
			}
			_ = minusculo
		}
	}

	// 3) Nome que aparece como criatura E como item: quase sempre é inversão.
	for minusculo := range criaturas {
		if _, tambemItem := itens[minusculo]; tambemItem && !nomesEmAmbosPapeis[minusculo] {
			add("%q aparece como criatura e como item (criatura e item trocados?)", primeira(criaturas[minusculo]))
		}
	}

	// 4) Nomes que diferem em uma letra: quase sempre erro de digitação.
	for _, grupo := range []struct {
		rotulo string
		nomes  map[string]map[string]bool
	}{{"criaturas", criaturas}, {"itens", itens}} {
		for _, par := range paresPorUmaLetra(chaves(grupo.nomes)) {
			if !paresParecidosPermitidos[par[0]+"|"+par[1]] {
				add("%s com nomes quase iguais (erro de digitação?): %q e %q",
					grupo.rotulo, primeira(grupo.nomes[par[0]]), primeira(grupo.nomes[par[1]]))
			}
		}
	}

	sort.Strings(problemas)
	return problemas
}

// tamanhoMaxNome é o tamanho da coluna name no banco (VARCHAR(100)).
const tamanhoMaxNome = 100

func registrar(m map[string]map[string]bool, nome string) {
	k := strings.ToLower(nome)
	if m[k] == nil {
		m[k] = map[string]bool{}
	}
	m[k][nome] = true
}

func primeira(grafias map[string]bool) string {
	for _, g := range listarOrdenado(grafias) {
		return g
	}
	return ""
}

func listarOrdenado(grafias map[string]bool) []string {
	var l []string
	for g := range grafias {
		l = append(l, g)
	}
	sort.Strings(l)
	return l
}

func listar(grafias map[string]bool) string {
	return strings.Join(listarOrdenado(grafias), " / ")
}

func chaves(m map[string]map[string]bool) []string {
	var l []string
	for k := range m {
		l = append(l, k)
	}
	sort.Strings(l)
	return l
}

// paresPorUmaLetra devolve os pares (em ordem alfabética) de nomes que diferem
// por exatamente uma letra a mais, a menos ou trocada.
func paresPorUmaLetra(nomes []string) [][2]string {
	var pares [][2]string
	for i := 0; i < len(nomes); i++ {
		for j := i + 1; j < len(nomes); j++ {
			if umaLetraDeDiferenca(nomes[i], nomes[j]) {
				pares = append(pares, [2]string{nomes[i], nomes[j]})
			}
		}
	}
	return pares
}

// umaLetraDeDiferenca diz se a e b diferem em exatamente uma inserção,
// remoção ou troca de letra. Percorre as duas palavras uma única vez.
func umaLetraDeDiferenca(a, b string) bool {
	ra, rb := []rune(a), []rune(b)
	if len(ra) < len(rb) {
		ra, rb = rb, ra // ra é sempre a mais longa
	}
	if len(ra)-len(rb) > 1 {
		return false
	}
	i, j, diferencas := 0, 0, 0
	for i < len(ra) && j < len(rb) {
		if ra[i] == rb[j] {
			i++
			j++
			continue
		}
		diferencas++
		if diferencas > 1 {
			return false
		}
		if len(ra) == len(rb) {
			j++ // letra trocada: avança nas duas
		}
		i++ // letra a mais em ra: avança só nela
	}
	diferencas += (len(ra) - i) + (len(rb) - j)
	return diferencas == 1
}
