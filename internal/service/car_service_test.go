package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/brandon-anubis/cars-api/internal/domain"
	"github.com/brandon-anubis/cars-api/internal/repository"
	"github.com/brandon-anubis/cars-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCarRepo implements repository.CarRepository with configurable function fields.
// Any method whose fn field is nil returns a zero-value response without panicking.
type mockCarRepo struct {
	createFn  func(ctx context.Context, car *domain.Car) error
	getByIDFn func(ctx context.Context, id string) (*domain.Car, error)
	listFn    func(ctx context.Context, limit, offset int) ([]*domain.Car, int, error)
	updateFn  func(ctx context.Context, car *domain.Car) error
	deleteFn  func(ctx context.Context, id string) error
}

var _ repository.CarRepository = (*mockCarRepo)(nil)

func (m *mockCarRepo) Create(ctx context.Context, car *domain.Car) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, car)
}

func (m *mockCarRepo) GetByID(ctx context.Context, id string) (*domain.Car, error) {
	if m.getByIDFn == nil {
		return nil, nil
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockCarRepo) List(ctx context.Context, limit, offset int) ([]*domain.Car, int, error) {
	if m.listFn == nil {
		return nil, 0, nil
	}
	return m.listFn(ctx, limit, offset)
}

func (m *mockCarRepo) Update(ctx context.Context, car *domain.Car) error {
	if m.updateFn == nil {
		return nil
	}
	return m.updateFn(ctx, car)
}

func (m *mockCarRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, id)
}

// silentLogger returns a logger that discards all output, suitable for tests.
func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// validUpdateInput returns a fully populated CarUpdateInput with no validation errors.
func validUpdateInput() domain.CarUpdateInput {
	return domain.CarUpdateInput{
		Make:       "Toyota",
		Model:      "Camry",
		Year:       2020,
		Color:      "Blue",
		VIN:        "1HGCM82633A123456",
		Mileage:    10000,
		PriceCents: 2500000,
	}
}

// TestCarService_Create covers all Create paths.
func TestCarService_Create(t *testing.T) {
	ctx := context.Background()
	errRepo := errors.New("repo error")

	tests := []struct {
		name     string
		carMake  string
		model    string
		year     int
		color    string
		createFn func(ctx context.Context, car *domain.Car) error
		wantErr  bool
		checkErr func(t *testing.T, err error)
		checkCar func(t *testing.T, car *domain.Car)
	}{
		{
			name:    "valid input",
			carMake: "Honda", model: "Civic", year: 2022, color: "Red",
			wantErr: false,
			checkCar: func(t *testing.T, car *domain.Car) {
				t.Helper()
				assert.NotEmpty(t, car.ID)
				assert.Equal(t, "Honda", car.Make)
				assert.Equal(t, "Civic", car.Model)
			},
		},
		{
			name:    "domain validation error empty make",
			carMake: "", model: "Civic", year: 2022, color: "Red",
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, domain.ErrCarValidation)
			},
		},
		{
			name:    "repo returns ErrCarAlreadyExists",
			carMake: "Honda", model: "Civic", year: 2022, color: "Red",
			createFn: func(_ context.Context, _ *domain.Car) error { return domain.ErrCarAlreadyExists },
			wantErr:  true,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, domain.ErrCarAlreadyExists)
			},
		},
		{
			name:    "repo returns generic error",
			carMake: "Honda", model: "Civic", year: 2022, color: "Red",
			createFn: func(_ context.Context, _ *domain.Car) error { return errRepo },
			wantErr:  true,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, errRepo)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCarRepo{createFn: tc.createFn}
			svc := service.NewCarService(repo, silentLogger())

			car, err := svc.Create(ctx, tc.carMake, tc.model, tc.year, tc.color)
			if tc.wantErr {
				require.Error(t, err)
				if tc.checkErr != nil {
					tc.checkErr(t, err)
				}
				assert.Nil(t, car)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, car)
			if tc.checkCar != nil {
				tc.checkCar(t, car)
			}
		})
	}
}

