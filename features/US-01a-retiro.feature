Feature: US-01a - Retiro de efectivo
  Como cliente bancario
  Quiero retirar efectivo de mi cuenta
  Para disponer de mi dinero sin exceder el saldo disponible

  Scenario: Retiro aprobado con saldo suficiente
    Given la cuenta de Ana tiene un saldo de 100.00
    When Ana retira 35.50
    Then el retiro es aprobado
    And el nuevo saldo es 64.50

  Scenario: Retiro rechazado por saldo insuficiente
    Given la cuenta de Ana tiene un saldo de 100.00
    When Ana retira 120.00
    Then el retiro es rechazado por saldo insuficiente
    And el saldo permanece en 100.00
