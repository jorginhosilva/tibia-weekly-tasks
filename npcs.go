package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Este arquivo carrega quanto cada NPC paga pelos itens da weekly.
//
// Os dados ficam em dados/npcs, um arquivo .csv por NPC (rashid.csv,
// yasir.csv...), preenchidos à mão neste formato:
//
//	item;preco
//	Chaos Mace;9000
//
// O nome do NPC vem do nome do arquivo. O nome do item é guardado como TEXTO,
// sem exigir que ele exista na lista da weekly (populate.go): assim um nome
// escrito diferente não trava a carga. Esses nomes aparecem como AVISO, para
// serem corrigidos depois.
//
// Já um erro de FORMATO (preço com ponto, linha sem ";", item repetido) trava
// a carga inteira, porque gravaria dado quebrado no banco.

// pastaNpcs é onde ficam os arquivos dos NPCs.
const pastaNpcs = "dados/npcs"

// precoMaximo é o maior preço aceito. Serve para pegar dígitos a mais
// digitados por engano (e cabe com folga na coluna INT UNSIGNED).
const precoMaximo = 100_000_000

// cabecalhoNpc é a primeira linha obrigatória de cada arquivo.
const cabecalhoNpc = "item;preco"

// precoNpc é uma linha do arquivo: um item e quanto o NPC paga por ele.
type precoNpc struct {
	Item  string
	Preco int
}

// listaNpc é um arquivo inteiro: o NPC e todos os seus preços.
type listaNpc struct {
	Npc    string
	Precos []precoNpc
}

// ResultadoCargaNpc resume o que a carga fez, para o terminal mostrar.
type ResultadoCargaNpc struct {
	SemArquivos bool     // a pasta não existe ou não tem .csv: nada foi alterado no banco
	Npcs        int      // NPCs gravados
	Precos      int      // preços gravados
	Avisos      []string // nomes fora da lista da weekly, arquivos sem itens
}

// CarregarPrecosNpc lê a pasta dos NPCs e deixa o banco IGUAL a ela: apaga os
// preços antigos e grava os da pasta, numa transação. Se algo falhar no meio,
// o Rollback devolve o banco ao estado anterior.
//
// Se a pasta não existir ou estiver vazia, NÃO mexe no banco: isso protege os
// preços já gravados quando o programa é rodado de outra pasta por engano.
func CarregarPrecosNpc(ctx context.Context, db *sql.DB, pasta string) (ResultadoCargaNpc, error) {
	listas, err := lerListasNpc(pasta)
	if err != nil {
		return ResultadoCargaNpc{}, err
	}
	if len(listas) == 0 {
		return ResultadoCargaNpc{SemArquivos: true}, nil
	}

	_, itensWeekly := nomesDistintos(dropsIniciais)
	resultado := ResultadoCargaNpc{Avisos: avisosNpc(listas, itensWeekly)}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return ResultadoCargaNpc{}, fmt.Errorf("iniciando a transação dos NPCs: %w", err)
	}
	defer tx.Rollback() // depois do Commit, não faz nada

	if err := gravarPrecosNpc(ctx, tx, listas); err != nil {
		return ResultadoCargaNpc{}, err
	}
	if err := tx.Commit(); err != nil {
		return ResultadoCargaNpc{}, fmt.Errorf("confirmando os preços dos NPCs (commit): %w", err)
	}

	for _, l := range listas {
		if len(l.Precos) > 0 {
			resultado.Npcs++
			resultado.Precos += len(l.Precos)
		}
	}
	return resultado, nil
}

