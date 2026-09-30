package config

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// LoadDotEnv lê um arquivo .env e define as variáveis que ainda não existem no
// ambiente. Variáveis já definidas têm precedência, então a CI e a produção não são
// afetadas por um .env esquecido. Um arquivo inexistente não é erro.
func LoadDotEnv(path string) error {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }() // só leitura: o erro do Close não muda o resultado

	vars, err := parseDotEnv(bufio.NewScanner(f))
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	for key, value := range vars {
		if _, set := os.LookupEnv(key); !set {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseDotEnv aceita linhas KEY=VALUE, comentários com # e linhas em branco. Aspas
// simples ou duplas em volta do valor são removidas.
func parseDotEnv(sc *bufio.Scanner) (map[string]string, error) {
	vars := map[string]string{}
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("linha %d: esperado KEY=VALUE", n)
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
		vars[key] = value
	}
	return vars, sc.Err()
}
