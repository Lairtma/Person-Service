package migrations

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// EnsureDatabase создает базу данных, если она не существует
func EnsureDatabase(dbName string) error {
	// Подключаемся к postgres для создания базы данных
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("ошибка подключения к postgres: %v", err)
	}

	// Проверяем существование базы данных
	var exists bool
	err = db.Raw("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = ?)", dbName).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("ошибка проверки существования базы данных: %v", err)
	}

	if !exists {
		// Создаем базу данных
		err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error
		if err != nil {
			return fmt.Errorf("ошибка создания базы данных: %v", err)
		}
		log.Printf("Создана база данных: %s", dbName)
	}

	return nil
}

// RunMigrations выполняет все миграции
func RunMigrations(db *gorm.DB) error {
	// Создаем таблицу миграций, если её нет
	type Migration struct {
		ID        uint   `gorm:"primaryKey"`
		Name      string `gorm:"uniqueIndex"`
		AppliedAt int64  `gorm:"autoCreateTime"`
	}

	err := db.AutoMigrate(&Migration{})
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы миграций: %v", err)
	}

	// Получаем список примененных миграций
	var appliedMigrations []Migration
	if err := db.Find(&appliedMigrations).Error; err != nil {
		return fmt.Errorf("ошибка получения списка миграций: %v", err)
	}

	applied := make(map[string]bool)
	for _, m := range appliedMigrations {
		applied[m.Name] = true
	}

	// Определяем миграции
	migrations := []struct {
		name string
		fn   func(*gorm.DB) error
	}{
		{
			name: "001_init",
			fn: func(db *gorm.DB) error {
				type Person struct {
					ID         uint   `gorm:"primaryKey"`
					CreatedAt  int64  `gorm:"autoCreateTime"`
					UpdatedAt  int64  `gorm:"autoUpdateTime"`
					DeletedAt  *int64 `gorm:"index"`
					Name       string `gorm:"not null"`
					Surname    string `gorm:"not null"`
					Patronymic string
					Age        int
					Gender     string
					Country    string
				}

				return db.AutoMigrate(&Person{})
			},
		},
	}

	// Применяем миграции
	for _, m := range migrations {
		if applied[m.name] {
			continue
		}

		// Начинаем транзакцию
		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("ошибка начала транзакции для миграции %s: %v", m.name, tx.Error)
		}

		// Выполняем миграцию
		if err := m.fn(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("ошибка выполнения миграции %s: %v", m.name, err)
		}

		// Записываем информацию о примененной миграции
		if err := tx.Create(&Migration{Name: m.name}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("ошибка записи информации о миграции %s: %v", m.name, err)
		}

		// Подтверждаем транзакцию
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("ошибка подтверждения миграции %s: %v", m.name, err)
		}

		log.Printf("Применена миграция: %s", m.name)
	}

	return nil
}
