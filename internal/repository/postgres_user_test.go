package repository

import (
	"context"
	"errors"
	"testing"
	"todo_api/internal/models"
)

func Test_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	firstUser := models.User{
		Username:     "Bob",
		PasswordHash: "hello",
	}

	// проверяю что пользователь правильно создаётся, с корректным даннымb
	t.Run("goodTest", func(t *testing.T) {
		user, err := repo.Create(context.Background(), firstUser)
		if err != nil {
			t.Fatalf("user not created: %v", err)
		}

		dbUser := readTestUser(t, db, user.ID)

		if user.ID <= 0 {
			t.Errorf("id not be 0 or negative")
		}

		if user.ID != dbUser.ID {
			t.Errorf("got id: %v, want id: %v", user.ID, dbUser.ID)
		}

		if dbUser.Username != firstUser.Username || dbUser.PasswordHash != firstUser.PasswordHash {
			t.Errorf("got username: %v want username: %v; got password: %v want password: %v",
				dbUser.Username, firstUser.Username, dbUser.PasswordHash, firstUser.PasswordHash)
		}

		if dbUser.CreatedAt.IsZero() {
			t.Errorf("time not be zero")
		}

	})

	secondUser := models.User{
		Username:     "Petr",
		PasswordHash: "123",
	}

	// проверяю попытку создать второго пользователя с таким же именем
	t.Run("usernameTaken", func(t *testing.T) {
		createTestUser(t, db, secondUser.Username)
		_, err := repo.Create(context.Background(), secondUser)
		if !errors.Is(err, ErrUsernameTaken) {
			t.Fatalf("got error: %v, want error: %v", err, ErrUsernameTaken)
		}

		count := countUserInDB(t, db, secondUser.Username)
		if count != 1 {
			t.Errorf("there can only be one user with that name in the database")
		}
	})

	emptyUser := models.User{
		Username:     "",
		PasswordHash: "hello",
	}

	// проверяю что пользователь с пустыми именем не создасться
	t.Run("emptyUsername", func(t *testing.T) {
		_, err := repo.Create(context.Background(), emptyUser)
		if !errors.Is(err, ErrValidation) {
			t.Fatalf("got error: %v, want error: %v", err, ErrValidation)
		}

		count := countUserInDB(t, db, emptyUser.Username)
		if count != 0 {
			t.Errorf("with empty username, database could not be users")
		}
	})

	smallRegister := models.User{
		Username:     "vanya",
		PasswordHash: "test",
	}

	bigRegister := models.User{
		Username:     "Vanya",
		PasswordHash: "test",
	}

	// тест проверяет то что пользователи с одиннаковыми именами, но
	// разными регистрами букв создаются успешно
	t.Run("differentRegisters", func(t *testing.T) {
		smallReg, err := repo.Create(context.Background(), smallRegister)
		if err != nil {
			t.Fatalf("user not created: %v", err)
		}

		bigReg, err := repo.Create(context.Background(), bigRegister)
		if err != nil {
			t.Fatalf("user not created: %v", err)
		}

		smallCount := countUserInDB(t, db, smallReg.Username)
		bigCount := countUserInDB(t, db, bigReg.Username)

		if smallReg.ID == bigReg.ID {
			t.Errorf("different users cannot have the same id")
		}

		if smallCount != 1 || bigCount != 1 {
			t.Errorf("there should be one user with a different name in the database")
		}
	})

}
