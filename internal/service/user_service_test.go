package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"todo_api/internal/models"
	"todo_api/internal/repository"
)

// === User repository ===

type fakeUserRepo struct {
	CreateFuncUser    func(ctx context.Context, u models.User) (models.User, error)
	GetByUsernameFunc func(ctx context.Context, username string) (models.User, error)
}

func (f *fakeUserRepo) Create(ctx context.Context, u models.User) (models.User, error) {
	return f.CreateFuncUser(ctx, u)
}

func (f *fakeUserRepo) GetByUsername(ctx context.Context, username string) (models.User, error) {
	return f.GetByUsernameFunc(ctx, username)
}

// Test_Concurrency - воспроизводит гонку в register.
// 5 горутин одновременно пытаются зарегистрироваться c одним и тем же именем.
// Исправный код должен показать одну успешную регистрацию.
func Test_Concurrency(t *testing.T) {
	usernameTest := "test_user"
	password := "test"
	secret := []byte("test-secret")

	users := make(map[string]models.User)

	// wg - главная горутина теста ждёт, пока все пять Register завершатся,
	// и только потом читает счётчик.
	var wg sync.WaitGroup
	// ready: барьер внутри фейка. Не отпускает ни одну горутину из
	// GetByUsername, пока все пять не получили результат проверки.
	// Без него быстрая горутина успевает и проверить, и создать пользователя
	// до того, как остальные дойдут до проверки, и гонка не проявляется.
	var ready sync.WaitGroup
	var rmu sync.RWMutex
	// CountRegUsers - cчётчик успешных регистраций. Атомарный, потому что его пишут пять
	// горутин одновременно: обычный int терял бы инкременты.
	var countRegUsers atomic.Int64

	ready.Add(5)

	repo := &fakeUserRepo{
		CreateFuncUser: func(ctx context.Context, u models.User) (models.User, error) {
			rmu.Lock()
			defer rmu.Unlock()
			_, ok := users[u.Username]
			if ok {
				return models.User{}, repository.ErrUsernameTaken
			}
			users[u.Username] = u
			return u, nil
		},

		GetByUsernameFunc: func(ctx context.Context, username string) (models.User, error) {
			rmu.RLock()
			v, ok := users[username]
			rmu.RUnlock()
			ready.Done()
			ready.Wait()
			if !ok {
				return models.User{}, repository.ErrUserNotFound
			}
			return v, nil
		},
	}

	svc := NewUserService(repo, secret)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Register(context.Background(), usernameTest, password)
			if err != nil {
				if !errors.Is(err, ErrUsernameIsTaken) {
					t.Errorf("got: %v, want: %v", err, ErrUsernameIsTaken)
				}
			} else {
				countRegUsers.Add(1)
			}
		}()
	}

	wg.Wait()

	count := countRegUsers.Load()

	if count != 1 {
		t.Errorf("got: %d, want: %d", count, 1)
	}
}
