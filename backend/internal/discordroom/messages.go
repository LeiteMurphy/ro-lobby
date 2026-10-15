package discordroom

import (
	"fmt"
	"strings"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// Textos que o bot posta na sala (spec discord-bot, RN-08, RN-10, RN-11, RN-14, RN-15,
// RN-17, RN-26).

// clock é "20:00" em Brasília.
func clock(t time.Time) string {
	local := t.In(lobbies.Location)
	return fmt.Sprintf("%02d:%02d", local.Hour(), local.Minute())
}

func mention(discordID string) string { return "<@" + discordID + ">" }

func mentions(ids []string) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = mention(id)
	}
	return strings.Join(parts, " ")
}

// reminderMessage é o lembrete 15 minutos antes (RN-14).
func reminderMessage(startsAt time.Time, ids []string) postPayload {
	content := fmt.Sprintf("⏰ **Lembrete:** o grupo começa em 15 minutos, às %s (horário de Brasília).", clock(startsAt))
	if len(ids) > 0 {
		content += "\n" + mentions(ids)
	}
	return postPayload{Content: content, Mentions: ids}
}
