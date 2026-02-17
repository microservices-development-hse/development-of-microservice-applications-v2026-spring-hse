package postgres

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURL string) error {
	// указываем путь к папке с файлами и строку подключения к БД
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return err
	}

	// Применяет все новые миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
