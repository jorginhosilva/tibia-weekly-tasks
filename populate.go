package main

import "database/sql"

func InserirDadosTibia(db *sql.DB) {
	// Adicionando as primeiras criaturas para fins de teste
	CadastrarCriaturaComDrop(db, "Afflicted Strider", "Afflicted Strider Head")
	CadastrarCriaturaComDrop(db, "Afflicted Strider", "Afflicted Strider Worms")
	// agora irei inserir duas criaturas para testes, depois elas irão ser removidas.
	CadastrarCriaturaComDrop(db, "Dragon", "Dragon Shield")
	CadastrarCriaturaComDrop(db, "Dragon Lord", "Dragon Shield")

}