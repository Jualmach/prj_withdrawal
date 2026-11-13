# prj_withdrawal

Proyecto incremental de Semana 7c: retiro y abono en efectivo, registro JSON de transacciones y refactorización a una aplicación Web en Go.

## Entregables incluidos

1. **Punto de partida — CLI:** retiro de efectivo (US-01a) y consulta de saldo.
2. **Incremento — Abono:** abono en efectivo (US-01b).
3. **Incremento — Integridad:** registro JSON de operaciones exitosas, rechazadas y errores (US-04); el saldo y el historial se guardan en un mismo archivo para evitar actualizaciones parciales.
4. **Refactorización — App Web:** interfaz HTTP que reutiliza el mismo dominio, persistencia y reglas del CLI.
5. **Calidad y entrega:** escenarios Gherkin, pruebas Go, informe estilo IEEE y script para generar el ZIP fuente.

Los escenarios se encuentran en `features/`. Las pruebas ejecutables TDD/BDD se encuentran junto a los paquetes Go en `internal/`.

## Requisitos

- Go 1.23 o posterior.
- PowerShell 5.1 o posterior para crear el ZIP en Windows.

El proyecto solo usa la biblioteca estándar de Go; no requiere descargar dependencias.

## Ejecutar el CLI

Desde esta carpeta:

```powershell
go run ./cmd/cli
```

Opciones del menú: retirar, abonar, consultar saldo, ver transacciones o salir. El importe acepta hasta dos decimales (por ejemplo, `25.75`). La cuenta demo comienza con S/ 1,000.00.

## Ejecutar la aplicación Web

En una terminal:

```powershell
go run ./cmd/web
```

Abre <http://127.0.0.1:8080>. La aplicación permite retirar y abonar, muestra el saldo y presenta el historial. También expone `GET /api/state`, `GET /api/transactions`, `POST /api/withdraw` y `POST /api/deposit`.

El servidor se enlaza por defecto a `127.0.0.1` y no incluye autenticación: es una demo académica local, no un sistema listo para exponer en Internet. Ejecuta una sola instancia a la vez para evitar escrituras concurrentes desde varios procesos.

## Datos y transacciones

La aplicación lee y escribe `data/state.json`. Si el archivo aún no existe, carga la cuenta demo desde `internal/storage/default-state.json`. El estado contiene la cuenta y una lista JSON de transacciones:

```json
{
  "account": {
    "id": "acct-001",
    "owner": "Cliente",
    "balance_cents": 100000
  },
  "transactions": []
}
```

Los importes se representan como centavos enteros para evitar errores de redondeo de punto flotante. Cada registro incluye identificador, tipo, resultado (`success`, `declined` o `error`), importe, saldo posterior, fecha UTC y mensaje. El archivo completo se escribe en un temporal y se reemplaza después de sincronizarlo; no se presenta una operación como exitosa si falla el guardado.

## Pruebas y análisis estático

```powershell
go test ./...
go vet ./...
```

## Crear el ZIP fuente

```powershell
.\scripts\package.ps1
```

Por defecto, crea `prj_withdrawal-source.zip` junto a la carpeta del proyecto. Para reemplazar el ZIP generado en una ejecución anterior, agrega `-Force`. Puede indicarse otra ruta:

```powershell
.\scripts\package.ps1 -OutputPath 'C:\ruta\prj_withdrawal-source.zip'
```

El ZIP omite `.git`, binarios y el estado de datos local.

## Publicación

El código está preparado para publicarse como repositorio GitHub y cuenta con CI para pruebas y `go vet`. La publicación requiere que GitHub permita crear un repositorio nuevo con la sesión autenticada del propietario. El informe no afirma que el repositorio ya esté publicado.
