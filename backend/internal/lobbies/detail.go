package lobbies

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
)

// Participant é um membro aceito ou um candidato pendente, como quem olha pode ver
// (RN-28, RN-31, RN-32, D-06 da candidatura).
type Participant struct {
	ApplicationID string
	UserID        string
	// DiscordName fica vazio quando quem olha não é o dono nem membro (RN-32).
	DiscordName string
	// Os dados do personagem ficam vazios se ele foi excluído depois do lobby (RN-26).
	CharacterID string
	Nick        string
	ClassID     string
	Level       int
	Portrait    string
	Link        string
	Role        string
	// Message é a mensagem da candidatura, só para o dono (RN-28).
	Message   string
	CreatedAt time.Time
}

// ViewerApplication é a candidatura mais recente de quem olha, com a justificativa
// (RN-29), o bloqueio (RN-15) e o pedido de troca mais recente (D-12).
type ViewerApplication struct {
	ID          string
	CharacterID string
	Role        string
	Message     string
	Status      string
	Reason      string
	Blocked     bool
	CreatedAt   time.Time
	DecidedAt   time.Time
	// SwapRequest é nil quando a candidatura nunca teve pedido de troca.
	SwapRequest *SwapRequest
}

// SwapRequest é um pedido de troca de personagem do membro (D-08 da candidatura).
type SwapRequest struct {
	ID              string
	ApplicationID   string
	FromCharacterID string
	ToCharacterID   string
	ToRole          string
	Reason          string
	Status          string
	DecisionReason  string
	CreatedAt       time.Time
	// DecidedAt fica zerado enquanto o pedido está pendente.
	DecidedAt time.Time
}

// SwapCharacter é o personagem atual ou o novo de um pedido de troca; os dados ficam
// vazios se ele foi excluído.
type SwapCharacter struct {
	CharacterID string
	Nick        string
	ClassID     string
	Level       int
	Portrait    string
	Role        string
}

// LobbySwapRequest é um pedido de troca pendente como o dono vê (D-12 da candidatura).
type LobbySwapRequest struct {
	ID            string
	ApplicationID string
	UserID        string
	DiscordName   string
	From          SwapCharacter
	To            SwapCharacter
	Reason        string
	CreatedAt     time.Time
}

// Detail é o lobby como quem olha pode ver (D-06 da candidatura).
type Detail struct {
	Lobby
	Members []Participant
	// Pending só vem para o dono (RN-28); os outros veem PendingCount.
	Pending []Participant
	// SwapRequests são os pedidos de troca pendentes, só para o dono (D-12).
	SwapRequests []LobbySwapRequest
	// MyApplication é nil para o dono, para visitantes e para quem não se candidatou.
	MyApplication *ViewerApplication
}

// ApplicationStatus é o estado lido de uma candidatura: pendente de lobby iniciado conta
// como expirada (D-02 e RN-16 da candidatura).
func ApplicationStatus(status string, startsAt, now time.Time) string {
	if status == "pending" && !startsAt.After(now) {
		return "expired"
	}
	return status
}

// Detail devolve o lobby com membros, pendentes e a candidatura de quem olha. viewerID
// vazio é um visitante sem sessão.
func (s *Service) Detail(ctx context.Context, id, viewerID string) (Detail, error) {
	lid, err := parseUUID(id)
	if err != nil {
		return Detail{}, ErrNotFound
	}
	l, err := s.get(ctx, s.queries, lid)
	if err != nil {
		return Detail{}, err
	}
	rows, err := s.queries.ListLobbyApplications(ctx, lid)
	if err != nil {
		return Detail{}, fmt.Errorf("lobbies: candidaturas: %w", err)
	}
	isOwner := viewerID != "" && viewerID == l.Owner.UserID
	isMember := false
	for _, r := range rows {
		if r.Application.Status == "accepted" && r.Application.UserID.String() == viewerID {
			isMember = true
		}
	}

	out := Detail{Lobby: l, Members: []Participant{}}
	// RN-32: o Discord do anfitrião, como o dos membros, só para o dono e os membros.
	if !isOwner && !isMember {
		out.Owner.DiscordName = ""
	}
	if isOwner {
		out.Pending = []Participant{}
		out.SwapRequests = []LobbySwapRequest{}
		if l.Status == StatusOpen {
			swaps, err := s.queries.ListLobbySwapRequests(ctx, lid)
			if err != nil {
				return Detail{}, fmt.Errorf("lobbies: pedidos de troca: %w", err)
			}
			for _, r := range swaps {
				out.SwapRequests = append(out.SwapRequests, toLobbySwapRequest(r))
			}
		}
	}
	for _, r := range rows {
		p := toParticipant(r)
		if !isOwner {
			p.Message = ""
		}
		switch {
		case r.Application.Status == "accepted":
			if !isOwner && !isMember {
				p.DiscordName = ""
			}
			out.Members = append(out.Members, p)
		case isOwner && l.Status == StatusOpen:
			out.Pending = append(out.Pending, p)
		}
	}

	if viewerID != "" && !isOwner {
		uid, err := parseUUID(viewerID)
		if err != nil {
			return Detail{}, fmt.Errorf("lobbies: id de usuário inválido: %w", err)
		}
		mine, err := s.queries.GetUserApplicationInLobby(ctx, db.GetUserApplicationInLobbyParams{LobbyID: lid, UserID: uid})
		switch {
		case errors.Is(err, pgx.ErrNoRows):
		case err != nil:
			return Detail{}, fmt.Errorf("lobbies: minha candidatura: %w", err)
		default:
			out.MyApplication = toViewerApplication(mine, l.StartsAt, s.Now())
			swap, err := s.queries.GetLatestSwapRequest(ctx, mine.ID)
			switch {
			case errors.Is(err, pgx.ErrNoRows):
			case err != nil:
				return Detail{}, fmt.Errorf("lobbies: meu pedido de troca: %w", err)
			default:
				r := NewSwapRequest(swap)
				r.Status = ApplicationStatus(r.Status, l.StartsAt, s.Now()) // D-12
				out.MyApplication.SwapRequest = &r
			}
		}
	}
	return out, nil
}

