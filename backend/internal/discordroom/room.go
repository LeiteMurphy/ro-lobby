package discordroom

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordbot"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// Abrir a sala e mostrar a sala no detalhe (spec discord-bot, RN-05 a RN-09, RN-18 a RN-20,
// D-05).

// OpenTimeout é o prazo da abertura, somando as chamadas ao Discord (RNF-02).
const OpenTimeout = 10 * time.Second

var (
	// ErrDisabled: a sala no Discord está desligada (RN-22).
	ErrDisabled = errors.New("discordroom: sala no Discord desligada")
	// ErrNotFound vale para lobby inexistente.
	ErrNotFound = errors.New("discordroom: lobby não encontrado")
	// ErrUnavailable: o Discord falhou (RN-20).
	ErrUnavailable = errors.New("discordroom: Discord indisponível")
	// ErrGuildFull: o servidor oficial está no limite de canais (RN-20).
	ErrGuildFull = errors.New("discordroom: servidor do RO Lobby cheio")
)

// Códigos das regras (D-07 da candidatura); a API responde 409 com o código.
const (
	CodeNotOwner   = "not_owner"
	CodeNotOpen    = "not_open"
	CodeRoomExists = "room_exists"
	CodeNoRoom     = "no_room"
)

// RuleError é uma regra violada.
type RuleError struct{ Code string }

func (e *RuleError) Error() string { return "discordroom: regra violada: " + e.Code }

func rule(code string) error { return &RuleError{Code: code} }

// Room é a sala gravada.
type Room struct {
	Status  string
	TextID  string
	GuildID string
}

// URL é o link do canal de texto no Discord (RN-18).
func (r Room) URL() string { return "https://discord.com/channels/" + r.GuildID + "/" + r.TextID }

// Open abre a sala do lobby aberto do dono (RN-05 a RN-08, RN-20). A trava do lobby faz
// dois cliques ao mesmo tempo criarem uma sala só.
func (s *Service) Open(ctx context.Context, ownerID, lobbyID string) (Room, error) {
	if !s.Enabled() {
		return Room{}, ErrDisabled
	}
	lid, err := parseUUID(lobbyID)
	if err != nil {
		return Room{}, ErrNotFound
	}
	ctx, cancel := context.WithTimeout(ctx, OpenTimeout)
	defer cancel()

	now := s.Now().UTC()
	var created discordbot.Room
	err = s.inTx(ctx, func(q *db.Queries) error {
		l, err := q.LockLobbyForRoom(ctx, lid)
		if isNoRows(err) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if l.OwnerID.String() != ownerID {
			return rule(CodeNotOwner)
		}
		if l.CancelledAt.Valid || !l.StartsAt.After(now) {
			return rule(CodeNotOpen)
		}
		if _, err := q.GetDiscordRoom(ctx, lid); err == nil {
			return rule(CodeRoomExists) // RN-09: aberta, encerrando ou apagada
		} else if !isNoRows(err) {
			return err
		}
		members, err := q.RoomMembers(ctx, lid)
		if err != nil {
			return err
		}
		title := lobbyTitle(l.InstanceName)
		created, err = s.bot.CreateRoom(ctx, roomName(title, lobbyID), connectedIDs(members))
		switch {
		case errors.Is(err, discordbot.ErrGuildFull):
			return ErrGuildFull
		case err != nil:
			return fmt.Errorf("%w: %w", ErrUnavailable, err)
		}
		if err := q.InsertDiscordRoom(ctx, db.InsertDiscordRoomParams{
			LobbyID: lid, CategoryID: created.CategoryID, TextChannelID: created.TextID,
			VoiceChannelID: created.VoiceID, Now: now,
		}); err != nil {
			return err
		}
		return enqueuePost(ctx, q, lid, openingMessage(opening{
			Title: title, Server: l.Server, StartsAt: l.StartsAt, Members: members,
			URL: s.lobbyURL(lobbyID),
		}), now)
	})
	if err != nil {
		if created.CategoryID != "" {
			// A gravação falhou depois de criar os canais: apaga para não sobrar órfão.
			s.undo(created)
		}
		var re *RuleError
		if errors.As(err, &re) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrGuildFull) || errors.Is(err, ErrUnavailable) {
			return Room{}, err
		}
		return Room{}, fmt.Errorf("discordroom: abrir sala: %w", err)
	}
	return Room{Status: StatusOpen, TextID: created.TextID, GuildID: s.bot.GuildID()}, nil
}

func (s *Service) undo(r discordbot.Room) {
	ctx, cancel := context.WithTimeout(context.Background(), OpenTimeout)
	defer cancel()
	for _, ch := range []string{r.VoiceID, r.TextID, r.CategoryID} {
		if err := s.bot.DeleteChannel(ctx, ch); err != nil {
			s.log.Error("canal órfão no Discord depois de uma abertura que falhou", "channel_id", ch, "err", err)
		}
	}
}

// lobbyURL é o link público do lobby no web (RN-08).
func (s *Service) lobbyURL(lobbyID string) string {
	return strings.TrimRight(s.Origin, "/") + "/lobbies/" + lobbyID
}

// lobbyTitle é o nome da instância, o título livre ou "Qualquer instância" (RN-05).
func lobbyTitle(name pgtype.Text) string {
	if name.Valid && name.String != "" {
		return name.String
	}
	return lobbies.AnyInstanceName
}

// roomName é o título com o código curto do lobby (RN-05): "Templo do Demônio Rei · a1b2".
func roomName(title, lobbyID string) string {
	code := strings.ReplaceAll(lobbyID, "-", "")
	if len(code) > 4 {
		code = code[:4]
	}
	name := title
	if r := []rune(name); len(r) > 80 {
		name = string(r[:80])
	}
	return name + " · " + code
}

// Member é um membro como a seção "Sala no Discord" mostra (RN-18).
type Member struct {
	UserID    string
	Nick      string
	Connected bool
}

// View é o bloco discord do detalhe do lobby, só para membros (RN-18, RN-19, D-05).
type View struct {
	Enabled bool
	// Room é nil sem sala ou com a feature desligada.
	Room            *RoomView
	Members         []Member
	ViewerConnected bool
}

// RoomView é a sala como o membro vê.
type RoomView struct {
	Status string
	// URL só com a sala aberta (RN-18).
	URL string
}

// View monta o bloco discord para quem já se sabe que é membro (o dono ou um aceito).
func (s *Service) View(ctx context.Context, lobbyID, viewerID string) (View, error) {
	out := View{Enabled: s.Enabled(), Members: []Member{}}
	lid, err := parseUUID(lobbyID)
	if err != nil {
		return out, ErrNotFound
	}
	members, err := s.queries.RoomMembers(ctx, lid)
	if err != nil {
		return out, fmt.Errorf("discordroom: membros: %w", err)
	}
	for _, m := range members {
		out.Members = append(out.Members, Member{UserID: m.UserID.String(), Nick: m.Nick.String, Connected: m.Connected})
		if m.UserID.String() == viewerID {
			out.ViewerConnected = m.Connected
		}
	}
	if !s.Enabled() {
		return out, nil
	}
	room, err := s.queries.GetDiscordRoom(ctx, lid)
	switch {
	case isNoRows(err):
		return out, nil
	case err != nil:
		return out, fmt.Errorf("discordroom: sala: %w", err)
	}
	out.Room = &RoomView{Status: room.Status}
	if room.Status == StatusOpen {
		out.Room.URL = Room{TextID: room.TextChannelID, GuildID: s.bot.GuildID()}.URL()
	}
	return out, nil
}
