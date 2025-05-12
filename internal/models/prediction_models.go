package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// BalancePrediction представляет прогноз баланса на один день
type BalancePrediction struct {
	Date    time.Time       `json:"date"`     // Дата прогноза
	Balance decimal.Decimal `json:"balance"`  // Прогнозируемый баланс на эту дату
}

// BalancePredictionResponse представляет ответ на запрос прогноза баланса
type BalancePredictionResponse struct {
	AccountID           string              `json:"account_id"`            // ID счета, для которого сделан прогноз
	CurrentBalance      decimal.Decimal     `json:"current_balance"`       // Текущий баланс счета
	PredictionDays      int                 `json:"prediction_days"`       // Количество дней в прогнозе
	AverageDailyOutflow decimal.Decimal     `json:"avg_daily_outflow"`     // Средняя дневная сумма исходящих операций
	AverageDailyInflow  decimal.Decimal     `json:"avg_daily_inflow"`      // Средняя дневная сумма входящих операций
	Predictions         []BalancePrediction `json:"predictions"`           // Ежедневные прогнозы баланса
}
