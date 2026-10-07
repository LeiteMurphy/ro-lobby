package applications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// SwapRequest é um pedido de troca de personagem do membro (D-08); o tipo é o mesmo do
// detalhe do lobby.
type SwapRequest = lobbies.SwapRequest

// SwapInput é o que o membro informa no pedido de troca.
type SwapInput struct {
	CharacterID string
	Reason      string
}

// SwapOwnerCharacter troca o personagem do dono no próprio lobby, sem aprovação, se o
// personagem novo tiver vaga na função, o nível mínimo e o horário livre (RN-11, RN-19,
// RN-36, D-10, D-11).
func (s *Service) SwapOwnerCharacter(ctx context.Context, ownerID, lobbyID, characterID string) error {
	uid, err := parseUUID(ownerID)
	if err != nil {
		return fmt.Errorf("applications: id de usuário inválido: %w", err)
	}
	lid, err := parseUUID(lobbyID)
	if err != nil {
		return ErrNotFound
	}
	cid, err := characterField(characterID)
	if err != nil {
		return err
	}
	now := s.Now().UTC()

	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)
		if err := lockUser(ctx, q, uid); err != nil {
			return err
		}
		if _, err := q.LockLobby(ctx, lid); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		lobby, err := q.GetLobby(ctx, lid)
		if err != nil {
			return err
		}
		if lobby.Lobby.OwnerID != uid {
			return rule(CodeNotOwner) // CA-07.5
		}
		if !isOpen(lobby.Lobby, now) {
			return rule(CodeNotOpen)
		}
		character, err := ownCharacter(ctx, q, cid, uid)
		if err != nil {
			return err
		}
		if character.ID == lobby.Lobby.OwnerCharacterID {
			return nil // já é o personagem do dono
		}
		if err := checkSwap(ctx, q, lobby, character, lobby.Lobby.OwnerRole); err != nil {
			return err
		}
		return q.SetLobbyOwnerCharacter(ctx, db.SetLobbyOwnerCharacterParams{
			ID: lid, CharacterID: character.ID, Role: character.Role,
		})
	})
	if err != nil {
		return wrap(err, "trocar o personagem do dono")
	}
	return nil
}

// RequestSwap abre o pedido de troca do membro; ele continua no grupo com o personagem
// atual até o dono decidir (RN-20, RN-21, RN-36, D-10, D-11).
func (s *Service) RequestSwap(ctx context.Context, userID, applicationID string, in SwapInput) (SwapRequest, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return SwapRequest{}, fmt.Errorf("applications: id de usuário inválido: %w", err)
	}
	aid, err := parseUUID(applicationID)
	if err != nil {
		return SwapRequest{}, ErrNotFound
	}
	var errs []lobbies.FieldError
	reason, reasonErr := validReason(in.Reason)
	var ve *lobbies.ValidationError
	if errors.As(reasonErr, &ve) {
		errs = append(errs, ve.Fields...)
	}
	cid, cidErr := characterField(in.CharacterID)
	if errors.As(cidErr, &ve) {
		errs = append(errs, ve.Fields...)
	}
	if len(errs) > 0 {
		return SwapRequest{}, &lobbies.ValidationError{Fields: errs}
	}
	now := s.Now().UTC()

	var created db.SwapRequest
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)
		if err := lockUser(ctx, q, uid); err != nil {
			return err
		}
		app, err := q.GetApplicationForUpdate(ctx, aid)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if app.UserID != uid {
			return rule(CodeNotYours)
		}
		lobby, err := q.GetLobby(ctx, app.LobbyID)
		if err != nil {
			return err
		}
		if !isOpen(lobby.Lobby, now) {
			return rule(CodeNotOpen)
		}
		if app.Status != StatusAccepted {
			return rule(CodeNotMember)
		}
		if _, err := q.GetPendingSwapRequest(ctx, app.ID); err == nil {
			return rule(CodeSwapPending) // RN-20
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		character, err := ownCharacter(ctx, q, cid, uid)
		if err != nil {
			return err
		}
		if character.ID == app.CharacterID {
			return fieldError(lobbies.FieldCharacterID, lobbies.CodeInvalid) // já é o personagem atual
		}
		if character.Level < lobby.Lobby.MinLevel {
			return rule(CodeBelowMinLevel) // RN-36; a vaga só no aceite (D-11)
		}
		created, err = q.CreateSwapRequest(ctx, db.CreateSwapRequestParams{
			ApplicationID:   app.ID,
			FromCharacterID: app.CharacterID,
			ToCharacterID:   character.ID,
			ToRole:          character.Role,
			Reason:          reason,
			Now:             now,
		})
		if err != nil {
			return err
		}
		return q.InsertSwapRequestEvent(ctx, db.InsertSwapRequestEventParams{
			SwapRequestID: created.ID, ToStatus: StatusPending, ActorID: uid, Reason: optionalText(reason), Now: now,
		})
	})
	if err != nil {
		return SwapRequest{}, wrapSwap(err, "pedir troca")
	}
	return toSwapRequest(created), nil
}

