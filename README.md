# prj_withdrawal — Entrega 3: aplicación web refactorizada

Aplicación web académica en Go para retiro y abono en efectivo, con persistencia JSON íntegra, auditoría de transacciones y escenarios BDD.

## Alcance de esta entrega

- Interfaz web para consultar saldo e historial y procesar retiros y abonos.
- API HTTP: `GET /api/state`, `GET /api/transactions`, `POST /api/withdraw`, `POST /api/deposit`.
- Registro JSON con cuenta y transacciones en el mismo archivo; importes en centavos enteros.
- Escenarios Gherkin de US-01a, US-01b y US-04, pruebas automatizadas e informe IEEE en `docs/informe-ieee.md`.

## Requisitos y ejecución

Se requiere Go 1.23 o posterior. Desde la raíz del proyecto:

```powershell
go run ./cmd/web
```

Abre <http://127.0.0.1:8080>. Opciones: `go run ./cmd/web -address 127.0.0.1:8080 -data data/state.json`.

La cuenta demo se carga desde `internal/storage/default-state.json` cuando todavía no existe el archivo de estado. La aplicación es una demo académica local, no incluye autenticación y no debe exponerse directamente a Internet.

## BDD, pruebas y análisis

```powershell
go mod download
go test ./...
go vet ./...
```

Los escenarios están en `features/` y se ejecutan con Godog como parte de `go test ./...`. GitHub Actions automatiza pruebas con detector de condiciones de carrera y `go vet`.

## Crear el ZIP fuente

```powershell
./scripts/package.ps1 -OutputPath ./Entrega-3.zip -Force
```

El paquete incluye código, escenarios e informe, y excluye `.git`, datos locales y binarios.
