// Package lobbies guarda as regras dos lobbies (spec lobbies, RN-04 a RN-20, design D-01
// a D-08).
package lobbies

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	// D-04: a imagem distroless não tem a base de fusos; ela vai embutida no binário.
	_ "time/tzdata"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/catalog"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
)

const (
	// MaxSlots é o total de vagas de um lobby (RN-06, P-01 da candidatura-lobby).
	MaxSlots = 12
	// MaxOpenPerOwner é o limite de lobbies abertos por Usuário (RN-11).
	MaxOpenPerOwner = 5
	// ConflictWindow é a janela de cada lobby para conflito de horário (RN-10, D-03).
	ConflictWindow = 2 * time.Hour
	// Days é quantos dias do seletor da Home aceitam lobby: hoje e os 13 seguintes (RN-05).
	Days = 14
	// MaxLevel e os tamanhos de texto das regras RN-07, RN-09 e RN-19.
	MaxLevel        = 275
	MaxNoteLength   = 250
	MinReasonLength = 10
	MaxReasonLength = 250

	StatusOpen      = "open"
	StatusStarted   = "started"
	StatusCancelled = "cancelled"
)

// Location é o fuso em que o Usuário informa e vê os horários (RN-05, RN-09 da home-local).
var Location = mustLoad("America/Sao_Paulo")

func mustLoad(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

var (
	// ErrNotFound vale para lobby inexistente e, nas escritas, para lobby de outro
	// Usuário (RN-20, D-08).
	ErrNotFound = errors.New("lobbies: lobby não encontrado")
	// ErrNotOpen: o lobby já começou ou foi cancelado (RN-17, RN-19, D-08).
	ErrNotOpen = errors.New("lobbies: lobby não está aberto")
	// ErrLimitReached: o Usuário já tem MaxOpenPerOwner lobbies abertos (RN-11).
	ErrLimitReached = errors.New("lobbies: limite de lobbies abertos")
	// ErrInvalidRange: intervalo de datas da listagem inválido.
	ErrInvalidRange = errors.New("lobbies: intervalo de datas inválido")
)

// Campos e códigos de erro de validação (D-07). O web traduz para pt-BR.
const (
	FieldInstanceID  = "instanceId"
	FieldStartsAt    = "startsAt"
	FieldSlots       = "slots"
	FieldMinLevel    = "minLevel"
	FieldCharacterID = "characterId"
	FieldNote        = "note"
	FieldReason      = "reason"

	CodeRequired      = "required"
	CodeInvalid       = "invalid"
	CodeTooLong       = "too_long"
	CodeTooShort      = "too_short"
	CodeConflict      = "conflict"
	CodeBelowOccupied = "below_occupied"
	CodeAboveOwner    = "above_owner"
	CodeLevelTooLow   = "level_too_low"
)

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
	return "lobbies: dados inválidos (" + strings.Join(parts, ", ") + ")"
}

func fieldError(field, code string) error {
	return &ValidationError{Fields: []FieldError{{Field: field, Code: code}}}
}

// Slots são as vagas (ou os ocupantes) por função.
type Slots struct {
	Tank    int
	Support int
	Dps     int
}

func (s Slots) total() int { return s.Tank + s.Support + s.Dps }

func (s Slots) of(role string) int {
	switch role {
	case "tank":
		return s.Tank
	case "support":
		return s.Support
	default:
		return s.Dps
	}
}

// Input é o que o Usuário informa ao criar.
type Input struct {
	InstanceID  string
	StartsAt    time.Time
	Slots       Slots
	MinLevel    int
	CharacterID string
	Note        string
}

// UpdateInput é o que o dono muda ao editar; o personagem fica (RN-17). InstanceID vazio
// mantém a instância atual.
type UpdateInput struct {
	InstanceID string
	StartsAt   time.Time
	Slots      Slots
	MinLevel   int
	Note       string
}

