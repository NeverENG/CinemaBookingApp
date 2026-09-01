package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type fakeMovieRepo struct {
	movies map[int64]*domain.Movie
}

func (f *fakeMovieRepo) Create(ctx context.Context, movie *domain.Movie) error {
	if f.movies == nil {
		f.movies = make(map[int64]*domain.Movie)
	}
	movie.ID = int64(len(f.movies) + 1)
	f.movies[movie.ID] = movie
	return nil
}

func (f *fakeMovieRepo) GetByID(ctx context.Context, id int64) (*domain.Movie, error) {
	if m, ok := f.movies[id]; ok {
		return m, nil
	}
	return nil, domain.ErrMovieNotFound
}

func (f *fakeMovieRepo) Update(ctx context.Context, movie *domain.Movie) error {
	if _, ok := f.movies[movie.ID]; !ok {
		return domain.ErrMovieNotFound
	}
	f.movies[movie.ID] = movie
	return nil
}

func (f *fakeMovieRepo) List(ctx context.Context) ([]domain.Movie, error) {
	out := make([]domain.Movie, 0, len(f.movies))
	for _, m := range f.movies {
		out = append(out, *m)
	}
	return out, nil
}

func (f *fakeMovieRepo) SetStatus(ctx context.Context, id int64, status domain.MovieStatus) error {
	if m, ok := f.movies[id]; ok {
		m.Status = status
		return nil
	}
	return domain.ErrMovieNotFound
}

type fakeHallRepo struct {
	halls map[int64]*domain.Hall
}

func (f *fakeHallRepo) Create(ctx context.Context, hall *domain.Hall) error {
	if f.halls == nil {
		f.halls = make(map[int64]*domain.Hall)
	}
	hall.ID = int64(len(f.halls) + 1)
	f.halls[hall.ID] = hall
	return nil
}

func (f *fakeHallRepo) GetByID(ctx context.Context, id int64) (*domain.Hall, error) {
	if h, ok := f.halls[id]; ok {
		return h, nil
	}
	return nil, domain.ErrHallNotFound
}

func (f *fakeHallRepo) Update(ctx context.Context, hall *domain.Hall) error {
	if _, ok := f.halls[hall.ID]; !ok {
		return domain.ErrHallNotFound
	}
	f.halls[hall.ID] = hall
	return nil
}

func (f *fakeHallRepo) ListByCinema(ctx context.Context, cinemaID int64) ([]domain.Hall, error) {
	out := make([]domain.Hall, 0)
	for _, h := range f.halls {
		if h.CinemaID == cinemaID {
			out = append(out, *h)
		}
	}
	return out, nil
}

type fakeOperationLogRepo struct {
	logs []*domain.OperationLog
}

func (f *fakeOperationLogRepo) Create(ctx context.Context, l *domain.OperationLog) error {
	f.logs = append(f.logs, l)
	return nil
}

type fakeRefundRepo struct {
	refunds map[string]*domain.Refund
}

func (f *fakeRefundRepo) Create(ctx context.Context, refund *domain.Refund) error {
	if f.refunds == nil {
		f.refunds = make(map[string]*domain.Refund)
	}
	f.refunds[refund.RefundNo] = refund
	return nil
}

func (f *fakeRefundRepo) GetByRefundNo(ctx context.Context, refundNo string) (*domain.Refund, error) {
	if r, ok := f.refunds[refundNo]; ok {
		return r, nil
	}
	return nil, domain.ErrRefundNotFound
}

func (f *fakeRefundRepo) GetByOrderNo(ctx context.Context, orderNo string) (*domain.Refund, error) {
	for _, r := range f.refunds {
		if r.OrderNo == orderNo {
			return r, nil
		}
	}
	return nil, domain.ErrRefundNotFound
}

func (f *fakeRefundRepo) MarkSuccess(ctx context.Context, refundNo string) error {
	if r, ok := f.refunds[refundNo]; ok {
		r.Status = domain.RefundSuccess
	}
	return nil
}

var superAdminScope = domain.AdminScope{AdminID: 1, Role: domain.RoleSuperAdmin}

