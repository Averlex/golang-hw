package sql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql/mocks" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/types"     //nolint:depguard,nolintlint
	"github.com/jmoiron/sqlx"                                                           //nolint:depguard,nolintlint
	"github.com/stretchr/testify/mock"                                                  //nolint:depguard,nolintlint
	"github.com/stretchr/testify/suite"                                                 //nolint:depguard,nolintlint
)

type StorageSuite struct {
	suite.Suite
	dbMock        *mocks.DB
	txMock        *mocks.Tx
	storage       *Storage
	defaultUserID string
	shortTimeout  time.Duration
}

func (s *StorageSuite) SetupSuite() {
	s.dbMock = &mocks.DB{}
	s.txMock = &mocks.Tx{}
	s.storage = &Storage{
		driver:  "postgres",
		dsn:     "mock_dsn",
		timeout: 1 * time.Second,
		db:      s.dbMock,
	}
	s.defaultUserID = "user1"
	s.shortTimeout = 1 * time.Nanosecond
}

func (s *StorageSuite) SetupTest() {
	s.dbMock.ExpectedCalls = nil
	s.txMock.ExpectedCalls = nil
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}

func (s *StorageSuite) TestNewStorage() {
	testCases := []struct {
		name    string
		driver  string
		host    string
		port    string
		user    string
		pass    string
		dbname  string
		wantErr error
	}{
		{
			name:    "valid postgres driver",
			driver:  "postgres",
			host:    "localhost",
			port:    "5432",
			user:    "user",
			pass:    "pass",
			dbname:  "calendar",
			wantErr: nil,
		},
		{
			name:    "valid postgresql driver",
			driver:  "postgresql",
			host:    "localhost",
			port:    "5432",
			user:    "user",
			pass:    "pass",
			dbname:  "calendar",
			wantErr: nil,
		},
		{
			name:    "invalid driver",
			driver:  "mysql",
			host:    "localhost",
			port:    "3306",
			user:    "user",
			pass:    "pass",
			dbname:  "calendar",
			wantErr: types.ErrUnsupportedDriver,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			storage, err := NewStorage(5*time.Second, tC.driver, tC.host, tC.port, tC.user, tC.pass, tC.dbname)
			if tC.wantErr != nil {
				s.Require().ErrorIs(err, tC.wantErr)
				s.Require().Nil(storage)
				return
			}
			s.Require().NoError(err)
			s.Require().NotNil(storage)
			s.Require().NotEmpty(storage.dsn)
		})
	}
}

func (s *StorageSuite) TestConnect() {
	expErr := types.ErrTimeoutExceeded
	testCases := []struct {
		name      string
		mockSetup func()
		wantErr   bool
	}{
		{
			name: "successful connect",
			mockSetup: func() {
				s.dbMock.On("ConnectContext", mock.Anything,
					"postgres", "mock_dsn").Return(&sqlx.DB{}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "connection error",
			mockSetup: func() {
				s.dbMock.On("ConnectContext", mock.Anything, "postgres", "mock_dsn").Return((*sqlx.DB)(nil),
					errors.New("connect failed")).Once()
			},
			wantErr: true,
		},
		{
			name: "timeout exceeded",
			mockSetup: func() {
				s.dbMock.On("ConnectContext", mock.Anything, "postgres", "mock_dsn").Return((*sqlx.DB)(nil),
					context.DeadlineExceeded).Once()
			},
			wantErr: true,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.mockSetup()
			ctx, cancel := context.WithTimeout(context.Background(), s.shortTimeout)
			defer cancel()
			err := s.storage.Connect(ctx)
			if tC.wantErr {
				s.Require().Error(err)
				if tC.name == "timeout exceeded" {
					s.Require().ErrorIs(err, expErr)
				}
				return
			}
			s.Require().NoError(err)
			s.dbMock.AssertExpectations(s.T())
		})
	}
}

func (s *StorageSuite) TestClose() {
	testCases := []struct {
		name      string
		dbSetup   func()
		mockSetup func()
	}{
		{
			name:    "successful close",
			dbSetup: func() {}, // Already initialized in SetupSuite.
			mockSetup: func() {
				s.dbMock.On("Close").Return(nil).Once()
			},
		},
		{
			name:    "nil db",
			dbSetup: func() { s.storage.db = &SQLXWrapper{} }, // Underlying DB is nil.
			mockSetup: func() {
				s.dbMock.On("Close").Return(nil).Maybe()
			},
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.dbSetup()
			tC.mockSetup()
			s.storage.Close(context.Background())
			s.dbMock.AssertExpectations(s.T())
		})
	}
}
