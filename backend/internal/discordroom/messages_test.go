package discordroom

import (
	"strings"
	"testing"
)

// D-06: nick e título vindos do usuário não viram menção nem formatação.
func TestEscape_D06_UserTextCannotMention(t *testing.T) {
	got := escape("@everyone **Brasa** <@123>")
	if strings.Contains(got, "@everyone") || strings.Contains(got, "<@123>") || strings.Contains(got, "**") {
		t.Errorf("escape = %q", got)
	}
}

// RN-05: o nome da sala tem o título e o código curto do lobby.
func TestRoomName_RN05_TitleAndShortCode(t *testing.T) {
	if got := roomName("Templo do Demônio Rei", "4b1c2d3e-5f60-4a7b-8c9d-0e1f2a3b4c5d"); got != "Templo do Demônio Rei · 4b1c" {
		t.Errorf("roomName = %q", got)
	}
}