func TestAdminMovieCreate(t *testing.T) {
	movies := &fakeMovieRepo{}
	logs := &fakeOperationLogRepo{}
	svc := NewAdminMovieSvc(movies, logs)

	movie, err := svc.Create(context.Background(), superAdminScope, MovieInput{
		Title:           "沙丘3",
		DurationMinutes: 160,
		Rating:          8.8,
	})
	if err != nil {
		t.Fatalf("create movie: %v", err)
	}
	if movie.Status != domain.MovieOnSale || len(movies.movies) != 1 {
		t.Fatal("unexpected movie state")
	}
	if len(logs.logs) != 1 || logs.logs[0].Action != "CREATE_MOVIE" {
		t.Fatal("expected audit log")
	}
}

func TestAdminMovieCreateInvalid(t *testing.T) {
	movies := &fakeMovieRepo{}
	logs := &fakeOperationLogRepo{}
	svc := NewAdminMovieSvc(movies, logs)

	_, err := svc.Create(context.Background(), superAdminScope, MovieInput{})
	if !errors.Is(err, domain.ErrMovieInvalid) {
		t.Fatalf("expected ErrMovieInvalid, got %v", err)
	}
	if len(logs.logs) != 0 {
		t.Fatal("invalid create should not log")
	}
}

func TestAdminHallCreateSyncsSeats(t *testing.T) {
	halls := &fakeHallRepo{}
	seats := &fakeSeatRepo{}
	logs := &fakeOperationLogRepo{}
	svc := NewAdminHallSvc(halls, seats, logs)

	hall, err := svc.Create(context.Background(), superAdminScope, HallInput{
		CinemaID:   10,
		Name:       "1号厅",
		SeatLayout: `{"rows":2,"cols":3,"disabled":["A1"],"seat_types":{"A2":"VIP"}}`,
	})
	if err != nil {
		t.Fatalf("create hall: %v", err)
	}
	if hall.ID == 0 {
		t.Fatal("expected hall id")
	}
	if len(seats.synced) != 1 || len(seats.synced[0]) != 6 {
		t.Fatalf("expected 6 synced seats, got %d", len(seats.synced[0]))
	}
	byNo := make(map[string]domain.Seat, 6)
	for _, s := range seats.synced[0] {
		byNo[s.SeatNo] = s
	}
	if byNo["A1"].Status != domain.SeatDisabled {
		t.Fatal("A1 should be disabled")
	}
	if byNo["A2"].Type != "VIP" {
		t.Fatal("A2 should be VIP")
	}
	if len(logs.logs) != 1 {
		t.Fatal("expected audit log")
	}
}

func TestParseSeatLayoutInvalid(t *testing.T) {
	if _, err := parseSeatLayout(`{"rows":0,"cols":3}`); !errors.Is(err, domain.ErrSeatLayoutInvalid) {
		t.Fatalf("expected ErrSeatLayoutInvalid, got %v", err)
	}
}

func TestBuildSeatsSupportsMoreThan26Rows(t *testing.T) {
	seats := buildSeats(&seatLayout{Rows: 28, Cols: 1})
	if len(seats) != 28 || seats[25].SeatNo != "Z1" || seats[26].SeatNo != "AA1" || seats[27].SeatNo != "AB1" {
		t.Fatalf("unexpected row labels: %+v", seats[24:])
	}
}

