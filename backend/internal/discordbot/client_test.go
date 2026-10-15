package discordbot

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordfake"
)

// Cliente do bot contra o Discord falso (spec discord-bot, T-02, RNF-03).

const (
	guild = "1560117933705797643"
	botID = "111111111111111111"
	token = "token-do-bot"
)

func setup(t *testing.T) (*Client, *discordfake.Fake, *bytes.Buffer) {
	t.Helper()
	fake := discordfake.New("id", "segredo", discordfake.User{ID: "1", Username: "u"})
	fake.BotToken = token
	srv := httptest.NewServer(fake.Handler())
	t.Cleanup(srv.Close)
	var logs bytes.Buffer
	c := New(srv.URL+"/api", token, guild, botID, WithTimeout(2*time.Second),
		WithLogger(slog.New(slog.NewTextHandler(&logs, nil))))
	return c, fake, &logs
}

const view = 1 << 10

func hasView(o discordfake.Overwrite, field string) bool {
	v := o.Allow
	if field == "deny" {
		v = o.Deny
	}
	n, _ := strconv.ParseInt(v, 10, 64)
	return n&view != 0
}

// RN-05 / RN-07 / D-04: categoria, texto e voz; @everyone não vê; o bot e os membros veem.
func TestCreateRoom_RN05_RN07_PrivateCategoryWithTextAndVoice(t *testing.T) {
	c, fake, _ := setup(t)
	room, err := c.CreateRoom(t.Context(), "Templo do Demônio Rei · a1b2", []string{"201", "202"})
	if err != nil {
		t.Fatal(err)
	}
	cat, ok1 := fake.Channel(room.CategoryID)
	txt, ok2 := fake.Channel(room.TextID)
	voice, ok3 := fake.Channel(room.VoiceID)
	if !ok1 || !ok2 || !ok3 {
		t.Fatalf("canais = %+v", room)
	}
	if cat.Type != 4 || txt.Type != 0 || voice.Type != 2 || txt.ParentID != cat.ID || voice.ParentID != cat.ID {
		t.Errorf("tipos e categoria: %+v %+v %+v", cat, txt, voice)
	}
	if txt.Name != "templo-do-demônio-rei-a1b2" {
		t.Errorf("nome do texto = %q", txt.Name)
	}
	for _, ch := range []discordfake.Channel{cat, txt, voice} {
		if !hasView(ch.Overwrites[guild], "deny") {
			t.Errorf("%s: @everyone vê", ch.Name)
		}
		for _, id := range []string{botID, "201", "202"} {
			if !hasView(ch.Overwrites[id], "allow") {
				t.Errorf("%s: %s não vê", ch.Name, id)
			}
		}
	}
}

// CA-02.7 / RN-20: uma falha no meio apaga o que já foi criado.
func TestCreateRoom_CA02_7_UndoOnFailure(t *testing.T) {
	c, fake, _ := setup(t)
	fake.SetBotFailures(discordfake.BotFailures{FailCreateAfter: 2})
	if _, err := c.CreateRoom(t.Context(), "Templo", nil); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, quer ErrUnavailable", err)
	}
	if n := fake.ChannelCount(); n != 0 {
		t.Errorf("%d canais sobraram", n)
	}
}

// CA-02.8 / RN-20: o limite de canais vira ErrGuildFull.
func TestCreateRoom_CA02_8_GuildFull(t *testing.T) {
	c, fake, _ := setup(t)
	fake.SetBotFailures(discordfake.BotFailures{GuildFull: true})
	if _, err := c.CreateRoom(t.Context(), "Templo", nil); !errors.Is(err, ErrGuildFull) {
		t.Fatalf("err = %v, quer ErrGuildFull", err)
	}
}

// RNF-02: o 429 é respeitado uma vez; seguido, vira ErrUnavailable.
func TestDo_RNF02_RateLimit(t *testing.T) {
	c, fake, _ := setup(t)
	room, err := c.CreateRoom(t.Context(), "Templo", nil)
	if err != nil {
		t.Fatal(err)
	}
	fake.SetBotFailures(discordfake.BotFailures{RateLimitNext: 1, RetryAfter: 50 * time.Millisecond})
	if err := c.Post(t.Context(), room.TextID, "oi", nil); err != nil {
		t.Errorf("depois de um 429: %v", err)
	}
	fake.SetBotFailures(discordfake.BotFailures{RateLimitNext: 5, RetryAfter: 50 * time.Millisecond})
	if err := c.Post(t.Context(), room.TextID, "oi", nil); !errors.Is(err, ErrUnavailable) {
		t.Errorf("com 429 seguido: err = %v", err)
	}
	fake.SetBotFailures(discordfake.BotFailures{RateLimitNext: 1, RetryAfter: time.Minute})
	if err := c.Post(t.Context(), room.TextID, "oi", nil); !errors.Is(err, ErrUnavailable) {
		t.Errorf("com espera maior que o prazo: err = %v", err)
	}
}

