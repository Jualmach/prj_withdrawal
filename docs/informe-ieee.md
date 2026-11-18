# Aplicación Web `prj_withdrawal`: retiro, abono e integridad de transacciones

**Julio Alberto Marquina Chahuayo** · Código: 20242017G  
Facultad de Ingeniería Industrial y de Sistemas · Ingeniería de Software  
Universidad Nacional de Ingeniería · Docente: Ing. Jorge Medianero Acosta  
Curso: SW608U — Lenguajes de Dominio Específico · Unidad III, Semana 7c  
**Entrega 3: Aplicación Web refactorizada**

## Resumen

Se presenta la versión final web de `prj_withdrawal`, una aplicación académica en Go para retiros y abonos en efectivo. La solución aplica reglas de negocio centralizadas, persistencia JSON íntegra y pruebas BDD/TDD. El estado de cuenta y su historial se guardan juntos, y cada operación se identifica como exitosa, rechazada por fondos insuficientes o inválida. Los escenarios de US-01a, US-01b y US-04 se conservan como criterios verificables de la funcionalidad final.

**Palabras clave—** Go, aplicación Web, BDD, Gherkin, TDD, JSON, transacciones.

## I. Introducción

El objetivo es ofrecer una interfaz Web para las operaciones bancarias de efectivo y mantener la trazabilidad entre historias de usuario, escenarios Gherkin, reglas de dominio y pruebas. Esta carpeta y el repositorio corresponden a la tercera entrega: el producto refactorizado como aplicación Web.

## II. Requisitos

| ID | Requisito | Criterio observable |
|---|---|---|
| US-01a | Retiro en efectivo | Se descuenta un importe positivo hasta el saldo disponible; si lo excede, se rechaza y el saldo permanece. |
| US-01b | Abono en efectivo | Un importe positivo incrementa el saldo. |
| US-04 | Integridad de transacción y datos | El JSON conserva cuenta e historial; registra éxito, rechazo y error de validación. |
| REF-01 | Aplicación Web | La interfaz presenta saldo e historial y permite las operaciones anteriores. |

Los importes admiten hasta dos decimales y se representan como centavos enteros para evitar errores de redondeo.

## III. Especificación BDD y pruebas

Los archivos `features/US-01a-retiro.feature`, `features/US-01b-abono.feature` y `features/US-04-integridad-transacciones.feature` contienen los escenarios Given/When/Then. Godog enlaza los escenarios con los pasos de `internal/bdd`; `go test ./...` ejecuta BDD y las pruebas unitarias de dominio, persistencia y servidor HTTP.

## IV. Arquitectura

- `internal/bank`: tipos, conversión monetaria y reglas de transacción.
- `internal/storage`: carga y persistencia del estado JSON.
- `internal/web`: interfaz HTML y API HTTP.
- `cmd/web`: punto de entrada del servidor.
- `features/`: especificación Gherkin.
- `docs/`: este informe IEEE.

El archivo de estado contiene la cuenta y el historial. La escritura usa un archivo temporal sincronizado y reemplazo final para no informar éxito si falla la persistencia. Los intentos rechazados o inválidos mantienen el saldo sin cambios y registran su resultado cuando el almacenamiento está disponible.

## V. Interfaz y API

La aplicación se ejecuta localmente en `127.0.0.1:8080`.

| Método | Ruta | Función |
|---|---|---|
| GET | `/` | Interfaz Web. |
| GET | `/api/state` | Cuenta, saldo e historial. |
| GET | `/api/transactions` | Historial JSON. |
| POST | `/api/withdraw` | Procesar un retiro. |
| POST | `/api/deposit` | Procesar un abono. |

## VI. Verificación y ejecución

Requiere Go 1.23 o posterior. Desde la raíz del proyecto:

```powershell
go run ./cmd/web
go test ./...
go vet ./...
```

La aplicación es una demostración académica local: no implementa autenticación ni está destinada a exposición directa a Internet. GitHub Actions ejecuta pruebas y análisis estático.

## VII. Conclusiones

La entrega ofrece una aplicación Web organizada por responsabilidades, mantiene el registro JSON de transacciones y proporciona escenarios BDD y pruebas automatizadas para retiro, abono e integridad. La estructura publicada contiene únicamente los fuentes y documentos necesarios para esta entrega final.

## Referencias

[1] Cucumber, “Gherkin Reference,” documentación oficial.  
[2] Go Authors, “The Go Programming Language Documentation,” documentación oficial.
