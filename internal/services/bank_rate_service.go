package services

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

const (
	cbrURL      = "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx"
	soapAction  = "http://web.cbr.ru/KeyRate"
	soapRequest = `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <KeyRate xmlns="http://web.cbr.ru/">
      <fromDate>%s</fromDate>
      <ToDate>%s</ToDate>
    </KeyRate>
  </soap:Body>
</soap:Envelope>`
)

// KeyRateResponse представляет XML-структуру ответа Центрального банка о ключевой ставке
type KeyRateResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName      xml.Name `xml:"Body"`
		KeyRateXML   string   `xml:"KeyRateResponse>KeyRateResult>diffgram>KeyRate"`
		KeyRateTable struct {
			XMLName xml.Name `xml:"KeyRate"`
			Rates   []struct {
				Date string  `xml:"DT"`
				Rate float64 `xml:"Rate"`
			} `xml:"KR"`
		} `xml:"KeyRateResponse>KeyRateResult>diffgram>KeyRate"`
	}
}

// Кэш для ключевой ставки, чтобы избежать частых внешних запросов
var cachedKeyRate struct {
	rate decimal.Decimal
	time time.Time
}
var keyRateMutex sync.Mutex

// FetchCentralBankRate получает текущую ключевую ставку из Центрального банка
// Возвращает ставку в виде decimal.Decimal и ошибку, если запрос не удался
func FetchCentralBankRate() (decimal.Decimal, error) {
	// Блокируем мьютекс для безопасного доступа к кэшу
	keyRateMutex.Lock()
	defer keyRateMutex.Unlock()

	// Проверяем, есть ли актуальное значение в кэше (не старше 1 часа)
	if !cachedKeyRate.rate.IsZero() && time.Since(cachedKeyRate.time) < time.Hour {
		log.Println("Используем кэшированное значение ключевой ставки")
		return cachedKeyRate.rate, nil
	}

	log.Println("Запрашиваем ключевую ставку из Центрального банка России")

	// Для демонстрационных целей используем запасное значение, если запрос не удастся
	fallbackRate := decimal.NewFromFloat(16.0)

	// Пытаемся получить актуальную ставку из Центрального банка
	rate, err := fetchKeyRateFromCBR()
	if err != nil {
		log.Printf("Не удалось получить ключевую ставку из ЦБ РФ: %v. Используем запасное значение: %s", err, fallbackRate.String())

		// Обновляем кэш запасным значением
		cachedKeyRate.rate = fallbackRate
		cachedKeyRate.time = time.Now()

		return fallbackRate, nil
	}

	// Обновляем кэш актуальным значением
	cachedKeyRate.rate = rate
	cachedKeyRate.time = time.Now()

	log.Printf("Успешно получена ключевая ставка из ЦБ РФ: %s", rate.String())
	return rate, nil
}

// fetchKeyRateFromCBR делает SOAP-запрос к Центральному банку России
// для получения текущей ключевой ставки
func fetchKeyRateFromCBR() (decimal.Decimal, error) {
	// Устанавливаем диапазон дат для запроса (сегодня и вчера)
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Форматируем SOAP-запрос
	soapBody := fmt.Sprintf(soapRequest, yesterday, today)

	// Создаем HTTP-запрос
	req, err := http.NewRequest("POST", cbrURL, bytes.NewBufferString(soapBody))
	if err != nil {
		return decimal.Zero, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", soapAction)

	// Отправляем запрос
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("не удалось отправить запрос: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return decimal.Zero, fmt.Errorf("не удалось прочитать ответ: %w", err)
	}

	// Используем регулярное выражение для извлечения ставки из XML
	// Это более простой подход, чем полный парсинг сложного SOAP-ответа
	re := regexp.MustCompile(`<Rate>([0-9,.]+)</Rate>`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	if len(matches) == 0 {
		return decimal.Zero, fmt.Errorf("ключевая ставка не найдена в ответе")
	}

	// Получаем самую последнюю ставку
	rateStr := matches[len(matches)-1][1]
	rateStr = strings.Replace(rateStr, ",", ".", -1) // Заменяем запятую на точку для парсинга десятичного числа

	// Парсим ставку как decimal
	rate, err := decimal.NewFromString(rateStr)
	if err != nil {
		return decimal.Zero, fmt.Errorf("не удалось распарсить ставку '%s': %w", rateStr, err)
	}

	return rate, nil
}
