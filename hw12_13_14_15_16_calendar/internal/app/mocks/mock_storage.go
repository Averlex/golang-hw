// Code generated by mockery v2.53.4. DO NOT EDIT.

package mocks

import (
	context "context"
	time "time"

	mock "github.com/stretchr/testify/mock"

	types "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types"

	uuid "github.com/google/uuid"
)

// Storage is an autogenerated mock type for the Storage type
type Storage struct {
	mock.Mock
}

type Storage_Expecter struct {
	mock *mock.Mock
}

func (_m *Storage) EXPECT() *Storage_Expecter {
	return &Storage_Expecter{mock: &_m.Mock}
}

// Connect provides a mock function with given fields: ctx
func (_m *Storage) Connect(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Connect")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_Connect_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Connect'
type Storage_Connect_Call struct {
	*mock.Call
}

// Connect is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Storage_Expecter) Connect(ctx interface{}) *Storage_Connect_Call {
	return &Storage_Connect_Call{Call: _e.mock.On("Connect", ctx)}
}

func (_c *Storage_Connect_Call) Run(run func(ctx context.Context)) *Storage_Connect_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Storage_Connect_Call) Return(_a0 error) *Storage_Connect_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_Connect_Call) RunAndReturn(run func(context.Context) error) *Storage_Connect_Call {
	_c.Call.Return(run)
	return _c
}

// CreateEvent provides a mock function with given fields: ctx, event
func (_m *Storage) CreateEvent(ctx context.Context, event *types.Event) (*types.Event, error) {
	ret := _m.Called(ctx, event)

	if len(ret) == 0 {
		panic("no return value specified for CreateEvent")
	}

	var r0 *types.Event
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.Event) (*types.Event, error)); ok {
		return rf(ctx, event)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *types.Event) *types.Event); ok {
		r0 = rf(ctx, event)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Event)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *types.Event) error); ok {
		r1 = rf(ctx, event)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_CreateEvent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateEvent'
type Storage_CreateEvent_Call struct {
	*mock.Call
}

// CreateEvent is a helper method to define mock.On call
//   - ctx context.Context
//   - event *types.Event
func (_e *Storage_Expecter) CreateEvent(ctx interface{}, event interface{}) *Storage_CreateEvent_Call {
	return &Storage_CreateEvent_Call{Call: _e.mock.On("CreateEvent", ctx, event)}
}

func (_c *Storage_CreateEvent_Call) Run(run func(ctx context.Context, event *types.Event)) *Storage_CreateEvent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*types.Event))
	})
	return _c
}

func (_c *Storage_CreateEvent_Call) Return(_a0 *types.Event, _a1 error) *Storage_CreateEvent_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_CreateEvent_Call) RunAndReturn(run func(context.Context, *types.Event) (*types.Event, error)) *Storage_CreateEvent_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteEvent provides a mock function with given fields: ctx, id
func (_m *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for DeleteEvent")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_DeleteEvent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteEvent'
type Storage_DeleteEvent_Call struct {
	*mock.Call
}

// DeleteEvent is a helper method to define mock.On call
//   - ctx context.Context
//   - id uuid.UUID
func (_e *Storage_Expecter) DeleteEvent(ctx interface{}, id interface{}) *Storage_DeleteEvent_Call {
	return &Storage_DeleteEvent_Call{Call: _e.mock.On("DeleteEvent", ctx, id)}
}

func (_c *Storage_DeleteEvent_Call) Run(run func(ctx context.Context, id uuid.UUID)) *Storage_DeleteEvent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uuid.UUID))
	})
	return _c
}

func (_c *Storage_DeleteEvent_Call) Return(_a0 error) *Storage_DeleteEvent_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_DeleteEvent_Call) RunAndReturn(run func(context.Context, uuid.UUID) error) *Storage_DeleteEvent_Call {
	_c.Call.Return(run)
	return _c
}

// GetAllUserEvents provides a mock function with given fields: ctx, userID
func (_m *Storage) GetAllUserEvents(ctx context.Context, userID string) ([]*types.Event, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetAllUserEvents")
	}

	var r0 []*types.Event
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*types.Event, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*types.Event); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*types.Event)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_GetAllUserEvents_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAllUserEvents'
type Storage_GetAllUserEvents_Call struct {
	*mock.Call
}

// GetAllUserEvents is a helper method to define mock.On call
//   - ctx context.Context
//   - userID string
func (_e *Storage_Expecter) GetAllUserEvents(ctx interface{}, userID interface{}) *Storage_GetAllUserEvents_Call {
	return &Storage_GetAllUserEvents_Call{Call: _e.mock.On("GetAllUserEvents", ctx, userID)}
}

func (_c *Storage_GetAllUserEvents_Call) Run(run func(ctx context.Context, userID string)) *Storage_GetAllUserEvents_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Storage_GetAllUserEvents_Call) Return(_a0 []*types.Event, _a1 error) *Storage_GetAllUserEvents_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_GetAllUserEvents_Call) RunAndReturn(run func(context.Context, string) ([]*types.Event, error)) *Storage_GetAllUserEvents_Call {
	_c.Call.Return(run)
	return _c
}

