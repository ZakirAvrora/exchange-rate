package app

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	// migrate tools
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func initMigrate(dbURL string) error {

	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return fmt.Errorf("migrate error: postgres connect error: %w", err)
	}

	err = m.Up()
	defer m.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate error: up error: %w", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		// NoReturnErr: migration already implemented
		log.Println("Migrate: no change")
	}

	return nil
}