// lerListasNpc lê todos os .csv da pasta (em ordem alfabética) e confere o
// formato. Devolve TODOS os problemas de uma vez, para corrigir tudo numa
// rodada só, em vez de um erro por vez.
func lerListasNpc(pasta string) ([]listaNpc, error) {
	arquivos, err := filepath.Glob(filepath.Join(pasta, "*.csv"))
	if err != nil {
		return nil, fmt.Errorf("procurando os arquivos de NPC em %s: %w", pasta, err)
	}
	sort.Strings(arquivos)

	var listas []listaNpc
	var problemas []string
	vistos := map[string]string{} // NPC em minúsculas -> arquivo onde apareceu
	for _, arq := range arquivos {
		base := filepath.Base(arq)
		npc := nomeDoNpc(base)
		if outro, repetido := vistos[strings.ToLower(npc)]; repetido {
			problemas = append(problemas, fmt.Sprintf("%s e %s são do mesmo NPC (%s): deixe só um", outro, base, npc))
			continue
		}
		vistos[strings.ToLower(npc)] = base

		f, err := os.Open(arq)
		if err != nil {
			return nil, fmt.Errorf("abrindo %s: %w", arq, err)
		}
		precos, probs := lerCsvNpc(f, base)
		f.Close()

		problemas = append(problemas, probs...)
		listas = append(listas, listaNpc{Npc: npc, Precos: precos})
	}

	if len(problemas) > 0 {
		return nil, fmt.Errorf("os arquivos de NPC têm %d problema(s) e NADA foi gravado:\n  - %s",
			len(problemas), strings.Join(problemas, "\n  - "))
	}
	return listas, nil
}

// lerCsvNpc lê um arquivo de NPC e devolve os preços e os problemas de formato.
// "arquivo" é só o nome usado nas mensagens.
func lerCsvNpc(r io.Reader, arquivo string) ([]precoNpc, []string) {
	leitor := csv.NewReader(r)
	leitor.Comma = ';'
	leitor.FieldsPerRecord = -1 // o número de colunas é conferido abaixo, com uma mensagem clara

	var precos []precoNpc
	var problemas []string
	add := func(linha int, formato string, args ...any) {
		problemas = append(problemas, fmt.Sprintf("%s, linha %d: ", arquivo, linha)+fmt.Sprintf(formato, args...))
	}

	vistos := map[string]int{} // item em minúsculas -> linha onde apareceu
	cabecalhoLido := false
	for {
		registro, err := leitor.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Erro de leitura (por exemplo, aspas soltas): as linhas seguintes não
			// são confiáveis, então para neste arquivo.
			var errCsv *csv.ParseError
			if errors.As(err, &errCsv) {
				add(errCsv.Line, "não foi possível ler a linha (%v)", errCsv.Err)
			} else {
				add(0, "erro lendo o arquivo: %v", err)
			}
			return nil, problemas
		}
		linha, _ := leitor.FieldPos(0)

		if !cabecalhoLido {
			cabecalhoLido = true
			// O Excel às vezes grava um caractere invisível (BOM) no começo do arquivo.
			registro[0] = strings.TrimPrefix(registro[0], string(rune(0xFEFF)))
			campos := make([]string, len(registro))
			for i, c := range registro {
				campos[i] = strings.ToLower(strings.TrimSpace(c))
			}
			if strings.Join(campos, ";") != cabecalhoNpc {
				add(linha, "a primeira linha deve ser %q, mas é %q", cabecalhoNpc, strings.Join(registro, ";"))
				return nil, problemas
			}
			continue
		}

		if len(registro) != 2 {
			add(linha, "esperava 2 colunas (item;preco), mas tem %d: %q", len(registro), strings.Join(registro, ";"))
			continue
		}
		// Espaços nas pontas não mudam o sentido ("Chaos Mace; 9000"): são ignorados.
		item := strings.TrimSpace(registro[0])
		textoPreco := strings.TrimSpace(registro[1])

		switch {
		case item == "":
			add(linha, "o nome do item está vazio")
			continue
		case strings.Contains(item, "  "):
			add(linha, "o item %q tem dois espaços seguidos", item)
			continue
		case utf8.RuneCountInString(item) > tamanhoMaxNome:
			add(linha, "o item %q passa de %d caracteres", item, tamanhoMaxNome)
			continue
		}

		preco, ok := lerPreco(textoPreco)
		switch {
		case !ok:
			add(linha, "o preço %q de %q deve ter só números, sem ponto nem vírgula (ex.: 9000)", textoPreco, item)
			continue
		case preco == 0:
			add(linha, "o preço de %q é zero", item)
			continue
		case preco > precoMaximo:
			add(linha, "o preço de %q (%d) é alto demais: confira se não sobrou um dígito", item, preco)
			continue
		}

		chave := strings.ToLower(item)
		if anterior, repetido := vistos[chave]; repetido {
			add(linha, "o item %q está repetido (já aparece na linha %d)", item, anterior)
			continue
		}
		vistos[chave] = linha
		precos = append(precos, precoNpc{Item: item, Preco: preco})
	}
	return precos, problemas
}