// GetEvent provides a mock function with given fields: ctx, id
func (_m *Storage) GetEvent(ctx context.Context, id uuid.UUID) (*types.Event, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetEvent")
	}

	var r0 *types.Event
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*types.Event, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *types.Event); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Event)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_GetEvent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEvent'
type Storage_GetEvent_Call struct {
	*mock.Call
}

// GetEvent is a helper method to define mock.On call
//   - ctx context.Context
//   - id uuid.UUID
func (_e *Storage_Expecter) GetEvent(ctx interface{}, id interface{}) *Storage_GetEvent_Call {
	return &Storage_GetEvent_Call{Call: _e.mock.On("GetEvent", ctx, id)}
}

func (_c *Storage_GetEvent_Call) Run(run func(ctx context.Context, id uuid.UUID)) *Storage_GetEvent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uuid.UUID))
	})
	return _c
}

func (_c *Storage_GetEvent_Call) Return(_a0 *types.Event, _a1 error) *Storage_GetEvent_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_GetEvent_Call) RunAndReturn(run func(context.Context, uuid.UUID) (*types.Event, error)) *Storage_GetEvent_Call {
	_c.Call.Return(run)
	return _c
}

// GetEventsForDay provides a mock function with given fields: ctx, date, userID
func (_m *Storage) GetEventsForDay(ctx context.Context, date time.Time, userID *string) ([]*types.Event, error) {
	ret := _m.Called(ctx, date, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetEventsForDay")
	}

	var r0 []*types.Event
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Time, *string) ([]*types.Event, error)); ok {
		return rf(ctx, date, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, time.Time, *string) []*types.Event); ok {
		r0 = rf(ctx, date, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*types.Event)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, time.Time, *string) error); ok {
		r1 = rf(ctx, date, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_GetEventsForDay_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEventsForDay'
type Storage_GetEventsForDay_Call struct {
	*mock.Call
}

// GetEventsForDay is a helper method to define mock.On call
//   - ctx context.Context
//   - date time.Time
//   - userID *string
func (_e *Storage_Expecter) GetEventsForDay(ctx interface{}, date interface{}, userID interface{}) *Storage_GetEventsForDay_Call {
	return &Storage_GetEventsForDay_Call{Call: _e.mock.On("GetEventsForDay", ctx, date, userID)}
}

func (_c *Storage_GetEventsForDay_Call) Run(run func(ctx context.Context, date time.Time, userID *string)) *Storage_GetEventsForDay_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(time.Time), args[2].(*string))
	})
	return _c
}

func (_c *Storage_GetEventsForDay_Call) Return(_a0 []*types.Event, _a1 error) *Storage_GetEventsForDay_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_GetEventsForDay_Call) RunAndReturn(run func(context.Context, time.Time, *string) ([]*types.Event, error)) *Storage_GetEventsForDay_Call {
	_c.Call.Return(run)
	return _c
}

// GetEventsForMonth provides a mock function with given fields: ctx, date, userID
func (_m *Storage) GetEventsForMonth(ctx context.Context, date time.Time, userID *string) ([]*types.Event, error) {
	ret := _m.Called(ctx, date, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetEventsForMonth")
	}

	var r0 []*types.Event
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Time, *string) ([]*types.Event, error)); ok {
		return rf(ctx, date, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, time.Time, *string) []*types.Event); ok {
		r0 = rf(ctx, date, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*types.Event)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, time.Time, *string) error); ok {
		r1 = rf(ctx, date, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_GetEventsForMonth_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEventsForMonth'
type Storage_GetEventsForMonth_Call struct {
	*mock.Call
}

// GetEventsForMonth is a helper method to define mock.On call
//   - ctx context.Context
//   - date time.Time
//   - userID *string
func (_e *Storage_Expecter) GetEventsForMonth(ctx interface{}, date interface{}, userID interface{}) *Storage_GetEventsForMonth_Call {
	return &Storage_GetEventsForMonth_Call{Call: _e.mock.On("GetEventsForMonth", ctx, date, userID)}
}

func (_c *Storage_GetEventsForMonth_Call) Run(run func(ctx context.Context, date time.Time, userID *string)) *Storage_GetEventsForMonth_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(time.Time), args[2].(*string))
	})
	return _c
}

func (_c *Storage_GetEventsForMonth_Call) Return(_a0 []*types.Event, _a1 error) *Storage_GetEventsForMonth_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_GetEventsForMonth_Call) RunAndReturn(run func(context.Context, time.Time, *string) ([]*types.Event, error)) *Storage_GetEventsForMonth_Call {
	_c.Call.Return(run)
	return _c
}