// AcceptSwap aceita o pedido: o membro passa a ocupar a vaga da função do personagem novo
// e a vaga antiga fica livre, de uma vez (RN-22, RN-23, RN-36, D-10, D-11).
func (s *Service) AcceptSwap(ctx context.Context, ownerID, swapID string) (SwapRequest, error) {
	return s.decideSwap(ctx, ownerID, swapID, func(q *db.Queries, req db.SwapRequest, app db.Application, lobby db.GetLobbyRow) error {
		character, err := q.GetOwnCharacter(ctx, db.GetOwnCharacterParams{ID: req.ToCharacterID, UserID: app.UserID})
		if errors.Is(err, pgx.ErrNoRows) {
			return rule(CodeNotPending) // RN-26 impede excluir com pedido pendente; só por segurança
		}
		if err != nil {
			return err
		}
		if err := checkSwap(ctx, q, lobby, character, app.Role); err != nil {
			return err
		}
		_, err = q.SetApplicationCharacter(ctx, db.SetApplicationCharacterParams{
			ID: app.ID, CharacterID: character.ID, Role: character.Role,
		})
		return err
	}, StatusAccepted, "")
}

// RejectSwap recusa o pedido com justificativa; o membro continua com o personagem atual
// (RN-09, RN-22).
func (s *Service) RejectSwap(ctx context.Context, ownerID, swapID, reason string) (SwapRequest, error) {
	reason, err := validReason(reason)
	if err != nil {
		return SwapRequest{}, err
	}
	return s.decideSwap(ctx, ownerID, swapID, nil, StatusRejected, reason)
}

// decideSwap trava, na ordem do D-10, o Usuário do membro, o lobby, o pedido e a
// candidatura; confere que quem pede é o dono e que o pedido está pendente; roda check e
// grava o novo estado.
func (s *Service) decideSwap(ctx context.Context, ownerID, swapID string,
	check func(*db.Queries, db.SwapRequest, db.Application, db.GetLobbyRow) error, to, reason string,
) (SwapRequest, error) {
	uid, err := parseUUID(ownerID)
	if err != nil {
		return SwapRequest{}, fmt.Errorf("applications: id de usuário inválido: %w", err)
	}
	sid, err := parseUUID(swapID)
	if err != nil {
		return SwapRequest{}, ErrNotFound
	}
	now := s.Now().UTC()

	var decided db.SwapRequest
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)
		peek, err := q.GetSwapRequest(ctx, sid)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		app, lobby, err := lockApplication(ctx, q, peek.ApplicationID)
		if err != nil {
			return err
		}
		req, err := q.GetSwapRequestForUpdate(ctx, sid)
		if err != nil {
			return err
		}
		if lobby.Lobby.OwnerID != uid {
			return rule(CodeNotOwner) // RN-22
		}
		if effectiveStatus(req.Status, lobby.Lobby, now) != StatusPending {
			return rule(CodeNotPending) // RN-24, D-12
		}
		if check != nil {
			if err := check(q, req, app, lobby); err != nil {
				return err
			}
		}
		decided, err = transitionSwap(ctx, q, req, to, uid, reason, now)
		return err
	})
	if err != nil {
		return SwapRequest{}, wrapSwap(err, "decidir troca")
	}
	return toSwapRequest(decided), nil
}