// RN-21: Discord fora do ar vira ErrUnavailable.
func TestDo_RN21_Down(t *testing.T) {
	c, fake, _ := setup(t)
	fake.SetBotFailures(discordfake.BotFailures{Down: true})
	if err := c.Post(t.Context(), "1", "oi", nil); !errors.Is(err, ErrUnavailable) {
		t.Errorf("err = %v", err)
	}
}

// D-06: a mensagem só menciona os IDs listados; @everyone no texto não marca ninguém.
func TestPost_D06_OnlyExplicitMentions(t *testing.T) {
	c, fake, _ := setup(t)
	room, _ := c.CreateRoom(t.Context(), "Templo", nil)
	if err := c.Post(t.Context(), room.TextID, "@everyone <@201>", []string{"201"}); err != nil {
		t.Fatal(err)
	}
	if err := c.Post(t.Context(), room.TextID, "sem menção", nil); err != nil {
		t.Fatal(err)
	}
	msgs := fake.Messages(room.TextID)
	if len(msgs) != 2 {
		t.Fatalf("mensagens = %+v", msgs)
	}
	if m := msgs[0].AllowedMentions; len(m.Parse) != 0 || len(m.Users) != 1 || m.Users[0] != "201" {
		t.Errorf("allowed_mentions = %+v", m)
	}
	if m := msgs[1].AllowedMentions; m.Parse == nil || m.Users == nil || len(m.Users) != 0 {
		t.Errorf("sem menções, allowed_mentions = %+v", m)
	}
}

// RN-10 / RN-11: liberar e revogar o acesso de um membro.
func TestAllowRevoke_RN10_RN11(t *testing.T) {
	c, fake, _ := setup(t)
	room, _ := c.CreateRoom(t.Context(), "Templo", nil)
	if err := c.Allow(t.Context(), room.TextID, "301"); err != nil {
		t.Fatal(err)
	}
	if ch, _ := fake.Channel(room.TextID); !hasView(ch.Overwrites["301"], "allow") {
		t.Error("não liberou")
	}
	if err := c.Revoke(t.Context(), room.TextID, "301"); err != nil {
		t.Fatal(err)
	}
	if ch, _ := fake.Channel(room.TextID); ch.Overwrites["301"] != (discordfake.Overwrite{}) {
		t.Error("não revogou")
	}
}

// CA-05.5: apagar, liberar ou revogar num canal que alguém apagou à mão é sucesso.
func TestDeleteChannel_CA05_5_MissingIsSuccess(t *testing.T) {
	c, fake, _ := setup(t)
	room, _ := c.CreateRoom(t.Context(), "Templo", nil)
	fake.DeleteChannelDirect(room.TextID)
	for name, err := range map[string]error{
		"apagar":  c.DeleteChannel(t.Context(), room.TextID),
		"liberar": c.Allow(t.Context(), room.TextID, "1"),
		"revogar": c.Revoke(t.Context(), room.TextID, "1"),
	} {
		if err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// RNF-01: token do bot recusado vira ErrUnavailable, e o log não traz o token.
func TestDo_RNF01_BadTokenNotLogged(t *testing.T) {
	_, fake, logs := setup(t)
	srv := httptest.NewServer(fake.Handler())
	defer srv.Close()
	bad := New(srv.URL+"/api", "token-errado-secreto", guild, botID, WithLogger(slog.New(slog.NewTextHandler(logs, nil))))
	if err := bad.Post(t.Context(), "1", "oi", nil); !errors.Is(err, ErrUnavailable) {
		t.Errorf("err = %v", err)
	}
	if bytes.Contains(logs.Bytes(), []byte("token-errado-secreto")) {
		t.Error("o log traz o token")
	}
	_, _ = io.Copy(io.Discard, logs)
}
