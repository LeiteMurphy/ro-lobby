package lobbies

import (
	"reflect"
	"strings"
	"testing"
)

// CA-07.1 / CA-07.2 / RN-23 (discord-bot): os três formatos aceitos e os recusados.
func TestCheckDiscordLink_CA07_1_CA07_2(t *testing.T) {
	for _, ok := range []string{
		"https://discord.gg/abc123",
		"https://discord.com/invite/Ab-12",
		"https://discord.com/channels/1560117933705797643/1560117933705797644",
		"  https://discord.gg/abc123/  ",
		"",
	} {
		if _, errs := CheckDiscordLink(ok); errs != nil {
			t.Errorf("%q recusado: %v", ok, errs)
		}
	}
	for link, want := range map[string]string{
		"https://exemplo.com/discord":                    CodeInvalid,
		"http://discord.gg/abc123":                       CodeInvalid,
		"https://discord.gg/":                            CodeInvalid,
		"https://discord.com/channels/abc/1":             CodeInvalid,
		"javascript:alert(1)":                            CodeInvalid,
		"https://discord.gg/" + strings.Repeat("a", 190): CodeTooLong,
	} {
		_, errs := CheckDiscordLink(link)
		if !reflect.DeepEqual(errs, []FieldError{{FieldDiscordLink, want}}) {
			t.Errorf("%q: %v, quer %s", link, errs, want)
		}
	}
}
