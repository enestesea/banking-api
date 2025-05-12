package utils

import (
	"time"

	"github.com/shopspring/decimal"

	"bankapp/internal/models"
)

// CalculateMonthlyPayment Рассчитывает ежемесячный платеж по кредиту
func CalculateMonthlyPayment(loanAmount decimal.Decimal, annualRate decimal.Decimal, termMonths int) decimal.Decimal {
	if termMonths <= 0 {
		return decimal.Zero
	}
	monthlyRate := annualRate.Div(decimal.NewFromInt(12)).Div(decimal.NewFromInt(100))

	if monthlyRate.IsZero() {
		return loanAmount.Div(decimal.NewFromInt(int64(termMonths)))
	}

	onePlusRate := decimal.NewFromInt(1).Add(monthlyRate)
	powOnePlusRate := onePlusRate.Pow(decimal.NewFromInt(int64(termMonths)))

	numerator := monthlyRate.Mul(powOnePlusRate)
	denominator := powOnePlusRate.Sub(decimal.NewFromInt(1))

	if denominator.IsZero() {
		return decimal.Zero
	}

	monthlyPayment := loanAmount.Mul(numerator.Div(denominator))

	return monthlyPayment.RoundBank(2)
}

// GeneratePaymentSchedule генерирует график платежей по кредиту
// Принимает сумму кредита, годовую процентную ставку, срок в месяцах, дату начала и ежемесячный платеж
// Возвращает срез объектов Payment, представляющих график платежей
func GeneratePaymentSchedule(loanAmount decimal.Decimal, annualRate decimal.Decimal, termMonths int, startDate time.Time, monthlyPayment decimal.Decimal) []models.Payment {
	schedule := make([]models.Payment, 0, termMonths)
	remainingPrincipal := loanAmount
	monthlyRate := annualRate.Div(decimal.NewFromInt(12)).Div(decimal.NewFromInt(100))

	for i := 0; i < termMonths; i++ {
		dueDate := startDate.AddDate(0, i+1, 0)

		interestPart := remainingPrincipal.Mul(monthlyRate).RoundBank(2)
		principalPart := monthlyPayment.Sub(interestPart)

		if i == termMonths-1 || remainingPrincipal.Sub(principalPart).LessThanOrEqual(decimal.Zero) {
			principalPart = remainingPrincipal
			monthlyPayment = principalPart.Add(interestPart).RoundBank(2)
		}

		payment := models.Payment{
			DueDate:       dueDate,
			Amount:        monthlyPayment,
			InterestPart:  interestPart,
			PrincipalPart: principalPart,
			Paid:          false,
		}
		schedule = append(schedule, payment)

		remainingPrincipal = remainingPrincipal.Sub(principalPart)
		if remainingPrincipal.LessThanOrEqual(decimal.Zero) {
			break
		}
	}
	return schedule
}