// Owner é o dono com o personagem dele. CharacterID e os dados do personagem ficam vazios
// se ele foi excluído depois do início (D-01).
type Owner struct {
	UserID      string
	DiscordName string
	CharacterID string
	Nick        string
	ClassID     string
	Level       int
	// Portrait é o retrato do personagem (RN-15), vazio se ele foi excluído.
	Portrait string
	// Link é o link externo do personagem, para o painel do jogador (RN-31).
	Link string
	Role string
}

type Lobby struct {
	ID            string
	InstanceID    string
	InstanceName  string
	InstanceLevel int
	// InstanceReset vem do catálogo; fica vazio se a instância saiu dele (D-02).
	InstanceReset string
	StartsAt      time.Time
	Status        string
	Slots         Slots
	// Occupied são o dono e os membros aceitos por função (D-03 da candidatura).
	Occupied Slots
	// PendingCount é a quantidade de candidaturas pendentes, pública (RN-28).
	PendingCount int
	MinLevel     int
	Note         string
	Owner        Owner
	CancelReason string
	CreatedAt    time.Time
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

// List devolve os lobbies abertos com início entre os dias from e to (YYYY-MM-DD, no fuso
// de São Paulo, to incluído), por horário (RN-13, RN-14, RN-22).
func (s *Service) List(ctx context.Context, from, to string) ([]Lobby, error) {
	fromDay, err1 := time.ParseInLocation(time.DateOnly, from, Location)
	toDay, err2 := time.ParseInLocation(time.DateOnly, to, Location)
	if err1 != nil || err2 != nil || toDay.Before(fromDay) || toDay.Sub(fromDay) > 62*24*time.Hour {
		return nil, ErrInvalidRange
	}
	rows, err := s.queries.ListOpenLobbies(ctx, db.ListOpenLobbiesParams{
		Now:    s.Now().UTC(),
		FromAt: fromDay.UTC(),
		ToAt:   toDay.AddDate(0, 0, 1).UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("lobbies: listar: %w", err)
	}
	out := make([]Lobby, len(rows))
	for i, r := range rows {
		out[i] = s.toLobby(r.Lobby, owner{r.OwnerNick, r.OwnerClassID, r.OwnerLevel, r.OwnerPortrait, r.OwnerLink, r.OwnerUsername, r.OwnerGlobalName},
			counts{r.AcceptedTank, r.AcceptedSupport, r.AcceptedDps, r.PendingCount})
		// RN-32 da candidatura: a lista é pública, então o Discord do anfitrião não vai.
		out[i].Owner.DiscordName = ""
	}
	return out, nil
}

// Get devolve o lobby em qualquer estado (RN-15).
func (s *Service) Get(ctx context.Context, id string) (Lobby, error) {
	lid, err := parseUUID(id)
	if err != nil {
		return Lobby{}, ErrNotFound
	}
	return s.get(ctx, s.queries, lid)
}

func (s *Service) get(ctx context.Context, q *db.Queries, id pgtype.UUID) (Lobby, error) {
	r, err := q.GetLobby(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Lobby{}, ErrNotFound
	}
	if err != nil {
		return Lobby{}, fmt.Errorf("lobbies: buscar: %w", err)
	}
	return s.toLobby(r.Lobby, owner{r.OwnerNick, r.OwnerClassID, r.OwnerLevel, r.OwnerPortrait, r.OwnerLink, r.OwnerUsername, r.OwnerGlobalName},
		counts{r.AcceptedTank, r.AcceptedSupport, r.AcceptedDps, r.PendingCount}), nil
}

// Create valida e cria o lobby com o personagem do dono na vaga dele (RN-04 a RN-11).
func (s *Service) Create(ctx context.Context, userID string, in Input) (Lobby, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return Lobby{}, fmt.Errorf("lobbies: id de usuário inválido: %w", err)
	}
	now := s.Now().UTC()

	var errs []FieldError
	instance, ok := catalog.InstanceByID(in.InstanceID)
	switch {
	case in.InstanceID == "":
		errs = append(errs, FieldError{FieldInstanceID, CodeRequired})
	case !ok:
		errs = append(errs, FieldError{FieldInstanceID, CodeInvalid})
	}
	errs = append(errs, checkStart(in.StartsAt, now)...)
	errs = append(errs, checkSlots(in.Slots)...)
	if ok && (in.MinLevel < instance.Level || in.MinLevel > MaxLevel) {
		errs = append(errs, FieldError{FieldMinLevel, CodeInvalid})
	}
	note, noteErrs := checkNote(in.Note)
	errs = append(errs, noteErrs...)
	cid, cidErr := parseUUID(in.CharacterID)
	if in.CharacterID == "" {
		errs = append(errs, FieldError{FieldCharacterID, CodeRequired})
	} else if cidErr != nil {
		errs = append(errs, FieldError{FieldCharacterID, CodeInvalid})
	}
	if len(errs) > 0 {
		return Lobby{}, &ValidationError{Fields: errs}
	}

	var created Lobby
	err = s.inTx(ctx, uid, func(q *db.Queries) error {
		character, err := q.GetOwnCharacter(ctx, db.GetOwnCharacterParams{ID: cid, UserID: uid})
		if errors.Is(err, pgx.ErrNoRows) {
			return fieldError(FieldCharacterID, CodeInvalid) // RN-08: de outro Usuário ou inexistente
		}
		if err != nil {
			return err
		}
		if int(character.Level) < in.MinLevel {
			return fieldError(FieldCharacterID, CodeLevelTooLow)
		}
		if in.Slots.of(character.Role) < 1 {
			return fieldError(FieldSlots, CodeInvalid) // RN-08: a função do dono precisa de vaga
		}
		open, err := q.CountOpenLobbiesByOwner(ctx, db.CountOpenLobbiesByOwnerParams{OwnerID: uid, Now: now})
		if err != nil {
			return err
		}
		if open >= MaxOpenPerOwner {
			return ErrLimitReached
		}
		if err := checkConflict(ctx, q, cid, pgtype.UUID{}, in.StartsAt); err != nil {
			return err
		}
		id, err := q.CreateLobby(ctx, db.CreateLobbyParams{
			OwnerID:          uid,
			InstanceID:       instance.ID,
			InstanceName:     instance.Name,
			InstanceLevel:    int16(instance.Level), //nolint:gosec // catálogo: 1..275
			StartsAt:         in.StartsAt.UTC(),
			SlotsTank:        int16(in.Slots.Tank),    //nolint:gosec // checkSlots: 0..12
			SlotsSupport:     int16(in.Slots.Support), //nolint:gosec // checkSlots: 0..12
			SlotsDps:         int16(in.Slots.Dps),     //nolint:gosec // checkSlots: 0..12
			MinLevel:         int16(in.MinLevel),      //nolint:gosec // validado: 1..275
			OwnerCharacterID: cid,
			OwnerRole:        character.Role,
			Note:             optionalText(note),
			Now:              now,
		})
		if err != nil {
			return err
		}
		created, err = s.get(ctx, q, id)
		return err
	})
	if err != nil {
		return Lobby{}, wrap(err, "criar")
	}
	return created, nil
}

// Update grava instância, horário, vagas, nível mínimo e observação do lobby aberto do
// dono (RN-17, RN-18, RN-20). Membros e candidaturas continuam ao trocar a instância.
func (s *Service) Update(ctx context.Context, userID, id string, in UpdateInput) (Lobby, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return Lobby{}, fmt.Errorf("lobbies: id de usuário inválido: %w", err)
	}
	lid, err := parseUUID(id)
	if err != nil {
		return Lobby{}, ErrNotFound
	}
	now := s.Now().UTC()

	var errs []FieldError
	instance, inCatalog := catalog.InstanceByID(in.InstanceID)
	errs = append(errs, checkStart(in.StartsAt, now)...)
	errs = append(errs, checkSlots(in.Slots)...)
	if in.MinLevel < 1 || in.MinLevel > MaxLevel {
		errs = append(errs, FieldError{FieldMinLevel, CodeInvalid})
	}
	note, noteErrs := checkNote(in.Note)
	errs = append(errs, noteErrs...)
	if len(errs) > 0 {
		return Lobby{}, &ValidationError{Fields: errs}
	}

	var updated Lobby
	err = s.inTx(ctx, uid, func(q *db.Queries) error {
		current, err := s.ownOpenLobby(ctx, q, uid, lid, now)
		if err != nil {
			return err
		}
		// RN-17: sem instância ou com a mesma, vale a gravada, mesmo que tenha saído do
		// catálogo; outra precisa estar no catálogo.
		instanceID, instanceName, instanceLevel := current.Lobby.InstanceID, current.Lobby.InstanceName, current.Lobby.InstanceLevel
		if in.InstanceID != "" && in.InstanceID != current.Lobby.InstanceID {
			if !inCatalog {
				return fieldError(FieldInstanceID, CodeInvalid)
			}
			instanceID, instanceName, instanceLevel = instance.ID, instance.Name, int16(instance.Level) //nolint:gosec // catálogo: 1..275
		}
		if in.MinLevel < int(instanceLevel) {
			return fieldError(FieldMinLevel, CodeInvalid) // RN-07
		}
		if current.Lobby.OwnerCharacterID.Valid && in.MinLevel > int(current.OwnerLevel.Int16) {
			return fieldError(FieldMinLevel, CodeAboveOwner) // RN-18
		}
		// RN-18: as vagas não ficam abaixo dos ocupantes, o dono e os membros (D-03 da
		// candidatura).
		occupied := s.toLobby(current.Lobby, owner{}, counts{current.AcceptedTank, current.AcceptedSupport, current.AcceptedDps, 0}).Occupied
		for _, role := range []string{"tank", "support", "dps"} {
			if in.Slots.of(role) < occupied.of(role) {
				return fieldError(FieldSlots, CodeBelowOccupied)
			}
		}
		if current.Lobby.OwnerCharacterID.Valid {
			if err := checkConflict(ctx, q, current.Lobby.OwnerCharacterID, lid, in.StartsAt); err != nil {
				return err
			}
		}
		if err := q.UpdateLobby(ctx, db.UpdateLobbyParams{
			ID:            lid,
			InstanceID:    instanceID,
			InstanceName:  instanceName,
			InstanceLevel: instanceLevel,
			StartsAt:      in.StartsAt.UTC(),
			SlotsTank:     int16(in.Slots.Tank),    //nolint:gosec // checkSlots: 0..12
			SlotsSupport:  int16(in.Slots.Support), //nolint:gosec // checkSlots: 0..12
			SlotsDps:      int16(in.Slots.Dps),     //nolint:gosec // checkSlots: 0..12
			MinLevel:      int16(in.MinLevel),      //nolint:gosec // validado: 1..275
			Note:          optionalText(note),
		}); err != nil {
			return err
		}
		updated, err = s.get(ctx, q, lid)
		return err
	})
	if err != nil {
		return Lobby{}, wrap(err, "editar")
	}
	return updated, nil
}

