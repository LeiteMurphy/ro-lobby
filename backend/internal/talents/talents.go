// Package talents guarda a disponibilidade dos personagens e calcula o banco de talentos
// (spec banco-de-talentos, design D-01 a D-06).
package talents

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/catalog"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// ErrNotFound vale para personagem ou lobby inexistente e para os de outro Usuário
// (RN-05; RN-09 e D-08 da lobbies).
var ErrNotFound = errors.New("talents: não encontrado")

// ErrNotOpen: o lobby já começou ou foi cancelado (RN-09).
var ErrNotOpen = errors.New("talents: lobby não está aberto")

// Campos e códigos dos erros de validação (D-06).
const (
	FieldDays        = "days"
	FieldStart       = "start"
	FieldEnd         = "end"
	FieldInstanceIDs = "instanceIds"
	FieldDay         = "day"
	FieldTime        = "time"

	CodeRequired    = "required"
	CodeInvalid     = "invalid"
	CodeSameAsStart = "same_as_start"
)

// FieldError é um erro num campo.
type FieldError struct {
	Field string
	Code  string
}

// ValidationError junta os erros de campo de um pedido.
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	parts := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		parts[i] = f.Field + ": " + f.Code
	}
	return "talents: " + strings.Join(parts, ", ")
}

// Availability é a disponibilidade de um personagem. Start e End são HH:MM de Brasília
// (D-01); End menor que Start passa da meia-noite (RN-03).
type Availability struct {
	Enabled     bool
	Days        []int
	Start       string
	End         string
	AnyInstance bool
	// InstanceIDs só traz as que continuam no catálogo (RN-04).
	InstanceIDs []string
}

// AvailabilityInput é o que o dono do personagem envia. Desligado, só Enabled conta.
type AvailabilityInput struct {
	Enabled     bool
	Days        []int
	Start       string
	End         string
	AnyInstance bool
	InstanceIDs []string
}

// FromRow converte a linha do banco, sem as instâncias que saíram do catálogo (RN-04).
func FromRow(r db.CharacterAvailability) Availability {
	return Availability{
		Enabled:     r.Enabled,
		Days:        daysOf(r.Days),
		Start:       clock(r.StartMinute),
		End:         clock(r.EndMinute),
		AnyInstance: r.AnyInstance,
		InstanceIDs: inCatalog(r.InstanceIds),
	}
}

type Service struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	// lobbies lê o lobby com o estado calculado pelo mesmo relógio (RN-09).
	lobbies *lobbies.Service
	// Now é o relógio do serviço; os testes trocam por um relógio fixo.
	Now func() time.Time
}

func NewService(pool *pgxpool.Pool) *Service {
	s := &Service{pool: pool, queries: db.New(pool), lobbies: lobbies.NewService(pool), Now: time.Now}
	s.lobbies.Now = func() time.Time { return s.Now() }
	return s
}

// SetAvailability liga, edita ou desliga o personagem no banco (RN-01 a RN-05). Desligar
// um personagem que nunca entrou devolve nil.
func (s *Service) SetAvailability(ctx context.Context, userID, characterID string, in AvailabilityInput) (*Availability, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("talents: id de usuário inválido: %w", err)
	}
	cid, err := parseUUID(characterID)
	if err != nil {
		return nil, ErrNotFound
	}
	var params db.UpsertAvailabilityParams
	if in.Enabled {
		if params, err = validate(in); err != nil {
			return nil, err
		}
		params.CharacterID = cid
	}
	now := s.Now().UTC()

	var out *Availability
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)
		// Trava o Usuário para não correr contra a exclusão do personagem (D-05 da lobbies).
		if _, err := q.LockUser(ctx, uid); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		if _, err := q.GetOwnCharacter(ctx, db.GetOwnCharacterParams{ID: cid, UserID: uid}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound // RN-05: de outro Usuário ou inexistente
			}
			return err
		}
		if !in.Enabled {
			row, err := q.DisableAvailability(ctx, db.DisableAvailabilityParams{CharacterID: cid, Now: now})
			if errors.Is(err, pgx.ErrNoRows) {
				return nil // nunca entrou no banco
			}
			if err != nil {
				return err
			}
			a := FromRow(row)
			out = &a
			return nil
		}
		params.Now = now
		row, err := q.UpsertAvailability(ctx, params)
		if err != nil {
			return err
		}
		a := FromRow(row)
		out = &a
		return nil
	})
	if err != nil {
		return nil, wrap(err, "gravar disponibilidade")
	}
	return out, nil
}