// GetEventsForPeriod provides a mock function with given fields: ctx, dateStart, dateEnd, userID
func (_m *Storage) GetEventsForPeriod(ctx context.Context, dateStart time.Time, dateEnd time.Time, userID *string) ([]*types.Event, error) {
	ret := _m.Called(ctx, dateStart, dateEnd, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetEventsForPeriod")
	}

	var r0 []*types.Event
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Time, time.Time, *string) ([]*types.Event, error)); ok {
		return rf(ctx, dateStart, dateEnd, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, time.Time, time.Time, *string) []*types.Event); ok {
		r0 = rf(ctx, dateStart, dateEnd, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*types.Event)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, time.Time, time.Time, *string) error); ok {
		r1 = rf(ctx, dateStart, dateEnd, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_GetEventsForPeriod_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEventsForPeriod'
type Storage_GetEventsForPeriod_Call struct {
	*mock.Call
}

// GetEventsForPeriod is a helper method to define mock.On call
//   - ctx context.Context
//   - dateStart time.Time
//   - dateEnd time.Time
//   - userID *string
func (_e *Storage_Expecter) GetEventsForPeriod(ctx interface{}, dateStart interface{}, dateEnd interface{}, userID interface{}) *Storage_GetEventsForPeriod_Call {
	return &Storage_GetEventsForPeriod_Call{Call: _e.mock.On("GetEventsForPeriod", ctx, dateStart, dateEnd, userID)}
}

func (_c *Storage_GetEventsForPeriod_Call) Run(run func(ctx context.Context, dateStart time.Time, dateEnd time.Time, userID *string)) *Storage_GetEventsForPeriod_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(time.Time), args[2].(time.Time), args[3].(*string))
	})
	return _c
}

func (_c *Storage_GetEventsForPeriod_Call) Return(_a0 []*types.Event, _a1 error) *Storage_GetEventsForPeriod_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_GetEventsForPeriod_Call) RunAndReturn(run func(context.Context, time.Time, time.Time, *string) ([]*types.Event, error)) *Storage_GetEventsForPeriod_Call {
	_c.Call.Return(run)
	return _c
}

// GetEventsForWeek provides a mock function with given fields: ctx, date, userID
func (_m *Storage) GetEventsForWeek(ctx context.Context, date time.Time, userID *string) ([]*types.Event, error) {
	ret := _m.Called(ctx, date, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetEventsForWeek")
	}

	var r0 []*types.Event
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Time, *string) ([]*types.Event, error)); ok {
		return rf(ctx, date, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, time.Time, *string) []*types.Event); ok {
		r0 = rf(ctx, date, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*types.Event)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, time.Time, *string) error); ok {
		r1 = rf(ctx, date, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_GetEventsForWeek_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEventsForWeek'
type Storage_GetEventsForWeek_Call struct {
	*mock.Call
}

// GetEventsForWeek is a helper method to define mock.On call
//   - ctx context.Context
//   - date time.Time
//   - userID *string
func (_e *Storage_Expecter) GetEventsForWeek(ctx interface{}, date interface{}, userID interface{}) *Storage_GetEventsForWeek_Call {
	return &Storage_GetEventsForWeek_Call{Call: _e.mock.On("GetEventsForWeek", ctx, date, userID)}
}

func (_c *Storage_GetEventsForWeek_Call) Run(run func(ctx context.Context, date time.Time, userID *string)) *Storage_GetEventsForWeek_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(time.Time), args[2].(*string))
	})
	return _c
}

func (_c *Storage_GetEventsForWeek_Call) Return(_a0 []*types.Event, _a1 error) *Storage_GetEventsForWeek_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_GetEventsForWeek_Call) RunAndReturn(run func(context.Context, time.Time, *string) ([]*types.Event, error)) *Storage_GetEventsForWeek_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateEvent provides a mock function with given fields: ctx, id, data
func (_m *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, data *types.EventData) (*types.Event, error) {
	ret := _m.Called(ctx, id, data)

	if len(ret) == 0 {
		panic("no return value specified for UpdateEvent")
	}

	var r0 *types.Event
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *types.EventData) (*types.Event, error)); ok {
		return rf(ctx, id, data)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *types.EventData) *types.Event); ok {
		r0 = rf(ctx, id, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Event)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, *types.EventData) error); ok {
		r1 = rf(ctx, id, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_UpdateEvent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateEvent'
type Storage_UpdateEvent_Call struct {
	*mock.Call
}

// UpdateEvent is a helper method to define mock.On call
//   - ctx context.Context
//   - id uuid.UUID
//   - data *types.EventData
func (_e *Storage_Expecter) UpdateEvent(ctx interface{}, id interface{}, data interface{}) *Storage_UpdateEvent_Call {
	return &Storage_UpdateEvent_Call{Call: _e.mock.On("UpdateEvent", ctx, id, data)}
}

func (_c *Storage_UpdateEvent_Call) Run(run func(ctx context.Context, id uuid.UUID, data *types.EventData)) *Storage_UpdateEvent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uuid.UUID), args[2].(*types.EventData))
	})
	return _c
}

func (_c *Storage_UpdateEvent_Call) Return(_a0 *types.Event, _a1 error) *Storage_UpdateEvent_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_UpdateEvent_Call) RunAndReturn(run func(context.Context, uuid.UUID, *types.EventData) (*types.Event, error)) *Storage_UpdateEvent_Call {
	_c.Call.Return(run)
	return _c
}

// NewStorage creates a new instance of Storage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *Storage {
	mock := &Storage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
