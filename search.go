package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Erros "conhecidos" da busca. Quando o site existir, ele poderá reconhecê-los
// com errors.Is e mostrar uma mensagem amigável em vez de um erro técnico.
var (
	ErrBuscaVazia = errors.New("digite ao menos um caractere para buscar")
	ErrBuscaLonga = errors.New("o termo de busca é longo demais")
)

const (
	tamanhoMaxBusca = 100 // igual ao tamanho da coluna name (VARCHAR(100))
	limiteDrops     = 500 // teto de linhas por busca, para uma busca não pesar o banco
)

// CriaturaComItens é o resultado de buscar por criatura:
// uma criatura e os itens de weekly task que ela dropa.
type CriaturaComItens struct {
	Criatura string
	Itens    []string
}

// ItemComCriaturas é o resultado de buscar por item:
// um item de weekly task e as criaturas que o dropam.
type ItemComCriaturas struct {
	Item      string
	Criaturas []string
}

// PrecoNoNpc é um nome com um preço. Na busca por NPC, o nome é o do ITEM;
// nos preços de um item, o nome é o do NPC que o compra.
type PrecoNoNpc struct {
	Nome  string
	Preco int
}

// NpcComPrecos é o resultado de buscar por NPC:
// um NPC e quanto ele paga por cada item.
type NpcComPrecos struct {
	Npc   string
	Itens []PrecoNoNpc
}

// limitePrecos é o teto de linhas da busca por NPC. É maior que limiteDrops
// porque um NPC sozinho já pode ter centenas de itens (o Yasir tem mais de 300).
const limitePrecos = 2000

// As duas consultas ficam como constantes: o texto do SQL nunca é montado
// juntando o que o usuário digitou. O termo entra só pelo "?" (parâmetro),
// e é isso que protege contra SQL injection, ainda mais num site público.
const (
	sqlBuscaPorCriatura = `
		SELECT c.name, i.name
		FROM creature_drops cd
		JOIN creatures c ON cd.creature_id = c.id
		JOIN weekly_items i ON cd.item_id = i.id
		WHERE c.name LIKE ?
		ORDER BY c.name, i.name
		LIMIT ?`

	sqlBuscaPorItem = `
		SELECT c.name, i.name
		FROM creature_drops cd
		JOIN creatures c ON cd.creature_id = c.id
		JOIN weekly_items i ON cd.item_id = i.id
		WHERE i.name LIKE ?
		ORDER BY i.name, c.name
		LIMIT ?`

	// O termo vale para o nome do NPC OU do item: "rashid" traz todos os itens
	// do Rashid; "banana sash" traz só esse item, no NPC que o compra.
	sqlBuscaPorNpc = `
		SELECT n.name, p.item_name, p.price
		FROM npc_prices p
		JOIN npcs n ON p.npc_id = n.id
		WHERE n.name LIKE ? OR p.item_name LIKE ?
		ORDER BY n.name, p.item_name
		LIMIT ?`

	// Preços dos itens cujo nome contém o termo: usada junto da busca por item,
	// para mostrar quanto cada NPC paga. O mais caro aparece primeiro.
	sqlPrecosPorItem = `
		SELECT p.item_name, n.name, p.price
		FROM npc_prices p
		JOIN npcs n ON p.npc_id = n.id
		WHERE p.item_name LIKE ?
		ORDER BY p.item_name, p.price DESC, n.name
		LIMIT ?`
)

// BuscarPorCriatura procura criaturas cujo nome CONTÉM o termo digitado
// (ex.: "bashmu" acha "Bashmu" e "Juvenile Bashmu") e devolve, para cada uma,
// os itens de weekly task que ela dropa.
func BuscarPorCriatura(ctx context.Context, db *sql.DB, termo string) ([]CriaturaComItens, error) {
	padrao, err := prepararTermo(termo)
	if err != nil {
		return nil, err
	}

	drops, err := consultarDrops(ctx, db, sqlBuscaPorCriatura, padrao)
	if err != nil {
		return nil, fmt.Errorf("buscando a criatura %q: %w", termo, err)
	}

	return agruparPorCriatura(drops), nil
}

// BuscarPorItem procura itens de weekly task cujo nome CONTÉM o termo digitado
// e devolve, para cada um, as criaturas que o dropam.
func BuscarPorItem(ctx context.Context, db *sql.DB, termo string) ([]ItemComCriaturas, error) {
	padrao, err := prepararTermo(termo)
	if err != nil {
		return nil, err
	}

	drops, err := consultarDrops(ctx, db, sqlBuscaPorItem, padrao)
	if err != nil {
		return nil, fmt.Errorf("buscando o item %q: %w", termo, err)
	}

	return agruparPorItem(drops), nil
}

