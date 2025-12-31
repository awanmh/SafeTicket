package repository

import (
	"context"
	"fmt"
	"safe-ticket/internal/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository interface {
	BookTicketUnsafe(ctx context.Context, eventID int, userID string) error
	BookTicketSafe(ctx context.Context, eventID int, userID string) error
	GetEvent(ctx context.Context, eventID int) (*domain.Event, error)
}

type PostgresEventRepository struct {
	DB *pgxpool.Pool
}

func NewPostgresEventRepository(db *pgxpool.Pool) EventRepository {
	return &PostgresEventRepository{DB: db}
}

// BookTicketUnsafe suffers from race conditions (check-then-act)
func (r *PostgresEventRepository) BookTicketUnsafe(ctx context.Context, eventID int, userID string) error {
	// 1. Cek Stok (Check)
	var totalTickets int
	err := r.DB.QueryRow(ctx, "SELECT total_tickets FROM events WHERE id = $1", eventID).Scan(&totalTickets)
	if err != nil {
		return err
	}

	if totalTickets <= 0 {
		return fmt.Errorf("sold out")
	}

	// --- AREA BERBAHAYA (CRITICAL SECTION) ---
	// Kita "pura-pura" server sedang sibuk mikir atau loading network selama 300ms.
	// Selama 300ms ini, database belum di-update.
	// Ribuan user lain akan lolos pengecekan di langkah 1 karena stok di mata mereka masih ada!
	time.Sleep(300 * time.Millisecond) 
	// -----------------------------------------

	// 2. Kurangi Stok (Act)
	_, err = r.DB.Exec(ctx, "UPDATE events SET total_tickets = total_tickets - 1 WHERE id = $1", eventID)
	if err != nil {
		return err
	}

	// 3. Catat Booking
	_, err = r.DB.Exec(ctx, "INSERT INTO bookings (event_id, user_id) VALUES ($1, $2)", eventID, userID)
	return err
}

// BookTicketSafe uses explicit locking
func (r *PostgresEventRepository) BookTicketSafe(ctx context.Context, eventID int, userID string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Lock the row for update
	var totalTickets int
	err = tx.QueryRow(ctx, "SELECT total_tickets FROM events WHERE id = $1 FOR UPDATE", eventID).Scan(&totalTickets)
	if err != nil {
		return err
	}

	if totalTickets <= 0 {
		return fmt.Errorf("sold out")
	}

	// 2. Decrement stock
	_, err = tx.Exec(ctx, "UPDATE events SET total_tickets = total_tickets - 1 WHERE id = $1", eventID)
	if err != nil {
		return err
	}

	// 3. Record booking
	_, err = tx.Exec(ctx, "INSERT INTO bookings (event_id, user_id) VALUES ($1, $2)", eventID, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresEventRepository) GetEvent(ctx context.Context, eventID int) (*domain.Event, error) {
	var event domain.Event
	err := r.DB.QueryRow(ctx, "SELECT id, name, total_tickets FROM events WHERE id = $1", eventID).Scan(&event.ID, &event.Name, &event.TotalTickets)
	if err != nil {
		return nil, err
	}
	return &event, nil
}