// Cancel cancela o lobby aberto do dono, com motivo de 10 a 250 caracteres (RN-19, RN-20).
func (s *Service) Cancel(ctx context.Context, userID, id, reason string) (Lobby, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return Lobby{}, fmt.Errorf("lobbies: id de usuário inválido: %w", err)
	}
	lid, err := parseUUID(id)
	if err != nil {
		return Lobby{}, ErrNotFound
	}
	reason = strings.TrimSpace(reason)
	switch n := utf8.RuneCountInString(reason); {
	case n < MinReasonLength:
		return Lobby{}, fieldError(FieldReason, CodeTooShort)
	case n > MaxReasonLength:
		return Lobby{}, fieldError(FieldReason, CodeTooLong)
	}
	now := s.Now().UTC()

	var cancelled Lobby
	err = s.inTx(ctx, uid, func(q *db.Queries) error {
		if _, err := s.ownOpenLobby(ctx, q, uid, lid, now); err != nil {
			return err
		}
		if err := q.CancelLobby(ctx, db.CancelLobbyParams{ID: lid, Now: now, Reason: reason}); err != nil {
			return err
		}
		// RN-16 da candidatura: as pendentes expiram, com evento no histórico (D-02).
		expired, err := q.ExpirePendingForLobby(ctx, db.ExpirePendingForLobbyParams{LobbyID: lid, Now: now})
		if err != nil {
			return err
		}
		for _, id := range expired {
			if err := q.InsertApplicationEvent(ctx, db.InsertApplicationEventParams{
				ApplicationID: id,
				FromStatus:    pgtype.Text{String: "pending", Valid: true},
				ToStatus:      "expired",
				ActorID:       uid,
				Now:           now,
			}); err != nil {
				return err
			}
		}
		// RN-16 e D-12 da candidatura: os pedidos de troca pendentes também expiram.
		expiredSwaps, err := q.ExpirePendingSwapsForLobby(ctx, db.ExpirePendingSwapsForLobbyParams{LobbyID: lid, Now: now})
		if err != nil {
			return err
		}
		for _, id := range expiredSwaps {
			if err := q.InsertSwapRequestEvent(ctx, db.InsertSwapRequestEventParams{
				SwapRequestID: id,
				FromStatus:    pgtype.Text{String: "pending", Valid: true},
				ToStatus:      "expired",
				ActorID:       uid,
				Now:           now,
			}); err != nil {
				return err
			}
		}
		cancelled, err = s.get(ctx, q, lid)
		return err
	})
	if err != nil {
		return Lobby{}, wrap(err, "cancelar")
	}
	return cancelled, nil
}

