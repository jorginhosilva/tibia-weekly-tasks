package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Este arquivo grava a lista do populate.go no banco EM LOTES.
//
// Por que em lotes? Cada comando enviado ao banco é uma "viagem" de ida e volta.
// Com o banco no mesmo computador, a viagem leva milésimos de segundo. Com o
// banco na nuvem, leva ~0,2 s. Gravar 2.331 ligações uma a uma custava mais de
// 7.000 viagens (cerca de 25 minutos). Em lotes, são cerca de 15 viagens.
//
// O resultado no banco é o mesmo; só muda a quantidade de viagens.

// tamanhoLote é quantas linhas vão em cada comando INSERT. O MySQL aceita até
// 65.535 parâmetros por comando; com 500 linhas de 2 colunas usamos só 1.000.
const tamanhoLote = 500

// executorDeLotes é o que a gravação precisa do banco: executar comandos e ler
// linhas. Tanto *sql.DB quanto *sql.Tx cumprem este contrato.
type executorDeLotes interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// gravarEmLotes grava todas as criaturas, todos os itens e todas as ligações
// da lista. Roda dentro da transação aberta por quem chamou.
func gravarEmLotes(ctx context.Context, ex executorDeLotes, drops []dropInicial) error {
	criaturas, itens := nomesDistintos(drops)

	// 1) Criaturas e itens: cada nome uma vez só, vários por comando.
	if err := inserirNomes(ctx, ex, "creatures", criaturas); err != nil {
		return err
	}
	if err := inserirNomes(ctx, ex, "weekly_items", itens); err != nil {
		return err
	}

	// 2) Pergunta ao banco o número (id) de cada nome, uma vez só.
	idCriatura, err := lerIds(ctx, ex, "creatures")
	if err != nil {
		return err
	}
	idItem, err := lerIds(ctx, ex, "weekly_items")
	if err != nil {
		return err
	}

	// 3) Troca os nomes da lista pelos números e grava as ligações em lotes.
	ligacoes := make([]int64, 0, len(drops)*2)
	for i, d := range drops {
		c, ok := idCriatura[strings.ToLower(d.Criatura)]
		if !ok {
			return fmt.Errorf("drop %d (%s -> %s): a criatura não foi encontrada no banco depois de gravada", i+1, d.Criatura, d.Item)
		}
		it, ok := idItem[strings.ToLower(d.Item)]
		if !ok {
			return fmt.Errorf("drop %d (%s -> %s): o item não foi encontrado no banco depois de gravado", i+1, d.Criatura, d.Item)
		}
		ligacoes = append(ligacoes, c, it)
	}
	return inserirLigacoes(ctx, ex, ligacoes)
}

// nomesDistintos devolve os nomes de criaturas e de itens sem repetição, na
// ordem em que aparecem. Maiúsculas e minúsculas contam como o mesmo nome (o
// banco também trata assim); vale a primeira grafia encontrada.
func nomesDistintos(drops []dropInicial) (criaturas, itens []string) {
	vistasC := make(map[string]bool)
	vistosI := make(map[string]bool)
	for _, d := range drops {
		if k := strings.ToLower(d.Criatura); !vistasC[k] {
			vistasC[k] = true
			criaturas = append(criaturas, d.Criatura)
		}
		if k := strings.ToLower(d.Item); !vistosI[k] {
			vistosI[k] = true
			itens = append(itens, d.Item)
		}
	}
	return criaturas, itens
}

// inserirNomes grava os nomes na tabela, tamanhoLote por comando. O nome da
// tabela vem sempre de uma constante do código, nunca de texto digitado.
//
// "ON DUPLICATE KEY UPDATE name = name" não muda nada quando o nome já existe,
// então rodar de novo é seguro. (Diferente do INSERT IGNORE, ele NÃO esconde
// outros erros, como um nome grande demais para a coluna.)
func inserirNomes(ctx context.Context, ex executorDeLotes, tabela string, nomes []string) error {
	for inicio := 0; inicio < len(nomes); inicio += tamanhoLote {
		fim := min(inicio+tamanhoLote, len(nomes))
		lote := nomes[inicio:fim]

		args := make([]any, len(lote))
		for i, n := range lote {
			args[i] = n
		}
		consulta := "INSERT INTO " + tabela + " (name) VALUES " + marcadores(len(lote), 1) +
			" ON DUPLICATE KEY UPDATE name = name"
		if _, err := ex.ExecContext(ctx, consulta, args...); err != nil {
			return fmt.Errorf("gravando nomes em %s (linhas %d a %d): %w", tabela, inicio+1, fim, err)
		}
	}
	return nil
}

// lerIds devolve um mapa "nome em minúsculas" -> id de todas as linhas da tabela.
func lerIds(ctx context.Context, ex executorDeLotes, tabela string) (map[string]int64, error) {
	linhas, err := ex.QueryContext(ctx, "SELECT id, name FROM "+tabela)
	if err != nil {
		return nil, fmt.Errorf("lendo os ids de %s: %w", tabela, err)
	}
	defer linhas.Close()

	ids := make(map[string]int64)
	for linhas.Next() {
		var id int64
		var nome string
		if err := linhas.Scan(&id, &nome); err != nil {
			return nil, fmt.Errorf("lendo uma linha de %s: %w", tabela, err)
		}
		ids[strings.ToLower(nome)] = id
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("percorrendo %s: %w", tabela, err)
	}
	return ids, nil
}

// inserirLigacoes grava os pares (id da criatura, id do item), tamanhoLote
// pares por comando. A lista chega "achatada": [c1, i1, c2, i2, ...].
func inserirLigacoes(ctx context.Context, ex executorDeLotes, pares []int64) error {
	total := len(pares) / 2
	for inicio := 0; inicio < total; inicio += tamanhoLote {
		fim := min(inicio+tamanhoLote, total)

		args := make([]any, 0, (fim-inicio)*2)
		for _, v := range pares[inicio*2 : fim*2] {
			args = append(args, v)
		}
		consulta := "INSERT INTO creature_drops (creature_id, item_id) VALUES " + marcadores(fim-inicio, 2) +
			" ON DUPLICATE KEY UPDATE creature_id = creature_id"
		if _, err := ex.ExecContext(ctx, consulta, args...); err != nil {
			return fmt.Errorf("gravando ligações (%d a %d de %d): %w", inicio+1, fim, total, err)
		}
	}
	return nil
}

// marcadores monta os "?" de um INSERT com várias linhas.
// marcadores(3, 2) devolve "(?,?),(?,?),(?,?)".
func marcadores(linhas, colunas int) string {
	linha := "(" + strings.TrimSuffix(strings.Repeat("?,", colunas), ",") + ")"
	return strings.TrimSuffix(strings.Repeat(linha+",", linhas), ",")
}