// BuscarPorNpc procura pelo nome do NPC ou do item e devolve os NPCs com os
// itens e quanto pagam por cada:
//   - o NPC cujo NOME contém o termo aparece com TODOS os seus itens;
//   - os outros NPCs aparecem só com os itens cujo nome contém o termo.
func BuscarPorNpc(ctx context.Context, db *sql.DB, termo string) ([]NpcComPrecos, error) {
	padrao, err := prepararTermo(termo)
	if err != nil {
		return nil, err
	}

	linhas, err := consultarPrecos(ctx, db, sqlBuscaPorNpc, padrao, padrao)
	if err != nil {
		return nil, fmt.Errorf("buscando o NPC %q: %w", termo, err)
	}

	return agruparPorNpc(linhas), nil
}

// PrecosPorItem devolve, para cada item cujo nome contém o termo, os NPCs que
// o compram e por quanto. A chave do mapa é o nome do item em minúsculas, para
// casar com o resultado de BuscarPorItem mesmo com maiúsculas diferentes.
func PrecosPorItem(ctx context.Context, db *sql.DB, termo string) (map[string][]PrecoNoNpc, error) {
	padrao, err := prepararTermo(termo)
	if err != nil {
		return nil, err
	}

	linhas, err := consultarPrecos(ctx, db, sqlPrecosPorItem, padrao)
	if err != nil {
		return nil, fmt.Errorf("buscando os preços do item %q: %w", termo, err)
	}

	return precosPorNomeDoItem(linhas), nil
}

// linhaPreco é uma linha das consultas de preço: o nome que agrupa (NPC ou
// item), o outro nome e o preço.
type linhaPreco struct {
	grupo, nome string
	preco       int
}

// consultarPrecos roda uma das consultas de preço. "padroes" são os valores dos
// "?" do WHERE (um por "?"); o limite de linhas é sempre o último "?".
func consultarPrecos(ctx context.Context, db *sql.DB, consulta string, padroes ...string) ([]linhaPreco, error) {
	args := make([]any, 0, len(padroes)+1)
	for _, p := range padroes {
		args = append(args, p)
	}
	args = append(args, limitePrecos)

	linhas, err := db.QueryContext(ctx, consulta, args...)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	var resultado []linhaPreco
	for linhas.Next() {
		var l linhaPreco
		if err := linhas.Scan(&l.grupo, &l.nome, &l.preco); err != nil {
			return nil, fmt.Errorf("lendo um preço: %w", err)
		}
		resultado = append(resultado, l)
	}
	return resultado, linhas.Err()
}

// limiteSugestoes é quantos nomes o autocomplete devolve (poucos, para a lista caber na tela).
const limiteSugestoes = 12

// Sugestões: nomes que contêm o termo, com os que COMEÇAM por ele primeiro
// (digitar "deep" deve mostrar "Deepworm" antes de "Sea Deepling").
// O texto do SQL é fixo; só o termo entra por "?".
const (
	sqlSugestoesCriatura = `
		SELECT name FROM creatures
		WHERE name LIKE ?
		ORDER BY (name LIKE ?) DESC, name
		LIMIT ?`

	sqlSugestoesItem = `
		SELECT name FROM weekly_items
		WHERE name LIKE ?
		ORDER BY (name LIKE ?) DESC, name
		LIMIT ?`

	// Na aba de NPC, as sugestões trazem nomes de NPC E de itens que os NPCs
	// compram (o UNION já tira os repetidos).
	sqlSugestoesNpc = `
		SELECT name FROM (
			SELECT name FROM npcs
			UNION
			SELECT item_name FROM npc_prices
		) AS nomes
		WHERE name LIKE ?
		ORDER BY (name LIKE ?) DESC, name
		LIMIT ?`
)

// SugerirCriaturas devolve nomes de criaturas para o autocomplete.
func SugerirCriaturas(ctx context.Context, db *sql.DB, termo string) ([]string, error) {
	return sugerir(ctx, db, sqlSugestoesCriatura, termo)
}

// SugerirItens devolve nomes de itens de weekly task para o autocomplete.
func SugerirItens(ctx context.Context, db *sql.DB, termo string) ([]string, error) {
	return sugerir(ctx, db, sqlSugestoesItem, termo)
}

// SugerirNpcs devolve nomes de NPCs e de itens que eles compram para o autocomplete.
func SugerirNpcs(ctx context.Context, db *sql.DB, termo string) ([]string, error) {
	return sugerir(ctx, db, sqlSugestoesNpc, termo)
}

