// Package catalog guarda as listas fixas que a API valida e serve: as classes do bRO e os
// retratos de personagem (spec personagens, RN-06, RN-10, RN-20, design D-03 e D-04).
package catalog

// Tier é o nível da classe na árvore do jogo, como no bROWiki.
type Tier string

const (
	TierNovice       Tier = "aprendiz"
	TierFirst        Tier = "primeira"
	TierSecond       Tier = "segunda"
	TierTranscendent Tier = "transcendental"
	TierThird        Tier = "terceira"
	TierFourth       Tier = "quarta"
	TierExpanded     Tier = "expandida"
)

// Class é uma classe do catálogo. O ID é o nome sem acento em kebab-case e é o que o
// personagem guarda (D-03). Family é a classe de 1ª (ou a base, nas expandidas) de onde a
// linha sai.
type Class struct {
	ID     string
	Name   string
	Plural string
	Tier   Tier
	Family string
}

// classes segue https://browiki.org/wiki/Classes (consultado em 2026-10-01), na ordem da
// página. Name é o singular usado na tela e Plural é o título da página no bROWiki (RN-22
// da spec home-local).
var classes = []Class{
	{ID: "aprendiz", Name: "Aprendiz", Plural: "Aprendizes", Tier: TierNovice, Family: "Aprendiz"},

	// Espadachim
	{ID: "espadachim", Name: "Espadachim", Plural: "Espadachins", Tier: TierFirst, Family: "Espadachim"},
	{ID: "cavaleiro", Name: "Cavaleiro", Plural: "Cavaleiros", Tier: TierSecond, Family: "Espadachim"},
	{ID: "lorde", Name: "Lorde", Plural: "Lordes", Tier: TierTranscendent, Family: "Espadachim"},
	{ID: "cavaleiro-runico", Name: "Cavaleiro Rúnico", Plural: "Cavaleiros Rúnicos", Tier: TierThird, Family: "Espadachim"},
	{ID: "cavaleiro-draconiano", Name: "Cavaleiro Draconiano", Plural: "Cavaleiros Draconianos", Tier: TierFourth, Family: "Espadachim"},
	{ID: "templario", Name: "Templário", Plural: "Templários", Tier: TierSecond, Family: "Espadachim"},
	{ID: "paladino", Name: "Paladino", Plural: "Paladinos", Tier: TierTranscendent, Family: "Espadachim"},
	{ID: "guardiao-real", Name: "Guardião Real", Plural: "Guardiões Reais", Tier: TierThird, Family: "Espadachim"},
	{ID: "guardiao-imperial", Name: "Guardião Imperial", Plural: "Guardiões Imperiais", Tier: TierFourth, Family: "Espadachim"},

	// Mago
	{ID: "mago", Name: "Mago", Plural: "Magos", Tier: TierFirst, Family: "Mago"},
	{ID: "bruxo", Name: "Bruxo", Plural: "Bruxos", Tier: TierSecond, Family: "Mago"},
	{ID: "arquimago", Name: "Arquimago", Plural: "Arquimagos", Tier: TierTranscendent, Family: "Mago"},
	{ID: "arcano", Name: "Arcano", Plural: "Arcanos", Tier: TierThird, Family: "Mago"},
	{ID: "magus", Name: "Magus", Plural: "Magus", Tier: TierFourth, Family: "Mago"},
	{ID: "sabio", Name: "Sábio", Plural: "Sábios", Tier: TierSecond, Family: "Mago"},
	{ID: "professor", Name: "Professor", Plural: "Professores", Tier: TierTranscendent, Family: "Mago"},
	{ID: "feiticeiro", Name: "Feiticeiro", Plural: "Feiticeiros", Tier: TierThird, Family: "Mago"},
	{ID: "elementalista", Name: "Elementalista", Plural: "Elementalistas", Tier: TierFourth, Family: "Mago"},

	// Gatuno
	{ID: "gatuno", Name: "Gatuno", Plural: "Gatunos", Tier: TierFirst, Family: "Gatuno"},
	{ID: "mercenario", Name: "Mercenário", Plural: "Mercenários", Tier: TierSecond, Family: "Gatuno"},
	{ID: "algoz", Name: "Algoz", Plural: "Algozes", Tier: TierTranscendent, Family: "Gatuno"},
	{ID: "sicario", Name: "Sicário", Plural: "Sicários", Tier: TierThird, Family: "Gatuno"},
	{ID: "executor", Name: "Executor", Plural: "Executores", Tier: TierFourth, Family: "Gatuno"},
	{ID: "arruaceiro", Name: "Arruaceiro", Plural: "Arruaceiros", Tier: TierSecond, Family: "Gatuno"},
	{ID: "desordeiro", Name: "Desordeiro", Plural: "Desordeiros", Tier: TierTranscendent, Family: "Gatuno"},
	{ID: "renegado", Name: "Renegado", Plural: "Renegados", Tier: TierThird, Family: "Gatuno"},
	{ID: "mandraque", Name: "Mandraque", Plural: "Mandraques", Tier: TierFourth, Family: "Gatuno"},

	// Mercador
	{ID: "mercador", Name: "Mercador", Plural: "Mercadores", Tier: TierFirst, Family: "Mercador"},
	{ID: "ferreiro", Name: "Ferreiro", Plural: "Ferreiros", Tier: TierSecond, Family: "Mercador"},
	{ID: "mestre-ferreiro", Name: "Mestre-Ferreiro", Plural: "Mestres-Ferreiros", Tier: TierTranscendent, Family: "Mercador"},
	{ID: "mecanico", Name: "Mecânico", Plural: "Mecânicos", Tier: TierThird, Family: "Mercador"},
	{ID: "engenheiro", Name: "Engenheiro", Plural: "Engenheiros", Tier: TierFourth, Family: "Mercador"},
	{ID: "alquimista", Name: "Alquimista", Plural: "Alquimistas", Tier: TierSecond, Family: "Mercador"},
	{ID: "criador", Name: "Criador", Plural: "Criadores", Tier: TierTranscendent, Family: "Mercador"},
	{ID: "bioquimico", Name: "Bioquímico", Plural: "Bioquímicos", Tier: TierThird, Family: "Mercador"},
	{ID: "cientista", Name: "Cientista", Plural: "Cientistas", Tier: TierFourth, Family: "Mercador"},

	// Noviço
	{ID: "novico", Name: "Noviço", Plural: "Noviços", Tier: TierFirst, Family: "Noviço"},
	{ID: "sacerdote", Name: "Sacerdote", Plural: "Sacerdotes", Tier: TierSecond, Family: "Noviço"},
	{ID: "sumo-sacerdote", Name: "Sumo Sacerdote", Plural: "Sumo Sacerdotes", Tier: TierTranscendent, Family: "Noviço"},
	{ID: "arcebispo", Name: "Arcebispo", Plural: "Arcebispos", Tier: TierThird, Family: "Noviço"},
	{ID: "cardeal", Name: "Cardeal", Plural: "Cardeais", Tier: TierFourth, Family: "Noviço"},
	{ID: "monge", Name: "Monge", Plural: "Monges", Tier: TierSecond, Family: "Noviço"},
	{ID: "mestre", Name: "Mestre", Plural: "Mestres", Tier: TierTranscendent, Family: "Noviço"},
	{ID: "shura", Name: "Shura", Plural: "Shuras", Tier: TierThird, Family: "Noviço"},
	{ID: "inquisidor", Name: "Inquisidor", Plural: "Inquisidores", Tier: TierFourth, Family: "Noviço"},

	// Arqueiro
	{ID: "arqueiro", Name: "Arqueiro", Plural: "Arqueiros", Tier: TierFirst, Family: "Arqueiro"},
	{ID: "cacador", Name: "Caçador", Plural: "Caçadores", Tier: TierSecond, Family: "Arqueiro"},
	{ID: "atirador-de-elite", Name: "Atirador de Elite", Plural: "Atiradores de Elite", Tier: TierTranscendent, Family: "Arqueiro"},
	{ID: "sentinela", Name: "Sentinela", Plural: "Sentinelas", Tier: TierThird, Family: "Arqueiro"},
	{ID: "falcao-do-vento", Name: "Falcão do Vento", Plural: "Falcões do Vento", Tier: TierFourth, Family: "Arqueiro"},
	{ID: "bardo", Name: "Bardo", Plural: "Bardos", Tier: TierSecond, Family: "Arqueiro"},
	{ID: "odalisca", Name: "Odalisca", Plural: "Odaliscas", Tier: TierSecond, Family: "Arqueiro"},
	{ID: "menestrel", Name: "Menestrel", Plural: "Menestréis", Tier: TierTranscendent, Family: "Arqueiro"},
	{ID: "cigana", Name: "Cigana", Plural: "Ciganas", Tier: TierTranscendent, Family: "Arqueiro"},
	{ID: "trovador", Name: "Trovador", Plural: "Trovadores", Tier: TierThird, Family: "Arqueiro"},
	{ID: "musa", Name: "Musa", Plural: "Musas", Tier: TierThird, Family: "Arqueiro"},
	{ID: "maestro", Name: "Maestro", Plural: "Maestros", Tier: TierFourth, Family: "Arqueiro"},
	{ID: "diva", Name: "Diva", Plural: "Divas", Tier: TierFourth, Family: "Arqueiro"},

	// Taekwon
	{ID: "taekwon", Name: "Taekwon", Plural: "Taekwons", Tier: TierExpanded, Family: "Taekwon"},
	{ID: "mestre-taekwon", Name: "Mestre Taekwon", Plural: "Mestres Taekwons", Tier: TierExpanded, Family: "Taekwon"},
	{ID: "mestre-estelar", Name: "Mestre Estelar", Plural: "Mestres Estelares", Tier: TierExpanded, Family: "Taekwon"},
	{ID: "mestre-celestial", Name: "Mestre Celestial", Plural: "Mestres Celestiais", Tier: TierExpanded, Family: "Taekwon"},
	{ID: "espiritualista", Name: "Espiritualista", Plural: "Espiritualistas", Tier: TierExpanded, Family: "Taekwon"},
	{ID: "ceifador-de-almas", Name: "Ceifador de Almas", Plural: "Ceifadores de Almas", Tier: TierExpanded, Family: "Taekwon"},
	{ID: "asceta-das-almas", Name: "Asceta das Almas", Plural: "Ascetas das Almas", Tier: TierExpanded, Family: "Taekwon"},

	// Aprendiz
	{ID: "superaprendiz", Name: "Superaprendiz", Plural: "Superaprendizes", Tier: TierExpanded, Family: "Aprendiz"},
	{ID: "superaprendiz-ex", Name: "Superaprendiz EX", Plural: "Superaprendizes EX", Tier: TierExpanded, Family: "Aprendiz"},
	{ID: "hiperaprendiz", Name: "Hiperaprendiz", Plural: "Hiperaprendizes", Tier: TierExpanded, Family: "Aprendiz"},

	// Justiceiro
	{ID: "justiceiro", Name: "Justiceiro", Plural: "Justiceiros", Tier: TierExpanded, Family: "Justiceiro"},
	{ID: "insurgente", Name: "Insurgente", Plural: "Insurgentes", Tier: TierExpanded, Family: "Justiceiro"},
	{ID: "guerrilheiro", Name: "Guerrilheiro", Plural: "Guerrilheiros", Tier: TierExpanded, Family: "Justiceiro"},

	// Ninja
	{ID: "ninja", Name: "Ninja", Plural: "Ninjas", Tier: TierExpanded, Family: "Ninja"},
	{ID: "kagerou", Name: "Kagerou", Plural: "Kagerou", Tier: TierExpanded, Family: "Ninja"},
	{ID: "oboro", Name: "Oboro", Plural: "Oboro", Tier: TierExpanded, Family: "Ninja"},
	{ID: "shinkiro", Name: "Shinkiro", Plural: "Shinkiro", Tier: TierExpanded, Family: "Ninja"},
	{ID: "shiranui", Name: "Shiranui", Plural: "Shiranui", Tier: TierExpanded, Family: "Ninja"},

	// Druida
	{ID: "druida", Name: "Druida", Plural: "Druidas", Tier: TierExpanded, Family: "Druida"},
	{ID: "karnos", Name: "Karnos", Plural: "Karnos", Tier: TierExpanded, Family: "Druida"},
	{ID: "alitea", Name: "Alitea", Plural: "Alitea", Tier: TierExpanded, Family: "Druida"},

	// Invocador
	{ID: "invocador", Name: "Invocador", Plural: "Invocadores", Tier: TierExpanded, Family: "Invocador"},
	{ID: "animista", Name: "Animista", Plural: "Animistas", Tier: TierExpanded, Family: "Invocador"},
}

var classByID = func() map[string]Class {
	m := make(map[string]Class, len(classes))
	for _, c := range classes {
		m[c.ID] = c
	}
	return m
}()

// Classes devolve uma cópia do catálogo, na ordem do bROWiki.
func Classes() []Class {
	return append([]Class(nil), classes...)
}

// ClassByID procura a classe pelo ID (RN-06).
func ClassByID(id string) (Class, bool) {
	c, ok := classByID[id]
	return c, ok
}

// DefaultPortrait é o retrato de quem não escolhe nenhum (RN-10).
const DefaultPortrait = "retrato-1"

// portraits são os retratos de exemplo, arte original do RO Lobby servida pelo web
// (RN-10, D-04). A lista definitiva vem depois.
var portraits = []string{"retrato-1", "retrato-2", "retrato-3", "retrato-4"}

// Portraits devolve uma cópia da lista de retratos, com o padrão primeiro.
func Portraits() []string {
	return append([]string(nil), portraits...)
}

// HasPortrait diz se o retrato está na lista (RN-10).
func HasPortrait(id string) bool {
	for _, p := range portraits {
		if p == id {
			return true
		}
	}
	return false
}
