package catalog

import (
	"cmp"
	"slices"
)

// Reset é o tempo de retorno da instância, como o bROWiki agrupa (spec lobbies, RN-01).
type Reset string

const (
	ResetDaily     Reset = "daily"
	ResetThreeDays Reset = "three_days"
	ResetHours     Reset = "hours"
	ResetWeekly    Reset = "weekly"
)

// HighLevel é o nível de entrada a partir do qual a instância vem antes do separador na
// escolha (RN-02).
const HighLevel = 130

// Instance é uma instância do catálogo. O ID é o nome sem acento em kebab-case e é o que
// o lobby guarda, junto com uma cópia do nome e do nível (design D-02).
type Instance struct {
	ID    string
	Name  string
	Level int
	Reset Reset
}

// instances segue https://browiki.org/wiki/Inst%C3%A2ncias (consultado em 2026-10-06),
// só com as que aceitam grupo: as marcadas como "Solo" ficam de fora (RN-01). Nomes do
// bROWiki, que acompanha o Ragnarok LATAM (CLAUDE.md).
var instances = []Instance{
	// Retorno diário
	{ID: "vila-dos-porings", Name: "Vila dos Porings", Level: 30, Reset: ResetDaily},
	{ID: "batalha-dos-orcs", Name: "Batalha dos Orcs", Level: 60, Reset: ResetDaily},
	{ID: "memorias-de-sarah", Name: "Memórias de Sarah", Level: 99, Reset: ResetDaily},
	{ID: "base-militar", Name: "Base Militar", Level: 100, Reset: ResetDaily},
	{ID: "laboratorio-werner", Name: "Laboratório Werner", Level: 100, Reset: ResetDaily},
	{ID: "missao-os", Name: "Missão OS", Level: 110, Reset: ResetDaily},
	{ID: "memorial-cor", Name: "Memorial COR", Level: 110, Reset: ResetDaily},
	{ID: "sonho-sombrio", Name: "Sonho Sombrio", Level: 120, Reset: ResetDaily},
	{ID: "aos-pes-do-rei", Name: "Aos Pés do Rei", Level: 130, Reset: ResetDaily},
	{ID: "torre-do-demonio", Name: "Torre do Demônio", Level: 130, Reset: ResetDaily},
	{ID: "ortus-aqua", Name: "Ortus Aqua", Level: 130, Reset: ResetDaily},
	{ID: "jardim-secreto", Name: "Jardim Secreto", Level: 130, Reset: ResetDaily},
	{ID: "fazenda-de-pitayas", Name: "Fazenda de Pitayas", Level: 130, Reset: ResetDaily},
	{ID: "duelo-com-sweety", Name: "Duelo com Sweety", Level: 130, Reset: ResetDaily},
	{ID: "caverna-de-buwaya", Name: "Caverna de Buwaya", Level: 130, Reset: ResetDaily},
	{ID: "maldicao-de-glastheim", Name: "Maldição de Glastheim", Level: 130, Reset: ResetDaily},
	{ID: "ruina-de-glastheim", Name: "Ruína de Glastheim", Level: 130, Reset: ResetDaily},
	{ID: "covil-de-vermes", Name: "Covil de Vermes", Level: 140, Reset: ResetDaily},
	{ID: "fabrica-do-terror", Name: "Fábrica do Terror", Level: 140, Reset: ResetDaily},
	{ID: "laboratorio-central", Name: "Laboratório Central", Level: 140, Reset: ResetDaily},
	{ID: "sala-final", Name: "Sala Final", Level: 150, Reset: ResetDaily},
	{ID: "ilha-bios", Name: "Ilha Bios", Level: 160, Reset: ResetDaily},
	{ID: "caverna-de-mors", Name: "Caverna de Mors", Level: 160, Reset: ResetDaily},
	{ID: "templo-do-demonio-rei", Name: "Templo do Demônio Rei", Level: 160, Reset: ResetDaily},
	{ID: "edda-do-biolaboratorio", Name: "Edda do Biolaboratório", Level: 170, Reset: ResetDaily},
	{ID: "purificacao-do-santuario", Name: "Purificação do Santuário", Level: 170, Reset: ResetDaily},
	{ID: "memorias-de-thanatos", Name: "Memórias de Thanatos", Level: 180, Reset: ResetDaily},
	{ID: "mansao-da-desilusao", Name: "Mansão da Desilusão", Level: 200, Reset: ResetDaily},
	{ID: "mausoleu-das-magoas", Name: "Mausoléu das Mágoas", Level: 220, Reset: ResetDaily},

	// Retorno em 3 dias
	{ID: "torre-submersa", Name: "Torre Submersa", Level: 40, Reset: ResetThreeDays},
	{ID: "ninho-de-nidhogg", Name: "Ninho de Nidhogg", Level: 70, Reset: ResetThreeDays},
	{ID: "laboratorio-de-wolfchev", Name: "Laboratório de Wolfchev", Level: 90, Reset: ResetThreeDays},
	{ID: "fortaleza-voadora", Name: "Fortaleza Voadora", Level: 145, Reset: ResetThreeDays},
	{ID: "glastheim-sombria", Name: "Glastheim Sombria", Level: 160, Reset: ResetThreeDays},
	{ID: "queda-de-glastheim", Name: "Queda de Glastheim", Level: 170, Reset: ResetThreeDays},
	{ID: "glastheim-infernal", Name: "Glastheim Infernal", Level: 170, Reset: ResetThreeDays},
	{ID: "queda-do-aeroplano", Name: "Queda do Aeroplano", Level: 200, Reset: ResetThreeDays},
	{ID: "arena-noturna", Name: "Arena Noturna", Level: 210, Reset: ResetThreeDays},
	{ID: "torre-da-constelacao", Name: "Torre da Constelação", Level: 240, Reset: ResetThreeDays},

	// Retorno em horas
	{ID: "altar-do-selo", Name: "Altar do Selo", Level: 75, Reset: ResetHours},
	{ID: "caverna-do-polvo", Name: "Caverna do Polvo", Level: 90, Reset: ResetHours},
	{ID: "esgotos-de-malangdo", Name: "Esgotos de Malangdo", Level: 90, Reset: ResetHours},
	{ID: "labirinto-da-neblina", Name: "Labirinto da Neblina", Level: 99, Reset: ResetHours},
	{ID: "espaco-infinito", Name: "Espaço Infinito", Level: 100, Reset: ResetHours},

	// Retorno semanal
	{ID: "cripta", Name: "Cripta", Level: 60, Reset: ResetWeekly},
	{ID: "glastheim-infantil", Name: "Glastheim Infantil", Level: 65, Reset: ResetWeekly},
	{ID: "fabrica-infantil", Name: "Fábrica Infantil", Level: 70, Reset: ResetWeekly},
	{ID: "tumulo-do-monarca", Name: "Túmulo do Monarca", Level: 99, Reset: ResetWeekly},
	{ID: "hospital-abandonado", Name: "Hospital Abandonado", Level: 100, Reset: ResetWeekly},
	{ID: "lago-de-bakonawa", Name: "Lago de Bakonawa", Level: 140, Reset: ResetWeekly},
	{ID: "sarah-vs-fenril", Name: "Sarah vs Fenril", Level: 145, Reset: ResetWeekly},
}

var instanceByID = func() map[string]Instance {
	m := make(map[string]Instance, len(instances))
	for _, i := range instances {
		m[i.ID] = i
	}
	return m
}()

// Instances devolve o catálogo na ordem da escolha (RN-02): primeiro as de nível
// HighLevel ou mais, depois as de nível menor; em cada parte, nível decrescente e, no
// empate, nome.
func Instances() []Instance {
	out := append([]Instance(nil), instances...)
	slices.SortStableFunc(out, func(a, b Instance) int {
		if ah, bh := a.Level >= HighLevel, b.Level >= HighLevel; ah != bh {
			if ah {
				return -1
			}
			return 1
		}
		if c := cmp.Compare(b.Level, a.Level); c != 0 {
			return c
		}
		return cmp.Compare(a.Name, b.Name)
	})
	return out
}

// InstanceByID procura a instância pelo ID (RN-01).
func InstanceByID(id string) (Instance, bool) {
	i, ok := instanceByID[id]
	return i, ok
}
