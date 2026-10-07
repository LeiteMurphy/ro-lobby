package applications

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
)

// cancelPendingSwap cancela o pedido de troca pendente da candidatura, se houver, quando o
// membro sai ou é removido (RN-24, D-10).
func cancelPendingSwap(ctx context.Context, q *db.Queries, applicationID, actor pgtype.UUID, now time.Time) error {
	req, err := q.GetPendingSwapRequest(ctx, applicationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = transitionSwap(ctx, q, req, StatusCancelled, actor, "", now)
	return err
}

// transitionSwap grava o novo estado do pedido e o evento do histórico na mesma transação
// (D-08, RN-18).
func transitionSwap(ctx context.Context, q *db.Queries, req db.SwapRequest, to string, actor pgtype.UUID, reason string, now time.Time) (db.SwapRequest, error) {
	updated, err := q.SetSwapRequestStatus(ctx, db.SetSwapRequestStatusParams{
		ID: req.ID, Status: to, DecisionReason: optionalText(reason), Now: now,
	})
	if err != nil {
		return db.SwapRequest{}, err
	}
	err = q.InsertSwapRequestEvent(ctx, db.InsertSwapRequestEventParams{
		SwapRequestID: req.ID,
		FromStatus:    pgtype.Text{String: req.Status, Valid: true},
		ToStatus:      to,
		ActorID:       actor,
		Reason:        optionalText(reason),
		Now:           now,
	})
	return updated, err
}
