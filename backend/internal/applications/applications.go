// Package applications guarda as regras da candidatura a lobby (spec candidatura-lobby,
// RN-01 a RN-18 e RN-30, design D-01 a D-05 e D-07).
package applications

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// Estados da candidatura (RN-17). Os da Parte 2 (left, removed, cancelled) ainda não
// têm transição.
const (
	StatusPending   = "pending"
	StatusAccepted  = "accepted"
	StatusRejected  = "rejected"
	StatusWithdrawn = "withdrawn"
	StatusExpired   = "expired"

	// MaxMessageLength é o tamanho máximo da mensagem do candidato (RN-06).
	MaxMessageLength = 250

	FieldMessage = "message"
)

// Códigos das regras de negócio (D-07). O web traduz para pt-BR.
const (
	CodeNotOpen          = "not_open"
	CodeOwnLobby         = "own_lobby"
	CodeAlreadyActive    = "already_active"
	CodeRoleFull         = "role_full"
	CodeRejectedBefore   = "rejected_before"
	CodeBelowMinLevel    = "below_min_level"
	CodeScheduleConflict = "schedule_conflict"
	CodeNotPending       = "not_pending"
	CodeNotOwner         = "not_owner"
	CodeNotYours         = "not_yours"
)

// ErrNotFound vale para lobby ou candidatura inexistente.
var ErrNotFound = errors.New("applications: não encontrado")

// RuleError é uma regra de negócio violada; a API responde 409 com o código (D-07).
type RuleError struct {
	Code string
}

func (e *RuleError) Error() string { return "applications: regra violada: " + e.Code }

func rule(code string) error { return &RuleError{Code: code} }

// Application é uma candidatura com o estado efetivo (D-02).
type Application struct {
	ID          string
	LobbyID     string
	UserID      string
	CharacterID string
	Role        string
	Message     string
	Status      string
	Reason      string
	CreatedAt   time.Time
	// DecidedAt fica zerado enquanto a candidatura está pendente.
	DecidedAt time.Time
}

// Mine é uma candidatura do próprio Usuário, com o lobby e o personagem (RN-33).
type Mine struct {
	Application
	InstanceName string
	StartsAt     time.Time
	LobbyStatus  string
	// Os dados do personagem ficam vazios se ele foi excluído depois do lobby (RN-26).
	Nick     string
	ClassID  string
	Level    int
	Portrait string
}

// ApplyInput é o que o candidato informa.
type ApplyInput struct {
	CharacterID string
	Message     string
}

type Service struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	// Now é o relógio do serviço; os testes trocam por um relógio fixo.
	Now func() time.Time
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, queries: db.New(pool), Now: time.Now}
}