// ownOpenLobby trava o lobby do dono e confere que ele está aberto (RN-17, RN-20, D-08).
func (s *Service) ownOpenLobby(ctx context.Context, q *db.Queries, uid, lid pgtype.UUID, now time.Time) (db.GetLobbyRow, error) {
	if _, err := q.GetOwnLobbyForUpdate(ctx, db.GetOwnLobbyForUpdateParams{ID: lid, OwnerID: uid}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetLobbyRow{}, ErrNotFound
		}
		return db.GetLobbyRow{}, err
	}
	current, err := q.GetLobby(ctx, lid)
	if err != nil {
		return db.GetLobbyRow{}, err
	}
	if current.Lobby.CancelledAt.Valid || !current.Lobby.StartsAt.After(now) {
		return db.GetLobbyRow{}, ErrNotOpen
	}
	return current, nil
}

// checkStart confere que o início está no futuro e cai num dos 14 dias do seletor, no fuso
// de São Paulo (RN-05, D-04).
func checkStart(startsAt, now time.Time) []FieldError {
	if startsAt.IsZero() {
		return []FieldError{{FieldStartsAt, CodeRequired}}
	}
	today := dayOf(now)
	lastDay := today.AddDate(0, 0, Days)
	if !startsAt.After(now) || !startsAt.Before(lastDay) {
		return []FieldError{{FieldStartsAt, CodeInvalid}}
	}
	return nil
}

