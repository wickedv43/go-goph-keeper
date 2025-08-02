package storage

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestStorage_NewUser(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock)
		expectedError error
	}

	tests := []testCase{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE login = \$1 ORDER BY "users"\."id" LIMIT \$2`).
					WithArgs("test", sqlmock.AnyArg()).
					WillReturnError(gorm.ErrRecordNotFound)

				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "users"`).
					WithArgs("test", "hashed", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name: "login already used",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE login = \$1 ORDER BY "users"\."id" LIMIT \$2`).
					WithArgs("test", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password_hash", "created_at"}).
						AddRow(1, "test", "hashed", time.Now()))
			},
			expectedError: ErrLoginUsed,
		},
		{
			name: "db error on select",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE login = \$1 ORDER BY "users"\."id" LIMIT \$2`).
					WithArgs("test", sqlmock.AnyArg()).
					WillReturnError(errors.New("canceling query due to user request"))
			},
			expectedError: errors.New("canceling query due to user request"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			require.NoError(t, err)

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			require.NoError(t, err)

			storage := &Storage{
				db:  gdb,
				log: zap.NewNop().Sugar(),
			}

			tc.setupMock(mock)

			ctx := context.Background()
			user := &User{
				Login:        "test",
				PasswordHash: "hashed",
				CreatedAt:    time.Now(),
			}

			_, err = storage.NewUser(ctx, user)
			if tc.expectedError != nil {
				require.ErrorContains(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestStorage_User(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock)
		expectedUser  User
		expectedError error
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`).
					WithArgs(1, sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password_hash", "created_at"}).
						AddRow(1, "test", "hashed", time.Now()))
			},
			expectedUser: User{
				ID:           1,
				Login:        "test",
				PasswordHash: "hashed",
			},
			expectedError: nil,
		},
		{
			name: "not_found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`).
					WithArgs(0, sqlmock.AnyArg()).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "db_error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`).
					WithArgs(0, sqlmock.AnyArg()).
					WillReturnError(errors.New("db connection error"))
			},
			expectedError: errors.New("db connection error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			require.NoError(t, err)

			s := &Storage{
				db:  gdb,
				log: zap.NewNop().Sugar(),
			}

			tc.setupMock(mock)

			ctx := context.Background()
			id := uint64(tc.expectedUser.ID)

			user, err := s.User(ctx, id)

			if tc.expectedError != nil {
				require.ErrorContains(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedUser.Login, user.Login)
				require.Equal(t, tc.expectedUser.PasswordHash, user.PasswordHash)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestStorage_UserByLogin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		login         string
		setupMock     func(sqlmock.Sqlmock)
		expectedUser  User
		expectedError error
	}{
		{
			name:  "success",
			login: "test_user",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT * FROM "users" WHERE login = $1 ORDER BY "users"."id" LIMIT $2`).
					WithArgs("test_user", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password_hash", "created_at"}).
						AddRow(42, "test_user", "hashed", time.Now()))
			},
			expectedUser: User{
				ID:           42,
				Login:        "test_user",
				PasswordHash: "hashed",
			},
			expectedError: nil,
		},
		{
			name:  "not_found",
			login: "missing",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT * FROM "users" WHERE login = $1 ORDER BY "users"."id" LIMIT $2`).
					WithArgs("missing", sqlmock.AnyArg()).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "db_error",
			login: "broken",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT * FROM "users" WHERE login = $1 ORDER BY "users"."id" LIMIT $2`).
					WithArgs("broken", sqlmock.AnyArg()).
					WillReturnError(errors.New("db failure"))
			},
			expectedError: errors.New("db failure"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			require.NoError(t, err)

			s := &Storage{
				db:  gdb,
				log: zap.NewNop().Sugar(),
			}

			tc.setupMock(mock)

			user, err := s.UserByLogin(context.Background(), tc.login)

			if tc.expectedError != nil {
				require.ErrorContains(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedUser.Login, user.Login)
				require.Equal(t, tc.expectedUser.PasswordHash, user.PasswordHash)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
