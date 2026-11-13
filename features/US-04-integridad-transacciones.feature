Feature: US-04 - Integridad de transacción y datos
  Como operador del sistema
  Quiero registrar cada operación en formato JSON
  Para auditar operaciones aprobadas, rechazadas y errores sin alterar el saldo

  Scenario: Registrar una operación aprobada
    Given la cuenta de Ana tiene un saldo de 100.00
    When Ana abona 25.75 en efectivo
    Then el resultado aprobado y el saldo 125.75 quedan guardados juntos
    And la transacción se registra en JSON

  Scenario: Registrar un rechazo por saldo insuficiente
    Given la cuenta de Ana tiene un saldo de 100.00
    When Ana retira 120.00
    Then se registra el resultado rechazado en JSON
    And el saldo permanece en 100.00

  Scenario: Registrar un error de validación
    Given la cuenta de Ana tiene un saldo de 100.00
    When Ana intenta abonar un importe de 0.00
    Then se registra el resultado error en JSON
    And el saldo permanece en 100.00
