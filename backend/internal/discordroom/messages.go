package discordroom

import (
	"fmt"
	"strings"
	"time"
	_ "time/tzdata" // o fuso de Brasília dentro do binário, como em lobbies

	"github.com/LeiteMurphy/ro-lobby/backend/internal/catalog"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
)

// Textos que o bot posta na sala (spec discord-bot, RN-08, RN-10, RN-11, RN-14, RN-15,
// RN-17, RN-26). Todo texto vindo do usuário passa por escape (D-06).

// brasilia é o fuso dos horários na tela (RN-09 da home-local). É o mesmo de
// lobbies.Location; fica aqui porque o pacote lobbies chama os ganchos da sala.
var brasilia = func() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		panic(err)
	}
	return loc
}()

var weekdays = []string{"domingo", "segunda", "terça", "quarta", "quinta", "sexta", "sábado"}

// when é "sábado, 10/10 às 20:00 (horário de Brasília)" (RN-09 da home-local).
func when(t time.Time) string {
	local := t.In(brasilia)
	return fmt.Sprintf("%s, %02d/%02d às %02d:%02d (horário de Brasília)",
		weekdays[local.Weekday()], local.Day(), int(local.Month()), local.Hour(), local.Minute())
}

// clock é "20:00" em Brasília.
func clock(t time.Time) string {
	local := t.In(brasilia)
	return fmt.Sprintf("%02d:%02d", local.Hour(), local.Minute())
}

// zeroWidthSpace separa o "@" do resto, para um texto do usuário não virar menção.
const zeroWidthSpace = string(rune(0x200b))

// escaper tira o efeito de markdown e de menções de um texto do usuário (D-06). As
// menções já não notificam por causa do allowed_mentions; o escape também evita que
// apareçam formatadas.
var escaper = strings.NewReplacer(
	`\`, `\\`, "*", `\*`, "_", `\_`, "~", `\~`, "`", "\\`", "|", `\|`, ">", `\>`,
	"<", `\<`, "#", `\#`, "@", "@"+zeroWidthSpace,
)

func escape(s string) string { return escaper.Replace(s) }

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

// opening são os dados da mensagem de abertura (RN-08).
type opening struct {
	Title    string
	Server   string
	StartsAt time.Time
	Members  []db.RoomMembersRow
	URL      string
}

var roleLabels = map[string]string{"tank": "Tank", "support": "Suporte", "dps": "Dano"}

// serverName é o nome do servidor do jogo (RN-08 da servidor).
func serverName(id string) string {
	if s, ok := catalog.ServerByID(id); ok {
		return s.Name
	}
	return id
}

// openingMessage é a mensagem de abertura (RN-08): título, servidor e horário; a
// composição por função com nick e classe; a menção a cada conectado; quem está fora; o
// link do lobby.
func openingMessage(o opening) postPayload {
	var b strings.Builder
	fmt.Fprintf(&b, "🎮 **Sala aberta: %s**\n", escape(o.Title))
	fmt.Fprintf(&b, "Servidor %s · %s\n", serverName(o.Server), when(o.StartsAt))
	ids := []string{}
	var outside []string
	for _, role := range []string{"tank", "support", "dps"} {
		var line []string
		for _, m := range o.Members {
			if m.Role != role {
				continue
			}
			who := escape(nickOf(m))
			if class, ok := catalog.ClassByID(m.ClassID.String); ok {
				who += " (" + class.Name + ")"
			}
			if m.Connected {
				who += " " + mention(m.DiscordID)
				ids = append(ids, m.DiscordID)
			} else {
				outside = append(outside, escape(nickOf(m)))
			}
			line = append(line, who)
		}
		if len(line) > 0 {
			fmt.Fprintf(&b, "\n**%s:** %s", roleLabels[role], strings.Join(line, ", "))
		}
	}
	if len(outside) > 0 {
		fmt.Fprintf(&b, "\n\nAinda fora da sala: %s. Para entrar, conecte-se pelo perfil do RO Lobby.", strings.Join(outside, ", "))
	}
	fmt.Fprintf(&b, "\n\nLobby: %s", o.URL)
	return postPayload{Content: b.String(), Mentions: ids}
}

// nickOf é o nick, ou "Personagem excluído" (D-01 da lobbies).
func nickOf(m db.RoomMembersRow) string {
	if m.Nick.Valid {
		return m.Nick.String
	}
	return "Personagem excluído"
}

// joinedMessage é a entrada de um membro (RN-10); só menciona quem está conectado.
func joinedMessage(nick, discordID string, connected bool) postPayload {
	content := fmt.Sprintf("➕ **%s** entrou no grupo.", escape(nick))
	if !connected {
		return postPayload{Content: content + " Para entrar na sala, conecte-se pelo perfil do RO Lobby."}
	}
	return postPayload{Content: content + " " + mention(discordID), Mentions: []string{discordID}}
}

// leftMessage é a saída de um membro (RN-11), sem o motivo de uma remoção.
func leftMessage(nick string) postPayload {
	return postPayload{Content: fmt.Sprintf("➖ **%s** saiu do grupo.", escape(nick))}
}

// connectedMessage é o membro que se conectou depois da sala aberta (RN-13).
func connectedMessage(nick, discordID string) postPayload {
	return postPayload{Content: fmt.Sprintf("🔗 **%s** entrou na sala. %s", escape(nick), mention(discordID)), Mentions: []string{discordID}}
}

// startChangedMessage é o horário novo (RN-15).
func startChangedMessage(startsAt time.Time, ids []string) postPayload {
	content := "🕒 **Horário novo:** " + when(startsAt) + "."
	if len(ids) > 0 {
		content += "\n" + mentions(ids)
	}
	return postPayload{Content: content, Mentions: ids}
}

// cancelledMessage é o cancelamento com o motivo (RN-17).
func cancelledMessage(reason string, ids []string) postPayload {
	content := "❌ **O lobby foi cancelado.** Motivo: " + escape(reason) + "\nA sala será apagada em 1 hora."
	if len(ids) > 0 {
		content += "\n" + mentions(ids)
	}
	return postPayload{Content: content, Mentions: ids}
}

// alertMessage é o "Avisar no Discord" (RN-26): texto fixo, menção aos conectados sem o
// dono e o link do lobby.
func alertMessage(owner string, ids []string, url string) postPayload {
	content := fmt.Sprintf("📣 **%s** está chamando o grupo!", escape(owner))
	if len(ids) > 0 {
		content += " " + mentions(ids)
	}
	content += "\nLobby: " + url
	return postPayload{Content: content, Mentions: ids}
}
