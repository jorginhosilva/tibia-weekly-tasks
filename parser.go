package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ErrSemSecaoLoot é um "erro conhecido": quem chamar pode reconhecê-lo com
// errors.Is(err, ErrSemSecaoLoot) e reagir de um jeito específico.
var ErrSemSecaoLoot = errors.New("log inválido ou sem seção de loot")

// Item é um item com a sua quantidade. Serve tanto para o que foi lido do log
// quanto para o que foi identificado como item de Weekly Task.
type Item struct {
	Nome       string
	Quantidade int
}

// reLinhaItem descreve o formato de uma linha de item do log:
//
//	[quantidade + "x" + espaço] [artigo: a / an / the] nome do item
//
// Exemplos: "37x a platinum coin", "2x an onyx chip", "a fox tail".
// Como a quantidade só vale se for NÚMEROS no começo, o "x" de dentro de
// "fox" ou "onyx" nunca é confundido com o "x" da quantidade.
//
//	grupo 1 = quantidade (opcional)   grupo 2 = nome do item
var reLinhaItem = regexp.MustCompile(`(?i)^(?:(\d+)x\s+)?(?:(?:an|a|the)\s+)?(.+)$`)

// ehCabecalhoDeLoot diz se a linha é o título da seção de itens.
// Compara a linha INTEIRA, então "Loot: 251,888" (o total de gold, que fica no
// topo do log) NÃO é confundido com o início da lista de itens.
func ehCabecalhoDeLoot(linha string) bool {
	l := strings.ToLower(linha)
	return l == "looted items:" || l == "loot:"
}

// extrairItens lê o texto do log e devolve os itens da seção de loot.
// Não usa banco de dados nem imprime nada: só transforma texto em dados.
// Por isso é fácil de testar.
func extrairItens(logTexto string) ([]Item, error) {
	var itens []Item
	achouSecao := false

	scanner := bufio.NewScanner(strings.NewReader(logTexto))
	for scanner.Scan() {
		// TrimSpace também remove o \r dos logs copiados no Windows (\r\n).
		linha := strings.TrimSpace(scanner.Text())

		// Antes do cabeçalho de loot, só procuramos por ele.
		if !achouSecao {
			achouSecao = ehCabecalhoDeLoot(linha)
			continue
		}

		if linha == "" {
			continue
		}

		// Uma linha terminada em ":" é o título de OUTRA seção
		// (ex.: "Killed Monsters:"). A lista de itens acabou.
		if strings.HasSuffix(linha, ":") {
			break
		}

		// reLinhaItem aceita qualquer linha não vazia (o nome é "(.+)"), então
		// FindStringSubmatch nunca devolve nil aqui.
		partes := reLinhaItem.FindStringSubmatch(linha)

		quantidade := 1 // sem "Nx" na linha, consideramos 1 unidade
		if partes[1] != "" {
			n, err := strconv.Atoi(partes[1])
			if err != nil {
				return nil, fmt.Errorf("quantidade inválida na linha %q: %w", linha, err)
			}
			quantidade = n
		}

		nome := strings.ToLower(strings.TrimSpace(partes[2]))
		if nome == "" {
			continue
		}
		itens = append(itens, Item{Nome: nome, Quantidade: quantidade})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("lendo o texto do log: %w", err)
	}
	if !achouSecao {
		return nil, ErrSemSecaoLoot
	}

	return itens, nil
}

// carregarItensWeekly busca TODOS os itens de weekly task numa única consulta.
// Devolve um map: nome em minúsculas -> nome oficial.
// (Antes eram dezenas de consultas, uma por linha do log.)
func carregarItensWeekly(ctx context.Context, db *sql.DB) (map[string]string, error) {
	linhas, err := db.QueryContext(ctx, "SELECT name FROM weekly_items")
	if err != nil {
		return nil, fmt.Errorf("buscando os itens de weekly task: %w", err)
	}
	defer linhas.Close()

	itens := make(map[string]string)
	for linhas.Next() {
		var nome string
		if err := linhas.Scan(&nome); err != nil {
			return nil, fmt.Errorf("lendo um item de weekly task: %w", err)
		}
		itens[strings.ToLower(nome)] = nome
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("percorrendo os itens de weekly task: %w", err)
	}

	return itens, nil
}

// identificarWeekly compara os itens do log com o map de itens de weekly task.
// Itens repetidos no log têm as quantidades somadas.
// Também é uma função "pura": sem banco, sem impressão.
func identificarWeekly(itensDoLog []Item, weekly map[string]string) []Item {
	var encontrados []Item
	posicao := make(map[string]int) // nome oficial -> posição na lista "encontrados"

	for _, item := range itensDoLog {
		// "valor, ok := mapa[chave]": ok é false quando a chave não existe.
		nomeOficial, ok := weekly[item.Nome]
		if !ok {
			continue
		}

		if i, jaTem := posicao[nomeOficial]; jaTem {
			encontrados[i].Quantidade += item.Quantidade
			continue
		}

		posicao[nomeOficial] = len(encontrados)
		encontrados = append(encontrados, Item{Nome: nomeOficial, Quantidade: item.Quantidade})
	}

	return encontrados
}

// ProcessarLogHunt junta as três etapas e DEVOLVE o resultado.
// Quem chama decide o que fazer com ele (imprimir, salvar, enviar...).
func ProcessarLogHunt(ctx context.Context, db *sql.DB, logTexto string) ([]Item, error) {
	itensDoLog, err := extrairItens(logTexto)
	if err != nil {
		return nil, fmt.Errorf("lendo o log da hunt: %w", err)
	}

	weekly, err := carregarItensWeekly(ctx, db)
	if err != nil {
		return nil, err
	}

	return identificarWeekly(itensDoLog, weekly), nil
}