func sugerir(ctx context.Context, db *sql.DB, consulta, termo string) ([]string, error) {
	contem, err := prepararTermo(termo)
	if err != nil {
		return nil, err
	}
	comeca := strings.TrimSuffix(strings.TrimPrefix(contem, "%"), "%") + "%"

	linhas, err := db.QueryContext(ctx, consulta, contem, comeca, limiteSugestoes)
	if err != nil {
		return nil, fmt.Errorf("buscando sugestões para %q: %w", termo, err)
	}
	defer linhas.Close()

	nomes := []string{} // slice vazio (e não nil) vira [] no JSON
	for linhas.Next() {
		var nome string
		if err := linhas.Scan(&nome); err != nil {
			return nil, fmt.Errorf("lendo uma sugestão: %w", err)
		}
		nomes = append(nomes, nome)
	}
	return nomes, linhas.Err()
}

// prepararTermo valida o que o usuário digitou e o transforma no padrão do LIKE.
func prepararTermo(termo string) (string, error) {
	termo = strings.TrimSpace(termo)
	if termo == "" {
		return "", ErrBuscaVazia
	}
	if utf8.RuneCountInString(termo) > tamanhoMaxBusca {
		return "", ErrBuscaLonga
	}
	return "%" + escaparLike(termo) + "%", nil
}

// escaparLike neutraliza os caracteres que têm significado especial no LIKE:
// % (qualquer texto), _ (qualquer letra) e a própria barra invertida.
// Sem isso, quem digitasse só "%" faria o banco devolver tudo.
func escaparLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// consultarDrops roda uma das consultas acima e devolve as linhas lidas.
func consultarDrops(ctx context.Context, db *sql.DB, consulta, padrao string) ([]TaskDrop, error) {
	linhas, err := db.QueryContext(ctx, consulta, padrao, limiteDrops)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	return lerDrops(linhas)
}

// agruparPorCriatura junta as linhas (criatura, item) numa lista de criaturas,
// cada uma com todos os seus itens. A ordem de chegada é mantida.
func agruparPorCriatura(drops []TaskDrop) []CriaturaComItens {
	var resultado []CriaturaComItens
	posicao := make(map[string]int) // nome da criatura -> posição em "resultado"

	for _, d := range drops {
		i, existe := posicao[d.Criatura]
		if !existe {
			i = len(resultado)
			posicao[d.Criatura] = i
			resultado = append(resultado, CriaturaComItens{Criatura: d.Criatura})
		}
		resultado[i].Itens = append(resultado[i].Itens, d.Item)
	}

	return resultado
}

// agruparPorNpc junta as linhas (NPC, item, preço) numa lista de NPCs, cada um
// com os seus itens e preços. A ordem de chegada é mantida.
func agruparPorNpc(linhas []linhaPreco) []NpcComPrecos {
	var resultado []NpcComPrecos
	posicao := make(map[string]int) // nome do NPC -> posição em "resultado"

	for _, l := range linhas {
		i, existe := posicao[l.grupo]
		if !existe {
			i = len(resultado)
			posicao[l.grupo] = i
			resultado = append(resultado, NpcComPrecos{Npc: l.grupo})
		}
		resultado[i].Itens = append(resultado[i].Itens, PrecoNoNpc{Nome: l.nome, Preco: l.preco})
	}

	return resultado
}

// precosPorNomeDoItem junta as linhas (item, NPC, preço) num mapa
// "item em minúsculas" -> NPCs que o compram e por quanto.
func precosPorNomeDoItem(linhas []linhaPreco) map[string][]PrecoNoNpc {
	precos := make(map[string][]PrecoNoNpc)
	for _, l := range linhas {
		chave := strings.ToLower(l.grupo)
		precos[chave] = append(precos[chave], PrecoNoNpc{Nome: l.nome, Preco: l.preco})
	}
	return precos
}

// agruparPorItem faz o mesmo, mas agrupando por item.
func agruparPorItem(drops []TaskDrop) []ItemComCriaturas {
	var resultado []ItemComCriaturas
	posicao := make(map[string]int) // nome do item -> posição em "resultado"

	for _, d := range drops {
		i, existe := posicao[d.Item]
		if !existe {
			i = len(resultado)
			posicao[d.Item] = i
			resultado = append(resultado, ItemComCriaturas{Item: d.Item})
		}
		resultado[i].Criaturas = append(resultado[i].Criaturas, d.Criatura)
	}

	return resultado
}
