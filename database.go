package main

import (
	"database/sql"
	"fmt"
)

func CriarTabelas(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS creatures (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(100) NOT NULL UNIQUE);")
	if err != nil {
		fmt.Println("Erro ao criar tabela:", err)
		return
	}
	fmt.Println("Tabela criada com sucesso!")

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS weekly_items (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(100) NOT NULL UNIQUE);")
	if err != nil {
		fmt.Println("Erro ao criar tabela:", err)
	}
	fmt.Println("Tabela criada com Sucesso!")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS creature_drops (creature_id INT, item_id INT,
	PRIMARY KEY (creature_id, item_id),
	FOREIGN KEY (creature_id) REFERENCES creatures(id) ON DELETE CASCADE,
	FOREIGN KEY (item_id) REFERENCES weekly_items(id) ON DELETE CASCADE);`)
	if err != nil {
		fmt.Println("Erro ao criar tabela:", err)
	}
	fmt.Println("Tabela criada com Sucesso!")
}