// lerPreco aceita só dígitos: "9000" vale; "9.000", "9,000", "-5" e "+5" não.
func lerPreco(texto string) (int, bool) {
	if texto == "" || len(texto) > 12 {
		return 0, false
	}
	for _, c := range texto {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(texto)
	return n, err == nil
}

// nomeDoNpc transforma o nome do arquivo no nome do NPC:
// "rashid.csv" -> "Rashid"; "esrik_pompan.csv" -> "Esrik Pompan".
func nomeDoNpc(arquivo string) string {
	base := strings.TrimSuffix(arquivo, filepath.Ext(arquivo))
	palavras := strings.FieldsFunc(base, func(c rune) bool { return c == '_' || c == '-' || c == ' ' })
	for i, p := range palavras {
		r, tamanho := utf8.DecodeRuneInString(p)
		palavras[i] = string(unicode.ToUpper(r)) + p[tamanho:]
	}
	return strings.Join(palavras, " ")
}

// avisosNpc lista o que não impede a carga, mas merece atenção: itens cujo
// nome não está na lista da weekly e arquivos sem nenhum item.
func avisosNpc(listas []listaNpc, itensWeekly []string) []string {
	weekly := make(map[string]bool, len(itensWeekly))
	for _, it := range itensWeekly {
		weekly[strings.ToLower(it)] = true
	}

	var avisos []string
	for _, l := range listas {
		if len(l.Precos) == 0 {
			avisos = append(avisos, fmt.Sprintf("%s: o arquivo não tem itens (o NPC não aparece no site)", l.Npc))
			continue
		}
		for _, p := range l.Precos {
			if !weekly[strings.ToLower(p.Item)] {
				avisos = append(avisos, fmt.Sprintf("%s: %q não está na lista da weekly", l.Npc, p.Item))
			}
		}
	}
	return avisos
}

// gravarPrecosNpc apaga os preços antigos e grava os novos. Roda dentro da
// transação aberta por quem chamou. NPCs sem itens não são gravados.
func gravarPrecosNpc(ctx context.Context, ex executorDeLotes, listas []listaNpc) error {
	// A ordem importa por causa da chave estrangeira: primeiro os preços.
	for _, tabela := range []string{"npc_prices", "npcs"} {
		if _, err := ex.ExecContext(ctx, "DELETE FROM "+tabela); err != nil {
			return fmt.Errorf("limpando a tabela %s: %w", tabela, err)
		}
	}

	for _, l := range listas {
		if len(l.Precos) == 0 {
			continue
		}
		res, err := ex.ExecContext(ctx, "INSERT INTO npcs (name) VALUES (?)", l.Npc)
		if err != nil {
			return fmt.Errorf("gravando o NPC %s: %w", l.Npc, err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("lendo o id do NPC %s: %w", l.Npc, err)
		}

		for inicio := 0; inicio < len(l.Precos); inicio += tamanhoLote {
			fim := min(inicio+tamanhoLote, len(l.Precos))
			args := make([]any, 0, (fim-inicio)*3)
			for _, p := range l.Precos[inicio:fim] {
				args = append(args, id, p.Item, p.Preco)
			}
			consulta := "INSERT INTO npc_prices (npc_id, item_name, price) VALUES " + marcadores(fim-inicio, 3)
			if _, err := ex.ExecContext(ctx, consulta, args...); err != nil {
				return fmt.Errorf("gravando os preços do NPC %s (linhas %d a %d): %w", l.Npc, inicio+1, fim, err)
			}
		}
	}
	return nil
}