// TestCarService_Create_cancelledContext verifies that a cancelled context
// is detected before domain or repository work begins.
func TestCarService_Create_cancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc := service.NewCarService(&mockCarRepo{}, silentLogger())
	car, err := svc.Create(ctx, "Honda", "Civic", 2022, "Red")

	require.Error(t, err)
	assert.Nil(t, car)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestCarService_GetByID covers all GetByID paths.
func TestCarService_GetByID(t *testing.T) {
	ctx := context.Background()
	errRepo := errors.New("repo error")

	tests := []struct {
		name      string
		id        string
		getByIDFn func(ctx context.Context, id string) (*domain.Car, error)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name: "valid ID",
			id:   "abc-123",
			getByIDFn: func(_ context.Context, _ string) (*domain.Car, error) {
				car, err := domain.NewCar("Toyota", "Camry", 2020, "Blue")
				return car, err
			},
		},
		{
			name:    "empty ID",
			id:      "",
			wantErr: true,
			getByIDFn: func(_ context.Context, _ string) (*domain.Car, error) {
				t.Fatal("repo.GetByID must not be called with empty ID")
				return nil, nil
			},
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, domain.ErrCarValidation)
				var valErr *domain.ValidationError
				require.ErrorAs(t, err, &valErr)
				assert.Equal(t, "id", valErr.Field)
			},
		},
		{
			name:    "whitespace-only ID",
			id:      "   ",
			wantErr: true,
			getByIDFn: func(_ context.Context, _ string) (*domain.Car, error) {
				t.Fatal("repo.GetByID must not be called with whitespace-only ID")
				return nil, nil
			},
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, domain.ErrCarValidation)
			},
		},
		{
			name:      "repo returns ErrCarNotFound",
			id:        "missing-id",
			getByIDFn: func(_ context.Context, _ string) (*domain.Car, error) { return nil, domain.ErrCarNotFound },
			wantErr:   true,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, domain.ErrCarNotFound)
			},
		},
		{
			name:      "repo returns generic error",
			id:        "some-id",
			getByIDFn: func(_ context.Context, _ string) (*domain.Car, error) { return nil, errRepo },
			wantErr:   true,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, errRepo)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCarRepo{getByIDFn: tc.getByIDFn}
			svc := service.NewCarService(repo, silentLogger())

			car, err := svc.GetByID(ctx, tc.id)
			if tc.wantErr {
				require.Error(t, err)
				if tc.checkErr != nil {
					tc.checkErr(t, err)
				}
				assert.Nil(t, car)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, car)
		})
	}
}

// TestCarService_GetByID_cancelledContext verifies that a cancelled context
// is detected before ID validation or repository work begins.
func TestCarService_GetByID_cancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc := service.NewCarService(&mockCarRepo{}, silentLogger())
	car, err := svc.GetByID(ctx, "some-id")

	require.Error(t, err)
	assert.Nil(t, car)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestCarService_List covers limit clamping, offset flooring, and error propagation.
