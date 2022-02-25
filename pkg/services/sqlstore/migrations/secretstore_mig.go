package migrations

import . "github.com/grafana/grafana/pkg/services/sqlstore/migrator"

func addSecretStoreMigration(mg *Migrator) {
	// new table
	var tableV1 = Table{
		Name: "secret",
		Columns: []*Column{
			{Name: "id", Type: DB_BigInt, IsPrimaryKey: true, IsAutoIncrement: true},
			{Name: "org_id", Type: DB_BigInt, Nullable: false},
			{Name: "entity_uid", Type: DB_NVarchar, Length: 40, Nullable: false, Default: "0"},
			{Name: "secure_json_data", Type: DB_Text, Nullable: true},
			{Name: "created", Type: DB_DateTime, Nullable: false},
			{Name: "updated", Type: DB_DateTime, Nullable: false},
		},
		Indices: []*Index{
			{Cols: []string{"org_id"}},
			{Cols: []string{"org_id", "entity_uid"}, Type: UniqueIndex},
		},
	}

	// create table
	mg.AddMigration("create secrets table v1", NewAddTableMigration(tableV1))

	// add indíces
	addTableIndicesMigrations(mg, "v1", tableV1)
}