// NewSwapRequest converte o pedido de troca do banco, com o estado gravado.
func NewSwapRequest(r db.SwapRequest) SwapRequest {
	out := SwapRequest{
		ID:             r.ID.String(),
		ApplicationID:  r.ApplicationID.String(),
		ToRole:         r.ToRole,
		Reason:         r.Reason,
		Status:         r.Status,
		DecisionReason: r.DecisionReason.String,
		CreatedAt:      r.CreatedAt.UTC(),
	}
	if r.FromCharacterID.Valid {
		out.FromCharacterID = r.FromCharacterID.String()
	}
	if r.ToCharacterID.Valid {
		out.ToCharacterID = r.ToCharacterID.String()
	}
	if r.DecidedAt.Valid {
		out.DecidedAt = r.DecidedAt.Time.UTC()
	}
	return out
}

func toLobbySwapRequest(r db.ListLobbySwapRequestsRow) LobbySwapRequest {
	out := LobbySwapRequest{
		ID:            r.SwapRequest.ID.String(),
		ApplicationID: r.SwapRequest.ApplicationID.String(),
		UserID:        r.UserID.String(),
		DiscordName:   r.Username,
		From: SwapCharacter{
			Nick: r.FromNick.String, ClassID: r.FromClassID.String, Level: int(r.FromLevel.Int16),
			Portrait: r.FromPortrait.String, Role: r.FromRole,
		},
		To: SwapCharacter{
			Nick: r.ToNick.String, ClassID: r.ToClassID.String, Level: int(r.ToLevel.Int16),
			Portrait: r.ToPortrait.String, Role: r.SwapRequest.ToRole,
		},
		Reason:    r.SwapRequest.Reason,
		CreatedAt: r.SwapRequest.CreatedAt.UTC(),
	}
	if r.GlobalName.Valid && r.GlobalName.String != "" {
		out.DiscordName = r.GlobalName.String
	}
	if r.SwapRequest.FromCharacterID.Valid {
		out.From.CharacterID = r.SwapRequest.FromCharacterID.String()
	}
	if r.SwapRequest.ToCharacterID.Valid {
		out.To.CharacterID = r.SwapRequest.ToCharacterID.String()
	}
	return out
}

func toParticipant(r db.ListLobbyApplicationsRow) Participant {
	p := Participant{
		ApplicationID: r.Application.ID.String(),
		UserID:        r.Application.UserID.String(),
		DiscordName:   r.Username,
		Nick:          r.Nick.String,
		ClassID:       r.ClassID.String,
		Level:         int(r.Level.Int16),
		Portrait:      r.Portrait.String,
		Link:          r.Link.String,
		Role:          r.Application.Role,
		Message:       r.Application.Message.String,
		CreatedAt:     r.Application.CreatedAt.UTC(),
	}
	if r.GlobalName.Valid && r.GlobalName.String != "" {
		p.DiscordName = r.GlobalName.String
	}
	if r.Application.CharacterID.Valid {
		p.CharacterID = r.Application.CharacterID.String()
	}
	return p
}

func toViewerApplication(a db.Application, startsAt, now time.Time) *ViewerApplication {
	v := &ViewerApplication{
		ID:        a.ID.String(),
		Role:      a.Role,
		Message:   a.Message.String,
		Status:    ApplicationStatus(a.Status, startsAt, now),
		Reason:    a.Reason.String,
		Blocked:   a.Blocked,
		CreatedAt: a.CreatedAt.UTC(),
	}
	if a.CharacterID.Valid {
		v.CharacterID = a.CharacterID.String()
	}
	if a.DecidedAt.Valid {
		v.DecidedAt = a.DecidedAt.Time.UTC()
	}
	return v
}
