// Comando healthcheck consulta o GET /healthz da API local e sai com 0 só se a resposta
// for 200. A imagem da API (distroless) não tem shell nem curl, então o healthcheck do
// Docker Compose usa este binário (D-01 da spec home-local).
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	if err := check(context.Background(), "http://127.0.0.1:"+port+"/healthz"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func check(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthz respondeu %d", resp.StatusCode)
	}
	return nil
}
