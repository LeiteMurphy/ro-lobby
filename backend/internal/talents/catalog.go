package talents

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/catalog"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// MaxCatalog é o máximo de personagens numa resposta do catálogo (risco do design).
const MaxCatalog = 100

// Campos dos filtros e da contagem.
const (
	FieldInstanceID  = "instanceId"
	FieldRole        = "role"
	FieldCharacterID = "characterId"
)

var roles = []string{"tank", "support", "dps"}

// CatalogFilter são os filtros opcionais de /talentos (RN-11). Day é o dia da semana
// (domingo = 0) e Time, HH:MM de Brasília.
type CatalogFilter struct {
	InstanceID string
	Role       string
	Day        *int
	Time       string
}

// Catalog lista o banco com os filtros, por nível decrescente, até MaxCatalog (RN-08,
// RN-11). O nome no Discord vem sempre; quem chama decide se ele sai (D-05).
func (s *Service) Catalog(ctx context.Context, f CatalogFilter) ([]Talent, error) {
	var errs []FieldError
	params := db.CatalogParams{MaxRows: MaxCatalog}
	if f.InstanceID != "" {
		if _, ok := catalog.InstanceByID(f.InstanceID); !ok {
			errs = append(errs, FieldError{FieldInstanceID, CodeInvalid})
		}
		params.InstanceID = pgtype.Text{String: f.InstanceID, Valid: true}
	}
	if f.Role != "" {
		if !slices.Contains(roles, f.Role) {
			errs = append(errs, FieldError{FieldRole, CodeInvalid})
		}
		params.Role = pgtype.Text{String: f.Role, Valid: true}
	}
	if f.Day != nil {
		if *f.Day < 0 || *f.Day > 6 {
			errs = append(errs, FieldError{FieldDay, CodeInvalid})
		}
		params.Dow = pgtype.Int4{Int32: int32(*f.Day), Valid: true} //nolint:gosec // 0..6
	}
	if f.Time != "" {
		minute, ok := ParseClock(f.Time)
		if !ok {
			errs = append(errs, FieldError{FieldTime, CodeInvalid})
		}
		params.Minute = pgtype.Int4{Int32: int32(minute), Valid: true}
	}
	if len(errs) > 0 {
		return nil, &ValidationError{Fields: errs}
	}
	rows, err := s.queries.Catalog(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("talents: catálogo: %w", err)
	}
	out := make([]Talent, len(rows))
	for i, r := range rows {
		out[i] = toTalent(r)
	}
	return out, nil
}

// CountInput são os campos do formulário de criação do lobby (RN-10).
type CountInput struct {
	InstanceID  string
	StartsAt    time.Time
	MinLevel    int
	Tank        int
	Support     int
	Dps         int
	CharacterID string
	// Formation e FreeSlots: no grupo livre, o total de vagas do formulário (RN-13 da
	// grupo-livre); Tank, Support e Dps ficam de fora.
	Formation string
	FreeSlots int
}

// Count conta os personagens com afinidade com o lobby em criação (RN-10, D-04). As
// vagas abertas são as do formulário menos a do personagem do dono.
func (s *Service) Count(ctx context.Context, userID string, in CountInput) (int, error) {
	owner, err := parseUUID(userID)
	if err != nil {
		return 0, fmt.Errorf("talents: id de usuário inválido: %w", err)
	}
	var errs []FieldError
	if _, ok := catalog.InstanceByID(in.InstanceID); !ok {
		errs = append(errs, FieldError{FieldInstanceID, CodeInvalid})
	}
	cid, cidErr := parseUUID(in.CharacterID)
	if cidErr != nil {
		errs = append(errs, FieldError{FieldCharacterID, CodeInvalid})
	}
	if len(errs) > 0 {
		return 0, &ValidationError{Fields: errs}
	}
	character, err := s.queries.GetOwnCharacter(ctx, db.GetOwnCharacterParams{ID: cid, UserID: owner})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, &ValidationError{Fields: []FieldError{{FieldCharacterID, CodeInvalid}}}
	}
	if err != nil {
		return 0, fmt.Errorf("talents: personagem do dono: %w", err)
	}
	open := map[string]int{"tank": in.Tank, "support": in.Support, "dps": in.Dps}
	open[character.Role]--
	var openRoles []string
	for _, role := range roles {
		if in.Formation == lobbies.FormationFree {
			if in.FreeSlots-1 > 0 { // a vaga do dono
				openRoles = append(openRoles, role)
			}
		} else if open[role] > 0 {
			openRoles = append(openRoles, role)
		}
	}
	list, err := s.affinity(ctx, Probe{
		InstanceID: in.InstanceID, StartsAt: in.StartsAt, MinLevel: in.MinLevel, Roles: openRoles, Owner: owner,
	})
	return len(list), err
}