// Apply cria a candidatura pendente do Usuário no lobby (RN-01 a RN-07, RN-30).
func (s *Service) Apply(ctx context.Context, userID, lobbyID string, in ApplyInput) (Application, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return Application{}, fmt.Errorf("applications: id de usuário inválido: %w", err)
	}
	lid, err := parseUUID(lobbyID)
	if err != nil {
		return Application{}, ErrNotFound
	}
	var errs []lobbies.FieldError
	message := strings.TrimSpace(in.Message)
	if utf8.RuneCountInString(message) > MaxMessageLength {
		errs = append(errs, lobbies.FieldError{Field: FieldMessage, Code: lobbies.CodeTooLong})
	}
	cid, cidErr := parseUUID(in.CharacterID)
	if in.CharacterID == "" {
		errs = append(errs, lobbies.FieldError{Field: lobbies.FieldCharacterID, Code: lobbies.CodeRequired})
	} else if cidErr != nil {
		errs = append(errs, lobbies.FieldError{Field: lobbies.FieldCharacterID, Code: lobbies.CodeInvalid})
	}
	if len(errs) > 0 {
		return Application{}, &lobbies.ValidationError{Fields: errs}
	}
	now := s.Now().UTC()

	var created db.Application
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)
		if err := lockUser(ctx, q, uid); err != nil { // D-05: RN-02 sob concorrência
			return err
		}
		// D-05: trava o lobby depois do Usuário, para a candidatura não passar junto com o
		// cancelamento (RN-16).
		if _, err := q.LockLobby(ctx, lid); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		lobby, err := q.GetLobby(ctx, lid)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if !isOpen(lobby.Lobby, now) {
			return rule(CodeNotOpen)
		}
		if lobby.Lobby.OwnerID == uid {
			return rule(CodeOwnLobby) // RN-03
		}
		character, err := q.GetOwnCharacter(ctx, db.GetOwnCharacterParams{ID: cid, UserID: uid})
		if errors.Is(err, pgx.ErrNoRows) {
			// RN-01: de outro Usuário ou inexistente.
			return &lobbies.ValidationError{Fields: []lobbies.FieldError{{Field: lobbies.FieldCharacterID, Code: lobbies.CodeInvalid}}}
		}
		if err != nil {
			return err
		}
		active, err := q.HasActiveApplication(ctx, db.HasActiveApplicationParams{LobbyID: lid, UserID: uid})
		if err != nil {
			return err
		}
		if active {
			return rule(CodeAlreadyActive) // RN-02
		}
		rejected, err := q.WasRejected(ctx, db.WasRejectedParams{LobbyID: lid, UserID: uid})
		if err != nil {
			return err
		}
		if rejected {
			return rule(CodeRejectedBefore) // RN-07
		}
		if character.Level < lobby.Lobby.MinLevel {
			return rule(CodeBelowMinLevel) // RN-30
		}
		if freeSlots(lobby, character.Role) < 1 {
			return rule(CodeRoleFull) // RN-05
		}
		created, err = q.CreateApplication(ctx, db.CreateApplicationParams{
			LobbyID:     lid,
			UserID:      uid,
			CharacterID: cid,
			Role:        character.Role, // RN-04
			Message:     optionalText(message),
			Now:         now,
		})
		if err != nil {
			return err
		}
		return q.InsertApplicationEvent(ctx, db.InsertApplicationEventParams{
			ApplicationID: created.ID, ToStatus: StatusPending, ActorID: uid, Reason: optionalText(message), Now: now,
		})
	})
	if err != nil {
		return Application{}, wrap(err, "candidatar")
	}
	return toApplication(created), nil
}

