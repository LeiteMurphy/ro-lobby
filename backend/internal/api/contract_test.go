package api

import (
	"context"
	"slices"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

const contractPath = "../../../openapi.yaml"

func loadContract(t *testing.T) *openapi3.T {
	t.Helper()
	doc, err := openapi3.NewLoader().LoadFromFile(contractPath)
	if err != nil {
		t.Fatalf("carregar %s: %v", contractPath, err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("contrato inválido: %v", err)
	}
	return doc
}

// CA-04.1: GET /healthz está descrita com as respostas 200 e 503 e seus corpos.
func TestContrato_CA04_1_HealthzDescrita(t *testing.T) {
	doc := loadContract(t)

	path := doc.Paths.Find("/healthz")
	if path == nil || path.Get == nil {
		t.Fatal("GET /healthz não está no contrato")
	}

	wantBodies := map[int]map[string]string{
		200: {"status": "ok", "database": "ok"},
		503: {"status": "degraded", "database": "unavailable"},
	}
	for code, want := range wantBodies {
		resp := path.Get.Responses.Status(code)
		if resp == nil || resp.Value == nil {
			t.Errorf("resposta %d não descrita", code)
			continue
		}
		media := resp.Value.Content.Get("application/json")
		if media == nil || media.Schema == nil {
			t.Errorf("resposta %d sem corpo JSON", code)
			continue
		}
		schema := media.Schema.Value
		for field, value := range want {
			prop, ok := schema.Properties[field]
			if !ok {
				t.Errorf("resposta %d: campo %q ausente", code, field)
				continue
			}
			if !slices.Contains(prop.Value.Enum, any(value)) {
				t.Errorf("resposta %d: %q não aceita o valor %q", code, field, value)
			}
			if !slices.Contains(schema.Required, field) {
				t.Errorf("resposta %d: campo %q deveria ser obrigatório", code, field)
			}
		}
		if err := schema.VisitJSON(toAny(want)); err != nil {
			t.Errorf("resposta %d: corpo %v não bate com o schema: %v", code, want, err)
		}
	}
}

func toAny(m map[string]string) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