// ByUser devolve a disponibilidade de cada personagem do Usuário, pelo id do personagem,
// para o perfil (CA-01.4).
func ByUser(ctx context.Context, q *db.Queries, uid pgtype.UUID) (map[string]Availability, error) {
	rows, err := q.ListAvailabilityByUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("talents: disponibilidade do usuário: %w", err)
	}
	out := make(map[string]Availability, len(rows))
	for _, r := range rows {
		out[r.CharacterID.String()] = FromRow(r)
	}
	return out, nil
}

// validate confere dias, faixa e instâncias (RN-02 a RN-04) e monta os parâmetros.
func validate(in AvailabilityInput) (db.UpsertAvailabilityParams, error) {
	var errs []FieldError
	var days int16
	if len(in.Days) == 0 {
		errs = append(errs, FieldError{FieldDays, CodeRequired})
	}
	for _, d := range in.Days {
		if d < 0 || d > 6 {
			errs = append(errs, FieldError{FieldDays, CodeInvalid})
			break
		}
		days |= 1 << d
	}
	start, startOK := ParseClock(in.Start)
	if !startOK {
		errs = append(errs, FieldError{FieldStart, CodeInvalid})
	}
	end, endOK := ParseClock(in.End)
	switch {
	case !endOK:
		errs = append(errs, FieldError{FieldEnd, CodeInvalid})
	case startOK && end == start:
		errs = append(errs, FieldError{FieldEnd, CodeSameAsStart})
	}
	ids := []string{}
	if !in.AnyInstance {
		if len(in.InstanceIDs) == 0 {
			errs = append(errs, FieldError{FieldInstanceIDs, CodeRequired})
		}
		for _, id := range in.InstanceIDs {
			if _, ok := catalog.InstanceByID(id); !ok {
				errs = append(errs, FieldError{FieldInstanceIDs, CodeInvalid})
				break
			}
			if !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
	}
	if len(errs) > 0 {
		return db.UpsertAvailabilityParams{}, &ValidationError{Fields: errs}
	}
	return db.UpsertAvailabilityParams{
		Days: days, StartMinute: start, EndMinute: end, AnyInstance: in.AnyInstance, InstanceIds: ids,
	}, nil
}

// ParseClock lê HH:MM de 30 em 30 minutos e devolve os minutos desde a meia-noite (RN-03).
func ParseClock(s string) (int16, bool) {
	h, m, ok := strings.Cut(s, ":")
	if !ok || len(h) != 2 || len(m) != 2 {
		return 0, false
	}
	hh, err1 := strconv.Atoi(h)
	mm, err2 := strconv.Atoi(m)
	if err1 != nil || err2 != nil || hh < 0 || hh > 23 || (mm != 0 && mm != 30) {
		return 0, false
	}
	return int16(hh*60 + mm), true //nolint:gosec // 0..1410
}

func clock(minutes int16) string {
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}

func daysOf(mask int16) []int {
	out := []int{}
	for d := range 7 {
		if mask&(1<<d) != 0 {
			out = append(out, d)
		}
	}
	return out
}

func inCatalog(ids []string) []string {
	out := []string{}
	for _, id := range ids {
		if _, ok := catalog.InstanceByID(id); ok {
			out = append(out, id)
		}
	}
	return out
}

func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	err := u.Scan(s)
	return u, err
}

func wrap(err error, action string) error {
	var ve *ValidationError
	if errors.As(err, &ve) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrNotOpen) {
		return err
	}
	return fmt.Errorf("talents: %s: %w", action, err)
}
