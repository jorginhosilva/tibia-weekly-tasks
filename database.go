package main

import (
	"context"
	"database/sql"
	"fmt"
)

// comandoTabela guarda o nome de uma tabela e o SQL que a cria.
// Ter o nome separado ajuda a montar mensagens de erro claras.
type comandoTabela struct {
	nome string
	sql  string
}

// tabelas está na ORDEM CERTA: creature_drops depende de creatures e
// weekly_items (chaves estrangeiras), então ela precisa vir por último.
var tabelas = []comandoTabela{
	{
		nome: "creatures",
		sql:  `CREATE TABLE IF NOT EXISTS creatures (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(100) NOT NULL UNIQUE);`,
	},
	{
		nome: "weekly_items",
		sql:  `CREATE TABLE IF NOT EXISTS weekly_items (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(100) NOT NULL UNIQUE);`,
	},
	{
		nome: "creature_drops",
		sql: `CREATE TABLE IF NOT EXISTS creature_drops (creature_id INT, item_id INT,
			PRIMARY KEY (creature_id, item_id),
			FOREIGN KEY (creature_id) REFERENCES creatures(id) ON DELETE CASCADE,
			FOREIGN KEY (item_id) REFERENCES weekly_items(id) ON DELETE CASCADE);`,
	},
}

// CriarTabelas cria as tabelas que ainda não existem.
// Se qualquer criação falhar, PARA na hora e devolve o erro (antes o código
// continuava e ainda imprimia "criada com sucesso" mesmo com falha).
func CriarTabelas(ctx context.Context, db *sql.DB) error {
	for _, t := range tabelas {
		if _, err := db.ExecContext(ctx, t.sql); err != nil {
			// %w "embrulha" o erro original dentro do nosso, sem perder a causa.
			return fmt.Errorf("criando a tabela %s: %w", t.nome, err)
		}
	}
	return nil
}