// Accept aceita a candidatura pendente, se o dono pedir e ainda houver vaga, nível e
// horário livre (RN-08, RN-10, RN-11, RN-12, RN-30).
func (s *Service) Accept(ctx context.Context, ownerID, applicationID string) (Application, error) {
	return s.decide(ctx, ownerID, applicationID, func(q *db.Queries, app db.Application, lobby db.GetLobbyRow, now time.Time) error {
		character, err := q.GetOwnCharacter(ctx, db.GetOwnCharacterParams{ID: app.CharacterID, UserID: app.UserID})
		if errors.Is(err, pgx.ErrNoRows) {
			return rule(CodeNotPending) // RN-26 impede excluir com pendência; só por segurança
		}
		if err != nil {
			return err
		}
		if character.Level < lobby.Lobby.MinLevel {
			return rule(CodeBelowMinLevel)
		}
		if freeSlots(lobby, app.Role) < 1 {
			return rule(CodeRoleFull)
		}
		conflict, err := q.HasScheduleConflict(ctx, db.HasScheduleConflictParams{
			CharacterID: app.CharacterID,
			ExcludeID:   app.LobbyID,
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
	}, StatusAccepted, "")
}

// Reject recusa a candidatura pendente com justificativa de 10 a 250 caracteres (RN-08,
// RN-09).
func (s *Service) Reject(ctx context.Context, ownerID, applicationID, reason string) (Application, error) {
	reason = strings.TrimSpace(reason)
	switch n := utf8.RuneCountInString(reason); {
	case n == 0:
		return Application{}, fieldError(lobbies.FieldReason, lobbies.CodeRequired)
	case n < lobbies.MinReasonLength:
		return Application{}, fieldError(lobbies.FieldReason, lobbies.CodeTooShort)
	case n > lobbies.MaxReasonLength:
		return Application{}, fieldError(lobbies.FieldReason, lobbies.CodeTooLong)
	}
	return s.decide(ctx, ownerID, applicationID, nil, StatusRejected, reason)
}

// decide trava, na ordem do D-05, o Usuário do candidato, o lobby e a candidatura; confere
// que quem pede é o dono e que ela está pendente; roda check e grava o novo estado.
func (s *Service) decide(ctx context.Context, ownerID, applicationID string,
	check func(*db.Queries, db.Application, db.GetLobbyRow, time.Time) error, to, reason string,
) (Application, error) {
	uid, err := parseUUID(ownerID)
	if err != nil {
		return Application{}, fmt.Errorf("applications: id de usuário inválido: %w", err)
	}
	aid, err := parseUUID(applicationID)
	if err != nil {
		return Application{}, ErrNotFound
	}
	now := s.Now().UTC()

	var decided db.Application
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)
		peek, err := q.GetApplication(ctx, aid)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if err := lockUser(ctx, q, peek.UserID); err != nil {
			return err
		}
		if _, err := q.LockLobby(ctx, peek.LobbyID); err != nil {
			return err
		}
		app, err := q.GetApplicationForUpdate(ctx, aid)
		if err != nil {
			return err
		}
		lobby, err := q.GetLobby(ctx, app.LobbyID)
		if err != nil {
			return err
		}
		if lobby.Lobby.OwnerID != uid {
			return rule(CodeNotOwner) // RN-08
		}
		if effectiveStatus(app.Status, lobby.Lobby, now) != StatusPending {
			return rule(CodeNotPending) // RN-08, RN-16, RN-17
		}
		if check != nil {
			if err := check(q, app, lobby, now); err != nil {
				return err
			}
		}
		decided, err = transition(ctx, q, app, to, uid, reason, now)
		return err
	})
	if err != nil {
		return Application{}, wrap(err, "decidir")
	}
	return toApplication(decided), nil
}

// Withdraw retira a candidatura pendente do próprio Usuário (RN-13).
func (s *Service) Withdraw(ctx context.Context, userID, applicationID string) (Application, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return Application{}, fmt.Errorf("applications: id de usuário inválido: %w", err)
	}
	aid, err := parseUUID(applicationID)
	if err != nil {
		return Application{}, ErrNotFound
	}
	now := s.Now().UTC()

	var withdrawn db.Application
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
		if effectiveStatus(app.Status, lobby.Lobby, now) != StatusPending {
			return rule(CodeNotPending) // RN-17
		}
		withdrawn, err = transition(ctx, q, app, StatusWithdrawn, uid, "", now)
		return err
	})
	if err != nil {
		return Application{}, wrap(err, "retirar")
	}
	return toApplication(withdrawn), nil
}

// ListMine devolve as candidaturas do Usuário, as mais recentes primeiro (RN-33, CA-03.1).
func (s *Service) ListMine(ctx context.Context, userID string) ([]Mine, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("applications: id de usuário inválido: %w", err)
	}
	rows, err := s.queries.ListMyApplications(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("applications: listar: %w", err)
	}
	now := s.Now()
	out := make([]Mine, len(rows))
	for i, r := range rows {
		lobby := db.Lobby{StartsAt: r.StartsAt, CancelledAt: r.LobbyCancelledAt}
		app := toApplication(r.Application)
		app.Status = effectiveStatus(app.Status, lobby, now)
		out[i] = Mine{
			Application:  app,
			InstanceName: r.InstanceName,
			StartsAt:     r.StartsAt.UTC(),
			LobbyStatus:  lobbyStatus(lobby, now),
			Nick:         r.Nick.String,
			ClassID:      r.ClassID.String,
			Level:        int(r.Level.Int16),
			Portrait:     r.Portrait.String,
		}
	}
	return out, nil
}

