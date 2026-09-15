package main

import (
	"database/sql"
	"fmt"
)

func CadastrarCriaturaComDrop(db *sql.DB, nomeMonstro string, nomeItem string) {
	// insert ignore faz com que o monstro não seja duplicado, se ele já existir
	resultCreature, err := db.Exec("INSERT IGNORE INTO creatures (name) VALUES (?)", nomeMonstro)
	if err != nil {
		fmt.Println("Erro ao tentar inserir criatura", err)
		return
	}

	criaturaID, err := resultCreature.LastInsertId()
	if err != nil {
		fmt.Println("Erro ao capturar LastInsertId da criatura:", err)
		return
	}

	if criaturaID == 0 {
		// se o ID veio zero, o monstro já existi, dai busca o ID real dele no banco.
		// db.QueryRow algo que quer buscar apenas numa linha única.
		err = db.QueryRow("SELECT  id FROM creatures WHERE name = ?", nomeMonstro).Scan(&criaturaID)
		if err != nil {
			fmt.Println("Erro ao buscar ID da criatura existente:", err)
			return
		}
	}

	// ITEM DE LOOT
	resultItem, err := db.Exec("INSERT IGNORE INTO weekly_items (name) VALUES (?)", nomeItem)
	if err != nil {
		fmt.Println("Erro ao tentar inserir item:", err)
		return
	}

	itemID, err := resultItem.LastInsertId()
	if err != nil {
		fmt.Println("Erro ao capturar LastInsertId do item:", err)
		return
	}

	// se o ID veio zero, o item já existi, dai busca o ID real dele no banco.
	if itemID == 0 {
		// db.QueryRow algo que quer buscar apenas numa linha única.
		err = db.QueryRow("SELECT id FROM weekly_items WHERE name = ?",nomeItem).Scan(&itemID)
		if err != nil {
			fmt.Println("Erro ao buscar ID do item existente:", err)
			return
		}
	}

	// cruzar os dois id's na tabela que faz a intermediação.
	_, err = db.Exec("INSERT IGNORE INTO creature_drops (creature_id, item_id) VALUES (?, ?)", criaturaID, itemID)
	if err != nil {
		fmt.Println("Erro ao buscar ID do item existente:", err)
		return
	}

	fmt.Printf("Criatura: [%s] dropa [%s]\n", nomeMonstro, nomeItem)
}

func ListarTasksELeads() {
	
}