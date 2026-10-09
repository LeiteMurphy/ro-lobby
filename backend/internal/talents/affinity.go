package talents

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/catalog"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// window é a janela de conflito de horário, a mesma da lobbies (D-03 da lobbies).
const window = 2 * time.Hour

// Instance é uma instância de interesse, com o nome do catálogo.
type Instance struct {
	ID   string
	Name string
}

// Talent é um personagem do banco, como aparece nas listas (RN-12). DiscordUsername é o
// username do dono do personagem; quem chama decide se ele sai (D-05).
type Talent struct {
	CharacterID     string
	Nick            string
	ClassID         string
	Level           int
	Role            string
	Portrait        string
	Link            string
	Days            []int
	Start           string
	End             string
	AnyInstance     bool
	Instances       []Instance
	DiscordUsername string
	// Removed e Blocked: a pessoa já foi removida do lobby, e se com bloqueio (RN-13). Só
	// a afinidade de um lobby gravado preenche.
	Removed bool
	Blocked bool
}

// Probe é o lobby, gravado ou em criação, contra o qual se procura afinidade (D-04).
type Probe struct {
	InstanceID string
	StartsAt   time.Time
	MinLevel   int
	// Roles são as funções com vaga aberta.
	Roles []string
	// Owner é o Usuário dono do lobby; os personagens dele não entram (RN-07).
	Owner pgtype.UUID
	// Lobby é o lobby gravado, para o selo de removido (RN-13); vazio na criação.
	Lobby pgtype.UUID
}

// ForLobby devolve os personagens com afinidade com o lobby aberto do Usuário (RN-06 a
// RN-09). Lobby de outro Usuário é ErrNotFound; iniciado ou cancelado, ErrNotOpen.
func (s *Service) ForLobby(ctx context.Context, userID, lobbyID string) ([]Talent, error) {
	l, err := s.lobbies.Get(ctx, lobbyID)
	if errors.Is(err, lobbies.ErrNotFound) || (err == nil && l.Owner.UserID != userID) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("talents: lobby: %w", err)
	}
	if l.Status != lobbies.StatusOpen {
		return nil, ErrNotOpen
	}
	owner, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("talents: id de usuário inválido: %w", err)
	}
	var roles []string
	for role, open := range map[string]bool{
		"tank":    l.Slots.Tank > l.Occupied.Tank,
		"support": l.Slots.Support > l.Occupied.Support,
		"dps":     l.Slots.Dps > l.Occupied.Dps,
	} {
		if open {
			roles = append(roles, role)
		}
	}
	lid, err := parseUUID(l.ID)
	if err != nil {
		return nil, fmt.Errorf("talents: id do lobby: %w", err)
	}
	return s.affinity(ctx, Probe{
		InstanceID: l.InstanceID, StartsAt: l.StartsAt, MinLevel: l.MinLevel, Roles: roles, Owner: owner,
		Lobby: lid,
	})
}

// affinity roda a consulta da RN-06 para o Probe (D-04).
func (s *Service) affinity(ctx context.Context, p Probe) ([]Talent, error) {
	if len(p.Roles) == 0 {
		return []Talent{}, nil // sem vaga aberta, ninguém tem afinidade
	}
	dow, minute := wallClock(p.StartsAt)
	rows, err := s.queries.Affinity(ctx, db.AffinityParams{
		ExcludeUserID: p.Owner,
		InstanceID:    p.InstanceID,
		Dow:           int32(dow),        //nolint:gosec // 0..6
		Minute:        int32(minute),     //nolint:gosec // 0..1439
		MinLevel:      int32(p.MinLevel), //nolint:gosec // 1..275
		Roles:         p.Roles,
		WindowStart:   p.StartsAt.Add(-window),
		WindowEnd:     p.StartsAt.Add(window),
		LobbyID:       p.Lobby,
	})
	if err != nil {
		return nil, fmt.Errorf("talents: afinidade: %w", err)
	}
	out := make([]Talent, len(rows))
	for i, r := range rows {
		out[i] = toTalent(db.CatalogRow{
			ID: r.ID, Nick: r.Nick, ClassID: r.ClassID, Level: r.Level, Role: r.Role, Portrait: r.Portrait,
			Link: r.Link, Days: r.Days, StartMinute: r.StartMinute, EndMinute: r.EndMinute,
			AnyInstance: r.AnyInstance, InstanceIds: r.InstanceIds, Username: r.Username,
		})
		out[i].Removed, out[i].Blocked = r.Removed, r.Blocked
	}
	return out, nil
}

// wallClock é o dia da semana e o minuto do dia em Brasília (D-01).
func wallClock(t time.Time) (int, int) {
	local := t.In(lobbies.Location)
	return int(local.Weekday()), local.Hour()*60 + local.Minute()
}

func toTalent(r db.CatalogRow) Talent {
	t := Talent{
		CharacterID:     r.ID.String(),
		Nick:            r.Nick,
		ClassID:         r.ClassID,
		Level:           int(r.Level),
		Role:            r.Role,
		Portrait:        r.Portrait,
		Link:            r.Link.String,
		Days:            daysOf(r.Days),
		Start:           clock(r.StartMinute),
		End:             clock(r.EndMinute),
		AnyInstance:     r.AnyInstance,
		Instances:       []Instance{},
		DiscordUsername: r.Username,
	}
	for _, id := range r.InstanceIds {
		if i, ok := catalog.InstanceByID(id); ok {
			t.Instances = append(t.Instances, Instance{ID: i.ID, Name: i.Name})
		}
	}
	return t
}