// WithdrawSwap retira o pedido pendente do próprio membro; ele continua com o personagem
// atual e pode abrir outro pedido (RN-27).
func (s *Service) WithdrawSwap(ctx context.Context, userID, swapID string) (SwapRequest, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return SwapRequest{}, fmt.Errorf("applications: id de usuário inválido: %w", err)
	}
	sid, err := parseUUID(swapID)
	if err != nil {
		return SwapRequest{}, ErrNotFound
	}
	now := s.Now().UTC()

	var withdrawn db.SwapRequest
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)
		if err := lockUser(ctx, q, uid); err != nil {
			return err
		}
		req, err := q.GetSwapRequestForUpdate(ctx, sid)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		app, err := q.GetApplication(ctx, req.ApplicationID)
		if err != nil {
			return err
		}
		if app.UserID != uid {
			return rule(CodeNotYours) // CA-08.12
		}
		lobby, err := q.GetLobby(ctx, app.LobbyID)
		if err != nil {
			return err
		}
		if effectiveStatus(req.Status, lobby.Lobby, now) != StatusPending {
			return rule(CodeNotPending)
		}
		withdrawn, err = transitionSwap(ctx, q, req, StatusWithdrawn, uid, "", now)
		return err
	})
	if err != nil {
		return SwapRequest{}, wrapSwap(err, "retirar troca")
	}
	return toSwapRequest(withdrawn), nil
}

// checkSwap confere o personagem novo de uma troca: nível mínimo, vaga na função (a vaga
// que sai conta como livre) e conflito de horário fora deste lobby (RN-11, RN-19, RN-23,
// RN-36, D-11).
func checkSwap(ctx context.Context, q *db.Queries, lobby db.GetLobbyRow, character db.Character, currentRole string) error {
	if character.Level < lobby.Lobby.MinLevel {
		return rule(CodeBelowMinLevel)
	}
	if character.Role != currentRole && freeSlots(lobby, character.Role) < 1 {
		return rule(CodeRoleFull)
	}
	conflict, err := q.HasScheduleConflict(ctx, db.HasScheduleConflictParams{
		CharacterID: character.ID,
		ExcludeID:   lobby.Lobby.ID,
		WindowStart: lobby.Lobby.StartsAt.Add(-lobbies.ConflictWindow).UTC(),
		WindowEnd:   lobby.Lobby.StartsAt.Add(lobbies.ConflictWindow).UTC(),
	})
	if err != nil {
		return err
	}
	if conflict {
		return rule(CodeScheduleConflict)
	}
	return nil
}

// ownCharacter devolve o personagem do próprio Usuário; de outro Usuário ou inexistente é
// erro de campo (RN-20, CA-08.4).
func ownCharacter(ctx context.Context, q *db.Queries, cid, uid pgtype.UUID) (db.Character, error) {
	character, err := q.GetOwnCharacter(ctx, db.GetOwnCharacterParams{ID: cid, UserID: uid})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.Character{}, fieldError(lobbies.FieldCharacterID, lobbies.CodeInvalid)
	}
	return character, err
}

func characterField(characterID string) (pgtype.UUID, error) {
	if characterID == "" {
		return pgtype.UUID{}, fieldError(lobbies.FieldCharacterID, lobbies.CodeRequired)
	}
	cid, err := parseUUID(characterID)
	if err != nil {
		return pgtype.UUID{}, fieldError(lobbies.FieldCharacterID, lobbies.CodeInvalid)
	}
	return cid, nil
}

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

// wrapSwap é o wrap com o índice de um pedido pendente por membro como última defesa
// (RN-20).
func wrapSwap(err error, action string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.ConstraintName == "swap_requests_one_pending" {
		return rule(CodeSwapPending)
	}
	return wrap(err, action)
}

func toSwapRequest(r db.SwapRequest) SwapRequest {
	return lobbies.NewSwapRequest(r)
}