// dayOf é a meia-noite, em São Paulo, do dia do instante t.
func dayOf(t time.Time) time.Time {
	y, m, d := t.In(Location).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, Location)
}

// checkSlots confere cada função de 0 a 12 e o total de 1 a 12 (RN-06).
func checkSlots(s Slots) []FieldError {
	for _, n := range []int{s.Tank, s.Support, s.Dps} {
		if n < 0 || n > MaxSlots {
			return []FieldError{{FieldSlots, CodeInvalid}}
		}
	}
	if t := s.total(); t < 1 || t > MaxSlots {
		return []FieldError{{FieldSlots, CodeInvalid}}
	}
	return nil
}

func checkNote(note string) (string, []FieldError) {
	note = strings.TrimSpace(note)
	if utf8.RuneCountInString(note) > MaxNoteLength {
		return note, []FieldError{{FieldNote, CodeTooLong}}
	}
	return note, nil
}

// checkConflict procura outro lobby não cancelado do personagem a menos de 2 h (RN-10).
func checkConflict(ctx context.Context, q *db.Queries, characterID, exclude pgtype.UUID, startsAt time.Time) error {
	conflict, err := q.HasScheduleConflict(ctx, db.HasScheduleConflictParams{
		CharacterID: characterID,
		ExcludeID:   exclude,
		WindowStart: startsAt.Add(-ConflictWindow).UTC(),
		WindowEnd:   startsAt.Add(ConflictWindow).UTC(),
	})
	if err != nil {
		return err
	}
	if conflict {
		return fieldError(FieldStartsAt, CodeConflict)
	}
	return nil
}