// transition grava o novo estado e o evento do histórico na mesma transação (D-01, RN-18).
func transition(ctx context.Context, q *db.Queries, app db.Application, to string, actor pgtype.UUID, reason string, now time.Time) (db.Application, error) {
	updated, err := q.SetApplicationStatus(ctx, db.SetApplicationStatusParams{
		ID: app.ID, Status: to, Reason: optionalText(reason), Now: now,
	})
	if err != nil {
		return db.Application{}, err
	}
	err = q.InsertApplicationEvent(ctx, db.InsertApplicationEventParams{
		ApplicationID: app.ID,
		FromStatus:    pgtype.Text{String: app.Status, Valid: true},
		ToStatus:      to,
		ActorID:       actor,
		Reason:        optionalText(reason),
		Now:           now,
	})
	return updated, err
}

// effectiveStatus é o estado lido da candidatura: pendente de lobby iniciado conta como
// expirada (D-02, RN-16).
func effectiveStatus(status string, l db.Lobby, now time.Time) string {
	return lobbies.ApplicationStatus(status, l.StartsAt, now)
}

func isOpen(l db.Lobby, now time.Time) bool {
	return !l.CancelledAt.Valid && l.StartsAt.After(now)
}

func lobbyStatus(l db.Lobby, now time.Time) string {
	switch {
	case l.CancelledAt.Valid:
		return lobbies.StatusCancelled
	case !l.StartsAt.After(now):
		return lobbies.StatusStarted
	default:
		return lobbies.StatusOpen
	}
}

// freeSlots são as vagas da função menos os ocupantes: o dono e os aceitos (RN-05, D-03).
func freeSlots(l db.GetLobbyRow, role string) int {
	var slots, accepted int
	switch role {
	case "tank":
		slots, accepted = int(l.Lobby.SlotsTank), int(l.AcceptedTank)
	case "support":
		slots, accepted = int(l.Lobby.SlotsSupport), int(l.AcceptedSupport)
	default:
		slots, accepted = int(l.Lobby.SlotsDps), int(l.AcceptedDps)
	}
	if l.Lobby.OwnerRole == role {
		accepted++
	}
	return slots - accepted
}

func lockUser(ctx context.Context, q *db.Queries, uid pgtype.UUID) error {
	if _, err := q.LockUser(ctx, uid); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func wrap(err error, action string) error {
	var re *RuleError
	var ve *lobbies.ValidationError
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &re), errors.As(err, &ve), errors.Is(err, ErrNotFound):
		return err
	case errors.As(err, &pgErr) && pgErr.ConstraintName == "applications_one_active_per_user":
		return rule(CodeAlreadyActive) // RN-02: última defesa
	default:
		return fmt.Errorf("applications: %s: %w", action, err)
	}
}

func fieldError(field, code string) error {
	return &lobbies.ValidationError{Fields: []lobbies.FieldError{{Field: field, Code: code}}}
}

func toApplication(a db.Application) Application {
	out := Application{
		ID:        a.ID.String(),
		LobbyID:   a.LobbyID.String(),
		UserID:    a.UserID.String(),
		Role:      a.Role,
		Message:   a.Message.String,
		Status:    a.Status,
		Reason:    a.Reason.String,
		CreatedAt: a.CreatedAt.UTC(),
	}
	if a.CharacterID.Valid {
		out.CharacterID = a.CharacterID.String()
	}
	if a.DecidedAt.Valid {
		out.DecidedAt = a.DecidedAt.Time.UTC()
	}
	return out
}

func parseUUID(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(s)
	return id, err
}

func optionalText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}