func TestCarService_List(t *testing.T) {
	ctx := context.Background()
	errRepo := errors.New("repo error")

	tests := []struct {
		name       string
		limit      int
		offset     int
		wantLimit  int
		wantOffset int
		repoErr    error
		wantErr    bool
		listFn     func(ctx context.Context, limit, offset int) ([]*domain.Car, int, error)
		checkCars  func(t *testing.T, cars []*domain.Car, total int)
	}{
		{name: "valid limit and offset", limit: 10, offset: 5, wantLimit: 10, wantOffset: 5},
		{name: "limit 0 clamped to 1", limit: 0, offset: 0, wantLimit: 1, wantOffset: 0},
		{name: "limit 101 clamped to 100", limit: 101, offset: 0, wantLimit: 100, wantOffset: 0},
		{name: "limit negative clamped to 1", limit: -1, offset: 0, wantLimit: 1, wantOffset: 0},
		{name: "offset negative floored to 0", limit: 10, offset: -5, wantLimit: 10, wantOffset: 0},
		{name: "repo error propagates", limit: 10, offset: 0, wantLimit: 10, wantOffset: 0, repoErr: errRepo, wantErr: true},
		{
			name:       "returns cars and total from repo",
			limit:      10,
			offset:     0,
			wantLimit:  10,
			wantOffset: 0,
			listFn: func(_ context.Context, limit, offset int) ([]*domain.Car, int, error) {
				car1, _ := domain.NewCar("Toyota", "Camry", 2020, "Blue")
				car2, _ := domain.NewCar("Honda", "Civic", 2021, "Red")
				return []*domain.Car{car1, car2}, 5, nil
			},
			checkCars: func(t *testing.T, cars []*domain.Car, total int) {
				t.Helper()
				assert.Len(t, cars, 2)
				assert.Equal(t, 5, total)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotLimit, gotOffset int
			repo := &mockCarRepo{}
			if tc.listFn != nil {
				repo.listFn = tc.listFn
			} else {
				repo.listFn = func(_ context.Context, limit, offset int) ([]*domain.Car, int, error) {
					gotLimit = limit
					gotOffset = offset
					if tc.repoErr != nil {
						return nil, 0, tc.repoErr
					}
					return []*domain.Car{}, 0, nil
				}
			}
			svc := service.NewCarService(repo, silentLogger())

			cars, total, err := svc.List(ctx, tc.limit, tc.offset)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.repoErr)
				assert.Nil(t, cars)
				assert.Equal(t, 0, total)
				return
			}
			require.NoError(t, err)
			if tc.listFn == nil {
				assert.Equal(t, tc.wantLimit, gotLimit)
				assert.Equal(t, tc.wantOffset, gotOffset)
			}
			if tc.checkCars != nil {
				tc.checkCars(t, cars, total)
			}
		})
	}
}

// TestCarService_List_cancelledContext verifies that a cancelled context
// is detected before clamping or repository work begins.
func TestCarService_List_cancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc := service.NewCarService(&mockCarRepo{}, silentLogger())
	cars, total, err := svc.List(ctx, 10, 0)

	require.Error(t, err)
	assert.Nil(t, cars)
	assert.Equal(t, 0, total)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestCarService_Update covers all Update paths.
func TestCarService_Update(t *testing.T) {
	ctx := context.Background()
	errRepo := errors.New("repo error")

	tests := []struct {
		name      string
		id        string
		input     domain.CarUpdateInput
		getByIDFn func(ctx context.Context, id string) (*domain.Car, error)
		updateFn  func(ctx context.Context, car *domain.Car) error
		wantErr   bool
		checkErr  func(t *testing.T, err error)
		checkCar  func(t *testing.T, car *domain.Car)
	}{
		{
			name:  "valid update",
			id:    "some-id",
			input: validUpdateInput(),
			getByIDFn: func(_ context.Context, _ string) (*domain.Car, error) {
				car, err := domain.NewCar("Honda", "Civic", 2019, "Red")
				return car, err
			},
			checkCar: func(t *testing.T, car *domain.Car) {
				t.Helper()
				assert.Equal(t, "Toyota", car.Make)
				assert.Equal(t, "Camry", car.Model)
				assert.Equal(t, 2020, car.Year)
				assert.Equal(t, "Blue", car.Color)
				assert.Equal(t, "1HGCM82633A123456", car.VIN)
				assert.Equal(t, int64(10000), car.Mileage)
				assert.Equal(t, int64(2500000), car.PriceCents)
			},
		},
		{
			name:    "empty ID",
			id:      "",
			input:   validUpdateInput(),
			wantErr: true,
			getByIDFn: func(_ context.Context, _ string) (*domain.Car, error) {
				t.Fatal("repo.GetByID must not be called with empty ID")
				return nil, nil
			},
			updateFn: func(_ context.Context, _ *domain.Car) error {
				t.Fatal("repo.Update must not be called with empty ID")
				return nil
			},
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, domain.ErrCarValidation)
				var valErr *domain.ValidationError
				require.ErrorAs(t, err, &valErr)
				assert.Equal(t, "id", valErr.Field)
			},
		},
		{
			name:  "repo.GetByID returns ErrCarNotFound",
			id:    "missing-id",
			input: validUpdateInput(),
			getByIDFn: func(_ context.Context, _ string) (*domain.Car, error) {
				return nil, domain.ErrCarNotFound
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, domain.ErrCarNotFound)
			},
		},
		{
			name: "car.Update validation error year 3000",
			id:   "some-id",
			input: func() domain.CarUpdateInput {
				inp := validUpdateInput()
				inp.Year = 3000
				return inp
			}(),
			getByIDFn: func(_ context.Context, _ string) (*domain.Car, error) {
				car, err := domain.NewCar("Toyota", "Camry", 2020, "Blue")
				return car, err
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, domain.ErrCarValidation)
			},
		},
		{
			name:  "repo.Update fails",
			id:    "some-id",
			input: validUpdateInput(),
			getByIDFn: func(_ context.Context, _ string) (*domain.Car, error) {
				car, err := domain.NewCar("Toyota", "Camry", 2020, "Blue")
				return car, err
			},
			updateFn: func(_ context.Context, _ *domain.Car) error { return errRepo },
			wantErr:  true,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, errRepo)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCarRepo{
				getByIDFn: tc.getByIDFn,
				updateFn:  tc.updateFn,
			}
			svc := service.NewCarService(repo, silentLogger())

			car, err := svc.Update(ctx, tc.id, tc.input)
			if tc.wantErr {
				require.Error(t, err)
				if tc.checkErr != nil {
					tc.checkErr(t, err)
				}
				assert.Nil(t, car)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, car)
			if tc.checkCar != nil {
				tc.checkCar(t, car)
			}
		})
	}
}