func TestAdminSessionCreate(t *testing.T) {
	now := time.Now()
	sessions := &fakeSessionRepo{}
	movies := &fakeMovieRepo{movies: map[int64]*domain.Movie{
		1: {ID: 1, Title: "沙丘3", DurationMinutes: 90, Status: domain.MovieOnSale},
	}}
	halls := &fakeHallRepo{halls: map[int64]*domain.Hall{
		10: {ID: 10, CinemaID: 100, Name: "1号厅"},
	}}
	svc := NewAdminSessionSvc(sessions, movies, halls, &fakeSeatLockRepo{}, &fakeOrderRepo{}, &fakeCouponRepo{}, &fakeRefundRepo{}, &fakePaymentRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{}, &fakeOperationLogRepo{})

	session, err := svc.Create(context.Background(), superAdminScope, SessionInput{
		CinemaID:       100,
		HallID:         10,
		MovieID:        1,
		StartTime:      now.Add(2 * time.Hour),
		EndTime:        now.Add(3 * time.Hour),
		BasePriceCents: 5000,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if session.Status != domain.SessionOpen || session.ID == 0 {
		t.Fatal("unexpected session state")
	}
}

func TestAdminSessionPriceRules(t *testing.T) {
	now := time.Now()
	sessions := &fakeSessionRepo{}
	movies := &fakeMovieRepo{movies: map[int64]*domain.Movie{
		1: {ID: 1, Title: "沙丘3", DurationMinutes: 90, Status: domain.MovieOnSale},
	}}
	halls := &fakeHallRepo{halls: map[int64]*domain.Hall{
		10: {ID: 10, CinemaID: 100, Name: "IMAX厅"},
	}}
	svc := NewAdminSessionSvc(sessions, movies, halls, &fakeSeatLockRepo{}, &fakeOrderRepo{}, &fakeCouponRepo{}, &fakeRefundRepo{}, &fakePaymentRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{}, &fakeOperationLogRepo{})

	session, err := svc.Create(context.Background(), superAdminScope, SessionInput{
		CinemaID:       100,
		HallID:         10,
		MovieID:        1,
		StartTime:      now.Add(2 * time.Hour),
		EndTime:        now.Add(3 * time.Hour),
		BasePriceCents: 5000,
		PriceRulesJSON: `{"VIP":8000}`,
	})
	if err != nil {
		t.Fatalf("create session with price rules: %v", err)
	}
	if session.PriceRulesJSON != `{"VIP":8000}` {
		t.Fatalf("expected normalized VIP rules, got %s", session.PriceRulesJSON)
	}

	if err := svc.UpdatePrice(context.Background(), superAdminScope, session.ID, 6000, `{"VIP":9000}`); err != nil {
		t.Fatalf("update session price rules: %v", err)
	}
	updated := sessions.sessions[session.ID]
	if updated.BasePriceCents != 6000 || updated.PriceRulesJSON != `{"VIP":9000}` {
		t.Fatalf("unexpected updated prices: %+v", updated)
	}
}

func TestAdminSessionRejectsInvalidPriceRules(t *testing.T) {
	now := time.Now()
	sessions := &fakeSessionRepo{}
	movies := &fakeMovieRepo{movies: map[int64]*domain.Movie{1: {ID: 1, Title: "沙丘3", DurationMinutes: 90, Status: domain.MovieOnSale}}}
	halls := &fakeHallRepo{halls: map[int64]*domain.Hall{10: {ID: 10, CinemaID: 100, Name: "1号厅"}}}
	svc := NewAdminSessionSvc(sessions, movies, halls, &fakeSeatLockRepo{}, &fakeOrderRepo{}, &fakeCouponRepo{}, &fakeRefundRepo{}, &fakePaymentRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{}, &fakeOperationLogRepo{})

	_, err := svc.Create(context.Background(), superAdminScope, SessionInput{
		CinemaID:       100,
		HallID:         10,
		MovieID:        1,
		StartTime:      now.Add(2 * time.Hour),
		EndTime:        now.Add(3 * time.Hour),
		BasePriceCents: 5000,
		PriceRulesJSON: `{"VIP":0}`,
	})
	if !errors.Is(err, domain.ErrSessionInvalid) {
		t.Fatalf("expected invalid price rules, got %v", err)
	}
}

func TestAdminSessionCreateOverlap(t *testing.T) {
	now := time.Now()
	sessions := &fakeSessionRepo{sessions: map[int64]*domain.ShowSession{
		1: {ID: 1, HallID: 10, StartTime: now.Add(time.Hour), EndTime: now.Add(2 * time.Hour), Status: domain.SessionOpen},
	}}
	movies := &fakeMovieRepo{movies: map[int64]*domain.Movie{
		1: {ID: 1, Title: "沙丘3", DurationMinutes: 90, Status: domain.MovieOnSale},
	}}
	halls := &fakeHallRepo{halls: map[int64]*domain.Hall{
		10: {ID: 10, CinemaID: 100, Name: "1号厅"},
	}}
	svc := NewAdminSessionSvc(sessions, movies, halls, &fakeSeatLockRepo{}, &fakeOrderRepo{}, &fakeCouponRepo{}, &fakeRefundRepo{}, &fakePaymentRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{}, &fakeOperationLogRepo{})

	_, err := svc.Create(context.Background(), superAdminScope, SessionInput{
		CinemaID:       100,
		HallID:         10,
		MovieID:        1,
		StartTime:      now.Add(90 * time.Minute),
		EndTime:        now.Add(150 * time.Minute),
		BasePriceCents: 5000,
	})
	if !errors.Is(err, domain.ErrSessionTimeConflict) {
		t.Fatalf("expected ErrSessionTimeConflict, got %v", err)
	}
}

func TestAdminSessionCancel(t *testing.T) {
	now := time.Now()
	sessions := &fakeSessionRepo{sessions: map[int64]*domain.ShowSession{
		5: {ID: 5, CinemaID: 100, HallID: 10, StartTime: now.Add(2 * time.Hour), EndTime: now.Add(3 * time.Hour), Status: domain.SessionOpen},
	}}
	movies := &fakeMovieRepo{}
	halls := &fakeHallRepo{}
	locks := &fakeSeatLockRepo{}
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{
		"O1": {OrderNo: "O1", SessionID: 5, Status: domain.OrderPendingPayment},
		"O2": {OrderNo: "O2", SessionID: 5, Status: domain.OrderPaid, PaidCents: 5000, Version: 1},
	}}
	coupons := &fakeCouponRepo{}
	refunds := &fakeRefundRepo{}
	payments := &fakePaymentRepo{txns: map[string]*domain.PaymentTransaction{
		"T2": {TransactionNo: "T2", OrderNo: "O2", Status: domain.PaymentSuccess, Version: 1},
	}}
	points := &fakePointsRepo{}
	box := &fakeBoxOfficeRepo{}
	logs := &fakeOperationLogRepo{}
	svc := NewAdminSessionSvc(sessions, movies, halls, locks, orders, coupons, refunds, payments, points, box, logs)

	if err := svc.Cancel(context.Background(), superAdminScope, 5); err != nil {
		t.Fatalf("cancel session: %v", err)
	}
	if sessions.sessions[5].Status != domain.SessionCanceled {
		t.Fatal("expected session CANCELED")
	}
	if len(locks.released) != 1 || locks.released[0] != 5 {
		t.Fatal("expected locks released for session")
	}
	if orders.orders["O1"].Status != domain.OrderExpired {
		t.Fatal("expected pending order EXPIRED")
	}
	if len(coupons.unlocked) != 1 {
		t.Fatal("expected coupon unlocked")
	}
	if orders.orders["O2"].Status != domain.OrderRefunded {
		t.Fatalf("expected paid order REFUNDED, got %s", orders.orders["O2"].Status)
	}
	if payments.txns["T2"].Status != domain.PaymentRefunded {
		t.Fatal("expected payment REFUNDED")
	}
	if len(refunds.refunds) != 1 {
		t.Fatal("expected refund created for paid order")
	}
	if len(points.reclaimed) != 1 {
		t.Fatal("expected points reclaimed on refund")
	}
	if len(box.events) != 1 || box.events[0].BizType != domain.BoxOrderRefund || box.events[0].RefundDelta != 5000 {
		t.Fatalf("expected refund box event, got %+v", box.events)
	}
}

func TestAdminHallForbiddenForCinemaAdmin(t *testing.T) {
	halls := &fakeHallRepo{}
	seats := &fakeSeatRepo{}
	logs := &fakeOperationLogRepo{}
	svc := NewAdminHallSvc(halls, seats, logs)
	scope := domain.AdminScope{AdminID: 2, Role: domain.RoleCinemaAdmin, CinemaID: int64Ptr(999)}

	_, err := svc.Create(context.Background(), scope, HallInput{
		CinemaID:   10,
		Name:       "1号厅",
		SeatLayout: `{"rows":2,"cols":3}`,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
