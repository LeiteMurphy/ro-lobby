package discordroom

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
)

// Trabalhador da sala (spec discord-bot, D-02, RN-14, RN-16, RN-17, RN-21, RNF-04). Roda no
// processo da API: a cada volta, executa os jobs vencidos, posta os lembretes e apaga as
// salas vencidas. Tudo fica no banco, então um reinício só retoma de onde parou.

const (
	// Every é o intervalo entre as voltas do trabalhador.
	Every = 5 * time.Second
	// maxBackoff limita a espera entre tentativas de um job (RN-21).
	maxBackoff = 5 * time.Minute
	// reminderBefore, openFor e closingFor são as janelas das RN-14, RN-16 e RN-17.
	reminderBefore = 15 * time.Minute
	openFor        = 6 * time.Hour
	closingFor     = time.Hour
	// jobsPerTick limita uma volta para o lembrete e a remoção não esperarem demais.
	jobsPerTick = 50
)

// Run executa Tick a cada every até o ctx acabar.
func (s *Service) Run(ctx context.Context, every time.Duration) {
	if !s.Enabled() {
		return
	}
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		if err := s.Tick(ctx); err != nil && ctx.Err() == nil {
			s.log.Error("trabalhador da sala do Discord", "err", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// Tick é uma volta do trabalhador: jobs, lembretes e remoções. Um erro do Discord não
// interrompe a volta; ele fica para a próxima (RN-21).
func (s *Service) Tick(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}
	for range jobsPerTick {
		ran, err := s.runNextJob(ctx)
		if err != nil {
			return err
		}
		if !ran {
			break
		}
	}
	if err := s.sendReminders(ctx); err != nil {
		return err
	}
	return s.deleteDueRooms(ctx)
}

// backoff é 2ⁿ segundos, até maxBackoff (RN-21).
func backoff(attempts int32) time.Duration {
	if attempts >= 9 {
		return maxBackoff
	}
	d := time.Duration(1<<attempts) * 2 * time.Second
	return min(d, maxBackoff)
}

// runNextJob executa o próximo job vencido, numa transação que o trava. Devolve false se
// não havia job.
func (s *Service) runNextJob(ctx context.Context) (bool, error) {
	now := s.Now().UTC()
	ran := false
	err := s.inTx(ctx, func(q *db.Queries) error {
		job, err := q.NextDiscordJob(ctx, now)
		if isNoRows(err) {
			return nil
		}
		if err != nil {
			return err
		}
		ran = true
		// RN-21: sala apagada descarta o que estava pendente.
		if job.RoomStatus == StatusDeleted {
			return q.FinishDiscordJob(ctx, db.FinishDiscordJobParams{ID: job.ID, Now: now,
				LastError: pgtype.Text{String: "descartado: sala apagada", Valid: true}})
		}
		if err := s.execute(ctx, q, job); err != nil {
			s.log.Warn("job da sala do Discord falhou; nova tentativa depois", "job", job.ID, "kind", job.Kind, "err", err)
			return q.RetryDiscordJob(ctx, db.RetryDiscordJobParams{
				ID: job.ID, RunAt: now.Add(backoff(job.Attempts)), LastError: pgtype.Text{String: err.Error(), Valid: true},
			})
		}
		return q.FinishDiscordJob(ctx, db.FinishDiscordJobParams{ID: job.ID, Now: now})
	})
	return ran, err
}

func (s *Service) execute(ctx context.Context, q *db.Queries, job db.NextDiscordJobRow) error {
	switch job.Kind {
	case JobPost:
		var msg postPayload
		if err := json.Unmarshal(job.Payload, &msg); err != nil {
			return fmt.Errorf("payload inválido: %w", err)
		}
		return s.bot.Post(ctx, job.TextChannelID, msg.Content, msg.Mentions)
	case JobSyncAccess:
		// D-02: decide na hora de rodar, então rodar de novo ou fora de ordem não estraga.
		m, err := q.RoomMember(ctx, db.RoomMemberParams{LobbyID: job.LobbyID, UserID: job.UserID})
		if isNoRows(err) {
			return nil // Usuário excluído: não há o que liberar
		}
		if err != nil {
			return err
		}
		channels := []string{job.CategoryID, job.TextChannelID, job.VoiceChannelID}
		allow := m.Member && m.Connected && job.RoomStatus == StatusOpen
		for _, ch := range channels {
			if allow {
				err = s.bot.Allow(ctx, ch, m.DiscordID)
			} else {
				err = s.bot.Revoke(ctx, ch, m.DiscordID)
			}
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("tipo de job desconhecido: %q", job.Kind)
	}
}

// sendReminders posta o lembrete de cada sala na janela (RN-14), uma vez só.
func (s *Service) sendReminders(ctx context.Context) error {
	now := s.Now().UTC()
	return s.inTx(ctx, func(q *db.Queries) error {
		due, err := q.DueReminders(ctx, db.DueRemindersParams{Now: now, WindowEnd: now.Add(reminderBefore)})
		if err != nil {
			return err
		}
		for _, r := range due {
			members, err := q.RoomMembers(ctx, r.LobbyID)
			if err != nil {
				return err
			}
			msg := reminderMessage(r.StartsAt, connectedIDs(members))
			if err := s.bot.Post(ctx, r.TextChannelID, msg.Content, msg.Mentions); err != nil {
				s.log.Warn("lembrete da sala falhou; nova tentativa na próxima volta", "lobby", r.LobbyID.String(), "err", err)
				continue
			}
			if err := q.MarkReminderSent(ctx, db.MarkReminderSentParams{LobbyID: r.LobbyID, Now: now}); err != nil {
				return err
			}
		}
		return nil
	})
}

// deleteDueRooms apaga as salas vencidas (RN-16, RN-17): voz, texto e categoria. Se o
// Discord falhar, a sala fica para a próxima volta (RN-21).
func (s *Service) deleteDueRooms(ctx context.Context) error {
	now := s.Now().UTC()
	due, err := s.queries.DueRoomDeletions(ctx, db.DueRoomDeletionsParams{
		StartedBefore: now.Add(-openFor), ClosedBefore: now.Add(-closingFor),
	})
	if err != nil {
		return err
	}
	for _, r := range due {
		var failed error
		for _, ch := range []string{r.VoiceChannelID, r.TextChannelID, r.CategoryID} {
			if err := s.bot.DeleteChannel(ctx, ch); err != nil {
				failed = err
				break
			}
		}
		if failed != nil {
			s.log.Warn("remoção da sala falhou; nova tentativa na próxima volta", "lobby", r.LobbyID.String(), "err", failed)
			continue
		}
		if err := s.queries.MarkRoomDeleted(ctx, db.MarkRoomDeletedParams{LobbyID: r.LobbyID, Now: now}); err != nil {
			return err
		}
	}
	return nil
}

// connectedIDs são os IDs do Discord dos membros conectados, na ordem (RN-07).
func connectedIDs(members []db.RoomMembersRow) []string {
	var ids []string
	for _, m := range members {
		if m.Connected {
			ids = append(ids, m.DiscordID)
		}
	}
	return ids
}
