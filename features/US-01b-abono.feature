Feature: US-01b - Abono en efectivo
  Como cliente bancario
  Quiero abonar efectivo a mi cuenta
  Para incrementar mi saldo disponible

  Scenario: Abono aprobado
    Given la cuenta de Ana tiene un saldo de 100.00
    When Ana abona 25.75 en efectivo
    Then el abono es aprobado
    And el nuevo saldo es 125.75
