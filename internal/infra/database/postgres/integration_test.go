package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var idSeq int64

func nextID() int64 {
	return 500000 + atomic.AddInt64(&idSeq, 1)
}

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		os.Exit(m.Run()) // 测试内通过 t.Skip 跳过
	}
	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "integration db:", err)
		os.Exit(1)
	}
	if err := ApplyAllMigrations(db, "../../../../sql/migrations"); err != nil {
		fmt.Fprintln(os.Stderr, "integration migrate:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func integrationDB(t *testing.T) *DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set")
	}
	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return NewDB(db)
}

func createTestUser(t *testing.T, db *DB) int64 {
	t.Helper()
	email := fmt.Sprintf("it%d@test.com", nextID())
	user := &domain.User{Username: email, Email: email, Nickname: "it", Status: "ACTIVE"}
	if err := NewUserRepo(db).Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user.ID
}

func portFilter(statDate time.Time) port.BoxOfficeFilter {
	return port.BoxOfficeFilter{
		StartDate: statDate.Add(-time.Hour),
		EndDate:   statDate.Add(25 * time.Hour),
	}
}

func TestIntegrationPointsIdempotent(t *testing.T) {
	db := integrationDB(t)
	userID := createTestUser(t, db)
	orderNo := fmt.Sprintf("ITORDER%d", nextID())
	points := NewPointsRepo(db)

	if err := points.GrantOnPaid(context.Background(), userID, 10000, orderNo); err != nil {
		t.Fatalf("grant: %v", err)
	}
	if err := points.GrantOnPaid(context.Background(), userID, 10000, orderNo); err != nil {
		t.Fatalf("grant duplicate: %v", err)
	}

	balance, err := points.GetBalance(context.Background(), userID)
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if balance != 100 {
		t.Fatalf("expected balance 100, got %d", balance)
	}
	ledger, err := points.GetRecentLedger(context.Background(), userID, 10)
	if err != nil {
		t.Fatalf("ledger: %v", err)
	}
	if len(ledger) != 1 {
		t.Fatalf("expected 1 ledger row, got %d", len(ledger))
	}
}

func TestIntegrationBoxOfficeIdempotent(t *testing.T) {
	db := integrationDB(t)
	box := NewBoxOfficeRepo(db)
	now := time.Now()
	event := domain.NewPaidEvent(&domain.Order{
		CinemaID:  nextID(),
		MovieID:   nextID(),
		PaidCents: 10000,
		Items:     []domain.OrderItem{{}},
	}, fmt.Sprintf("ITBOX%d", nextID()), now)

	if err := box.Record(context.Background(), event); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := box.Record(context.Background(), event); err != nil {
		t.Fatalf("record duplicate: %v", err)
	}
	s, err := box.Summary(context.Background(), portFilter(event.StatDate))
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if s.GrossCents != 10000 || s.OrderCount != 1 {
		t.Fatalf("expected gross 10000/orders 1, got %+v", s)
	}
}

func TestIntegrationCouponLockAtomic(t *testing.T) {
	db := integrationDB(t)
	userID := createTestUser(t, db)
	repo := NewUserCouponRepo(db)
	tpl := &domain.CouponTemplate{
		Name: "测试券", Type: domain.CouponTypeFixed, ValueCents: 1000,
		Status: "ACTIVE", ValidDays: 30, PerUserLimit: 1,
	}
	if err := repo.CreateTemplate(context.Background(), tpl); err != nil {
		t.Fatalf("create template: %v", err)
	}
	coupon := &domain.UserCoupon{
		CouponNo: fmt.Sprintf("ITC%d", nextID()), UserID: userID, TemplateID: tpl.ID,
		Status: domain.CouponUnused, ExpireAt: time.Now().Add(24 * time.Hour),
	}
	if err := repo.CreateInstance(context.Background(), coupon); err != nil {
		t.Fatalf("create instance: %v", err)
	}
	if err := repo.LockForOrder(context.Background(), coupon.CouponNo, "ORDER1"); err != nil {
		t.Fatalf("lock: %v", err)
	}
	if err := repo.LockForOrder(context.Background(), coupon.CouponNo, "ORDER2"); !errors.Is(err, domain.ErrCouponNotAvailable) {
		t.Fatalf("expected ErrCouponNotAvailable on second lock, got %v", err)
	}
}

func TestIntegrationSeatLockUnique(t *testing.T) {
	db := integrationDB(t)
	gormDB := db.db(context.Background())
	cinemaID, hallID, movieID, sessionID, seatID := nextID(), nextID(), nextID(), nextID(), nextID()
	userID := createTestUser(t, db)

	inserts := []string{
		fmt.Sprintf(`INSERT INTO cinemas (id,name,city,address,status) VALUES (%d,'IT影院','上海','x','ACTIVE') ON CONFLICT (id) DO NOTHING`, cinemaID),
		fmt.Sprintf(`INSERT INTO movies (id,title,cover_url,duration_minutes,status) VALUES (%d,'IT电影','http://x',120,'ON_SALE') ON CONFLICT (id) DO NOTHING`, movieID),
		fmt.Sprintf(`INSERT INTO halls (id,cinema_id,name,seat_layout_json,status) VALUES (%d,%d,'IT厅','{}','ACTIVE') ON CONFLICT (id) DO NOTHING`, hallID, cinemaID),
		fmt.Sprintf(`INSERT INTO seats (id,hall_id,row_no,col_no,seat_no,type,status) VALUES (%d,%d,1,1,'Z1','STANDARD','ENABLED') ON CONFLICT (id) DO NOTHING`, seatID, hallID),
		fmt.Sprintf(`INSERT INTO show_sessions (id,cinema_id,hall_id,movie_id,start_time,end_time,base_price_cents,status) VALUES (%d,%d,%d,%d, now()+interval '1 day', now()+interval '1 day'+interval '2 hours', 5000,'OPEN') ON CONFLICT (id) DO NOTHING`, sessionID, cinemaID, hallID, movieID),
	}
	for _, sql := range inserts {
		if err := gormDB.Exec(sql).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	locks := NewSeatLockRepo(db)
	lock := domain.SeatLock{
		SessionID: sessionID, SeatID: seatID, UserID: userID, OrderNo: "ORD1",
		LockToken: "tok1", Status: domain.SeatLockLocked, ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := locks.CreateLocks(context.Background(), []domain.SeatLock{lock}); err != nil {
		t.Fatalf("create lock: %v", err)
	}
	lock.OrderNo = "ORD2"
	if err := locks.CreateLocks(context.Background(), []domain.SeatLock{lock}); !errors.Is(err, domain.ErrSeatLockConflict) {
		t.Fatalf("expected ErrSeatLockConflict, got %v", err)
	}
}
