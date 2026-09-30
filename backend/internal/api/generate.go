// Package api contém o código gerado a partir do openapi.yaml (ADR-05, RN-11, RN-12).
// Não edite api.gen.go à mão: mude o contrato e rode `go generate ./...`.
package api

//go:generate go tool oapi-codegen -config oapi.cfg.yaml ../../../openapi.yaml