// inTx roda fn numa transação com o Usuário travado (D-05).
func (s *Service) inTx(ctx context.Context, uid pgtype.UUID, fn func(q *db.Queries) error) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)
		if _, err := q.LockUser(ctx, uid); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		return fn(q)
	})
}

func wrap(err error, action string) error {
	var ve *ValidationError
	switch {
	case errors.As(err, &ve), errors.Is(err, ErrNotFound), errors.Is(err, ErrNotOpen), errors.Is(err, ErrLimitReached):
		return err
	default:
		return fmt.Errorf("lobbies: %s: %w", action, err)
	}
}

// counts são as contagens de candidaturas que GetLobby e ListOpenLobbies trazem (D-03).
type counts struct {
	tank, support, dps, pending int64
}

// owner são as colunas do dono que GetLobby e ListOpenLobbies trazem junto do lobby.
type owner struct {
	nick, classID pgtype.Text
	level         pgtype.Int2
	portrait      pgtype.Text
	link          pgtype.Text
	username      string
	globalName    pgtype.Text
}

func (s *Service) toLobby(l db.Lobby, o owner, c counts) Lobby {
	now := s.Now()
	status := StatusOpen
	switch {
	case l.CancelledAt.Valid:
		status = StatusCancelled
	case !l.StartsAt.After(now):
		status = StatusStarted
	}
	out := Lobby{
		ID:            l.ID.String(),
		InstanceID:    l.InstanceID,
		InstanceName:  l.InstanceName,
		InstanceLevel: int(l.InstanceLevel),
		StartsAt:      l.StartsAt.UTC(),
		Status:        status,
		Slots:         Slots{Tank: int(l.SlotsTank), Support: int(l.SlotsSupport), Dps: int(l.SlotsDps)},
		MinLevel:      int(l.MinLevel),
		Note:          l.Note.String,
		CancelReason:  l.CancelReason.String,
		CreatedAt:     l.CreatedAt.UTC(),
		Owner: Owner{
			UserID:      l.OwnerID.String(),
			DiscordName: o.username,
			Role:        l.OwnerRole,
		},
	}
	if o.globalName.Valid && o.globalName.String != "" {
		out.Owner.DiscordName = o.globalName.String
	}
	if i, ok := catalog.InstanceByID(l.InstanceID); ok {
		out.InstanceReset = string(i.Reset)
	}
	// D-01: o dono ocupa a vaga da função dele (P-02); os membros aceitos somam (D-03 da
	// candidatura).
	out.Occupied = Slots{Tank: int(c.tank), Support: int(c.support), Dps: int(c.dps)}
	if status == StatusOpen {
		out.PendingCount = int(c.pending) // D-02: no início, as pendentes contam como expiradas
	}
	switch l.OwnerRole {
	case "tank":
		out.Occupied.Tank++
	case "support":
		out.Occupied.Support++
	default:
		out.Occupied.Dps++
	}
	if l.OwnerCharacterID.Valid {
		out.Owner.CharacterID = l.OwnerCharacterID.String()
		out.Owner.Nick = o.nick.String
		out.Owner.ClassID = o.classID.String
		out.Owner.Level = int(o.level.Int16)
		out.Owner.Portrait = o.portrait.String
		out.Owner.Link = o.link.String
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
