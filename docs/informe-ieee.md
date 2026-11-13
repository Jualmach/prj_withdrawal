# Diseño e implementación incremental de `prj_withdrawal`

**Julio Alberto Marquina Chahuayo** · Código: 20242017G  
Facultad de Ingeniería Industrial y de Sistemas · Ingeniería de Software  
Universidad Nacional de Ingeniería · julio.marquina.c@uni.pe  
**Docente:** Ing. Jorge Medianero Acosta  
**Curso:** SW608U — Lenguajes de Dominio Específico · **Sección:** U · **Unidad III, Semana 7c**  
**Fecha de elaboración:** 12 de noviembre de 2026  
**Tipo de aplicación final:** Aplicación Web en Go

## Resumen

Este informe describe el desarrollo incremental de `prj_withdrawal`, una aplicación académica para operaciones bancarias de retiro y abono en efectivo. Se implementó primero una interfaz CLI interactiva en Go y luego se incorporaron el abono (US-01b), el registro JSON y la integridad del estado (US-04). Finalmente, se refactorizó la solución para ofrecer una interfaz Web que reutiliza el dominio y la persistencia del CLI. Los escenarios Gherkin definen los comportamientos esperados y las pruebas Go automatizan sus criterios principales. La solución usa centavos enteros y persiste el saldo y el historial en un único documento JSON reemplazado de manera atómica.

**Palabras clave—** BDD, Gherkin, Go, CLI, transacción, JSON, aplicación Web, TDD.

## I. Introducción

El objetivo es convertir la especificación de retiro y abono en comportamientos verificables, manteniendo trazabilidad entre historias de usuario, escenarios Gherkin, código y pruebas. El desarrollo sigue un proceso iterativo e incremental: el CLI es el punto de partida; las operaciones y la auditoría se agregan en incrementos; la interfaz Web se construye después como refactorización, sin duplicar las reglas de negocio.

## II. Alcance y requisitos

El producto académico cubre una cuenta demo local y una sola instancia de ejecución. No implementa usuarios, autenticación, conexión bancaria, cajero físico ni despliegue en un servicio de hosting.

| ID | Historia / requisito | Criterio observable |
|---|---|---|
| US-01a | Retiro en efectivo | Se descuenta el importe si hay saldo; si no, se rechaza y el saldo queda intacto. |
| US-01b | Abono en efectivo | Un importe positivo incrementa el saldo. |
| US-04 | Integridad y datos | El JSON conserva cuenta e historial juntos; registra éxitos, rechazos y errores de validación. |
| REF-01 | Refactorización Web | La UI Web invoca el mismo servicio de dominio y presenta saldo e historial. |

Todos los importes deben ser positivos y tener hasta dos decimales. Los importes se convierten a centavos enteros antes de modificar el saldo. Un retiro que supera el saldo se registra como `declined`; los datos inválidos se registran como `error`.

## III. Proceso de desarrollo BDD/TDD

Se redactaron escenarios de negocio en `features/US-01a-retiro.feature`, `features/US-01b-abono.feature` y `features/US-04-integridad-transacciones.feature`. Godog (Cucumber para Go) carga esos archivos y enlaza las condiciones Given/When/Then con los pasos definidos en `internal/bdd`; se ejecutan con `go test ./...`. Las pruebas unitarias adicionales comprueban reglas de dominio, CLI, servidor y persistencia.

La secuencia incremental es:

1. **Punto de partida:** cuenta y retiro por CLI, con saldo consultable.
2. **Incremento US-01b:** abono en efectivo y verificación del saldo resultante.
3. **Incremento US-04:** identificadores y resultados auditables en JSON; fallos y errores no alteran indebidamente el saldo.
4. **Refactorización REF-01:** servidor HTTP y UI Web sobre los mismos paquetes de dominio y almacenamiento.

## IV. Diseño

La solución se organiza en cuatro capas:

- `internal/bank`: cuenta, importes, servicio y reglas de transacción.
- `internal/storage`: lectura y escritura del estado JSON.
- `internal/cli`: menú interactivo.
- `internal/web`: UI HTML y API HTTP.
- `cmd/cli` y `cmd/web`: puntos de entrada independientes que comparten el servicio.

El documento `data/state.json` agrupa la cuenta y el historial. La escritura se realiza a un archivo temporal dentro del mismo directorio, se sincroniza y se reemplaza al finalizar. De este modo, el saldo y su registro no se guardan en archivos separados. El modelo usa `int64` en centavos para los cálculos monetarios.

### A. Flujo de retiro

1. El CLI o la Web recibe el importe y lo convierte a centavos.
2. El servicio carga el estado y valida cuenta e importe.
3. Si el retiro supera el saldo, agrega un evento `declined` sin modificarlo.
4. Si el saldo es suficiente, actualiza el saldo y agrega un evento `success`.
5. El estado completo se guarda como una sola operación de archivo.

### B. Flujo de abono

El servicio valida que el importe sea positivo, comprueba el límite entero para prevenir desbordamiento, suma el importe y persiste el saldo junto con el evento `success`.

### C. Manejo de errores

Los importes no válidos y los identificadores de cuenta incorrectos generan eventos `error` cuando la fuente de datos está disponible. Si el almacenamiento falla, el servicio devuelve el error y no reporta éxito. No es posible persistir un evento de error cuando el propio medio de almacenamiento está indisponible.

## V. Implementación de la aplicación Web

La aplicación se inicia localmente en `127.0.0.1:8080`. La interfaz permite retirar o abonar y consulta saldo e historial mediante los endpoints:

| Método | Ruta | Propósito |
|---|---|---|
| GET | `/` | Interfaz Web. |
| GET | `/api/state` | Cuenta, saldo e historial. |
| GET | `/api/transactions` | Historial de transacciones. |
| POST | `/api/withdraw` | Registrar un retiro. |
| POST | `/api/deposit` | Registrar un abono. |

La aplicación es demostrativa: no contiene autenticación ni se debe exponer directamente a Internet. GitHub aloja el código fuente y la automatización CI; GitHub Pages no ejecuta el backend Go.

## VI. Verificación

Las pruebas cubren retiros aprobados y rechazados, abonos, errores de validación, fallos de persistencia, conversión monetaria, lectura y escritura JSON, y rutas HTTP/CLI.

Comandos de validación:

```text
go test ./...
go vet ./...
```

El workflow de GitHub Actions ejecuta `go test -race -cover ./...` (incluidos los escenarios Godog) y `go vet ./...` en cada push y pull request. El ZIP se genera con `scripts/package.ps1`; no contiene binarios ni estado local.

## VII. Limitaciones y trabajo futuro

La persistencia JSON está destinada a una demo de una sola instancia; varios procesos que escriban simultáneamente al mismo archivo no están coordinados. La cuenta y la sesión son fijas, y el servicio Web no implementa autenticación, autorización, límites de tasa ni despliegue. Antes de un uso productivo, se recomienda una base de datos transaccional, control de concurrencia, gestión de identidades, HTTPS, auditoría protegida y pruebas de despliegue.

## VIII. Conclusiones

El proyecto demuestra un flujo incremental desde CLI hasta Web y mantiene las reglas de negocio centralizadas. Los escenarios Gherkin hacen explícitos los resultados esperados para retiro, abono e integridad; las pruebas automatizadas verifican los cambios de saldo y el registro JSON. El entregable incluye código fuente, informe y script reproducible de empaquetado.

## Referencias

[1] Cucumber, “Gherkin reference,” documentación oficial.  
[2] Go Authors, “The Go Programming Language Documentation,” documentación oficial.