// TestCarService_Update_cancelledContextBetweenReadAndWrite verifies that a context
// cancelled after GetByID causes the service to abort before calling repo.Update.
func TestCarService_Update_cancelledContextBetweenReadAndWrite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	updateCalled := false

	repo := &mockCarRepo{
		getByIDFn: func(_ context.Context, _ string) (*domain.Car, error) {
			cancel()
			car, err := domain.NewCar("Toyota", "Camry", 2020, "Blue")
			return car, err
		},
		updateFn: func(_ context.Context, _ *domain.Car) error {
			updateCalled = true
			return nil
		},
	}
	svc := service.NewCarService(repo, silentLogger())

	car, err := svc.Update(ctx, "some-id", validUpdateInput())
	require.Error(t, err)
	assert.Nil(t, car)
	assert.ErrorIs(t, err, context.Canceled)
	assert.False(t, updateCalled, "repo.Update must not be called after context cancellation")
}

// TestCarService_Delete covers all Delete paths.
func TestCarService_Delete(t *testing.T) {
	ctx := context.Background()
	errRepo := errors.New("repo error")

	tests := []struct {
		name     string
		id       string
		deleteFn func(ctx context.Context, id string) error
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name:     "valid ID",
			id:       "some-id",
			deleteFn: func(_ context.Context, _ string) error { return nil },
		},
		{
			name:    "empty ID",
			id:      "",
			wantErr: true,
			deleteFn: func(_ context.Context, _ string) error {
				t.Fatal("repo.Delete must not be called with empty ID")
				return nil
			},
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, domain.ErrCarValidation)
				var valErr *domain.ValidationError
				require.ErrorAs(t, err, &valErr)
				assert.Equal(t, "id", valErr.Field)
			},
		},
		{
			name:     "repo returns ErrCarNotFound",
			id:       "missing-id",
			deleteFn: func(_ context.Context, _ string) error { return domain.ErrCarNotFound },
			wantErr:  true,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, domain.ErrCarNotFound)
			},
		},
		{
			name:     "repo returns generic error",
			id:       "some-id",
			deleteFn: func(_ context.Context, _ string) error { return errRepo },
			wantErr:  true,
			checkErr: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, errRepo)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCarRepo{deleteFn: tc.deleteFn}
			svc := service.NewCarService(repo, silentLogger())

			err := svc.Delete(ctx, tc.id)
			if tc.wantErr {
				require.Error(t, err)
				if tc.checkErr != nil {
					tc.checkErr(t, err)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

// TestCarService_Delete_cancelledContext verifies that a cancelled context
// is detected before ID validation or repository work begins.
func TestCarService_Delete_cancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc := service.NewCarService(&mockCarRepo{}, silentLogger())
	err := svc.Delete(ctx, "some-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}
