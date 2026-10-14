package catalog

// Server é um servidor do Ragnarok LATAM (spec servidor, RN-01). O ID é o que o personagem
// e o lobby guardam.
type Server struct {
	ID   string
	Name string
}

// DefaultServer é o servidor dos dados que existiam antes da feature (RN-15).
const DefaultServer = "freya"

// servers segue a ordem da RN-01.
var servers = []Server{
	{ID: "freya", Name: "Freya"},
	{ID: "nidhogg", Name: "Nidhogg"},
}

// Servers devolve uma cópia do catálogo, na ordem da RN-01.
func Servers() []Server {
	out := make([]Server, len(servers))
	copy(out, servers)
	return out
}

// ServerByID devolve o servidor do catálogo com esse ID.
func ServerByID(id string) (Server, bool) {
	for _, s := range servers {
		if s.ID == id {
			return s, true
		}
	}
	return Server{}, false
}
