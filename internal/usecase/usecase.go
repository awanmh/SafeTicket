package usecase

import (
	"context"
	"safe-ticket/internal/domain"
	"safe-ticket/internal/repository"
)

type EventUsecase struct {
	Repo repository.EventRepository
}

func NewEventUsecase(repo repository.EventRepository) *EventUsecase {
	return &EventUsecase{Repo: repo}
}

func (u *EventUsecase) BookTicket(ctx context.Context, req domain.BookingRequest, safe bool) error {
	if safe {
		return u.Repo.BookTicketSafe(ctx, req.EventID, req.UserID)
	}
	return u.Repo.BookTicketUnsafe(ctx, req.EventID, req.UserID)
}

func (u *EventUsecase) GetEvent(ctx context.Context, eventID int) (*domain.Event, error) {
	return u.Repo.GetEvent(ctx, eventID)
}
