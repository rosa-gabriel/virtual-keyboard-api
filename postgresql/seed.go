package postgresql

import (
	"context"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var schema = `
create table combinations (
    options int[][]
);

create table users (
    username varchar(50) primary key,
    email varchar(50),
    password varchar(100) unique
);

INSERT INTO public.combinations ("options") VALUES
	 ('{{1,2},{3,4},{5,7},{6,8},{9,10}}'),
	 ('{{5,2},{3,4},{1,6},{7,8},{9,10}}'),
	 ('{{1,2},{3,4},{5,9},{7,8},{6,10}}'),
	 ('{{1,2},{3,4},{5,6},{7,8},{9,10}}'),
	 ('{{1,2},{3,5},{4,6},{7,8},{9,10}}');

INSERT INTO public.users (username,email,"password") VALUES
	 ('Among Us','among@us.br','$2a$14$Mzu9bIkdj.NSA6vNjLHl.O4xo5.AagKMnQvgwbhhnSLQ48ji/Dfry'),
	 ('Paulo Sérgio','paulo@kuba.ch','$2a$14$KMqaccLtCYFncwTgMkQ6FueEtoufOkUPmPNQ03UnC2G0fcejzVN2u');
`

func MigrateDB(db *pgxpool.Conn) error {
	_, err := db.Exec(context.Background(), schema)
	slog.Info("Migrating database...")

	if err != nil && !strings.Contains(err.Error(), "42P07") {
		slog.Error("Failed to migrate database", "error", err)
		return err
	}

	slog.Info("Database migrated")

	return nil
}
