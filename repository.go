package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// executor é "qualquer coisa que sabe executar SQL".
// Tanto *sql.DB (conexão normal) quanto *sql.Tx (transação) cumprem esse
// contrato, então as funções abaixo funcionam com os dois.
type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// obterOuCriarCriatura devolve o id da criatura, criando-a se ainda não existir.
//
// O truque "ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)" faz o MySQL
// devolver o id da linha JÁ EXISTENTE quando o nome está repetido. Assim não
// precisamos mais do INSERT IGNORE + SELECT extra do código antigo, e o
// INSERT IGNORE também escondia outros erros de verdade.
//
// Efeito colateral inofensivo: o MySQL "gasta" um número do AUTO_INCREMENT a
// cada tentativa de inserir um nome repetido. Por isso os ids têm saltos
// (podem passar de 10.000 com poucas centenas de linhas). Nada depende deles.
func obterOuCriarCriatura(ctx context.Context, ex executor, nome string) (int64, error) {
	resultado, err := ex.ExecContext(ctx,
		"INSERT INTO creatures (name) VALUES (?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)",
		nome)
	if err != nil {
		return 0, fmt.Errorf("inserindo criatura %q: %w", nome, err)
	}

	id, err := resultado.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("lendo o id da criatura %q: %w", nome, err)
	}
	return id, nil
}

// obterOuCriarItem faz o mesmo que obterOuCriarCriatura, para a tabela weekly_items.
func obterOuCriarItem(ctx context.Context, ex executor, nome string) (int64, error) {
	resultado, err := ex.ExecContext(ctx,
		"INSERT INTO weekly_items (name) VALUES (?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)",
		nome)
	if err != nil {
		return 0, fmt.Errorf("inserindo item %q: %w", nome, err)
	}

	id, err := resultado.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("lendo o id do item %q: %w", nome, err)
	}
	return id, nil
}

// CadastrarCriaturaComDrop registra "esta criatura dropa este item".
// Pode ser chamada várias vezes com os mesmos dados sem duplicar nada.
func CadastrarCriaturaComDrop(ctx context.Context, ex executor, nomeMonstro, nomeItem string) error {
	nomeMonstro = strings.TrimSpace(nomeMonstro)
	nomeItem = strings.TrimSpace(nomeItem)
	if nomeMonstro == "" || nomeItem == "" {
		return errors.New("o nome da criatura e o nome do item não podem ser vazios")
	}

	criaturaID, err := obterOuCriarCriatura(ctx, ex, nomeMonstro)
	if err != nil {
		return err
	}

	itemID, err := obterOuCriarItem(ctx, ex, nomeItem)
	if err != nil {
		return err
	}

	// Liga os dois ids na tabela intermediária. O INSERT IGNORE aqui é seguro:
	// o único "erro" que ele esconde é a ligação já existir (chave duplicada).
	_, err = ex.ExecContext(ctx,
		"INSERT IGNORE INTO creature_drops (creature_id, item_id) VALUES (?, ?)",
		criaturaID, itemID)
	if err != nil {
		return fmt.Errorf("ligando a criatura %q ao item %q: %w", nomeMonstro, nomeItem, err)
	}

	return nil
}

// TaskDrop é uma linha do resultado de ListarTasks: uma criatura e um item que ela dropa.
type TaskDrop struct {
	Criatura string
	Item     string
}

// ListarTasks devolve todos os drops cadastrados, ordenados por criatura.
// Ela só BUSCA os dados; quem chamar decide se imprime, salva ou envia.
func ListarTasks(ctx context.Context, db *sql.DB) ([]TaskDrop, error) {
	linhas, err := db.QueryContext(ctx, `
		SELECT c.name, i.name
		FROM creature_drops cd
		JOIN creatures c ON cd.creature_id = c.id
		JOIN weekly_items i ON cd.item_id = i.id
		ORDER BY c.name;
	`)
	if err != nil {
		return nil, fmt.Errorf("buscando as tasks no banco: %w", err)
	}
	// Libera a conexão e a memória quando a função terminar.
	defer linhas.Close()

	return lerDrops(linhas)
}

// lerDrops percorre as linhas de uma consulta com duas colunas
// (nome da criatura, nome do item) e as converte em TaskDrop.
// É usada por ListarTasks e pelas buscas, para não repetir o laço.
func lerDrops(linhas *sql.Rows) ([]TaskDrop, error) {
	var drops []TaskDrop
	for linhas.Next() {
		var d TaskDrop
		if err := linhas.Scan(&d.Criatura, &d.Item); err != nil {
			return nil, fmt.Errorf("lendo uma linha de drops: %w", err)
		}
		drops = append(drops, d)
	}

	// Se o laço terminou por causa de um erro (e não por acabarem as linhas),
	// ele aparece aqui. Esquecer esta checagem é um erro muito comum.
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("percorrendo as linhas de drops: %w", err)
	}

	return drops, nil
}
