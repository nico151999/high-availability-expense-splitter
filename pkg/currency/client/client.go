package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
)

var _ Client = (*client)(nil)

type Client interface {
	FetchCurrencies(context.Context) (map[string]string, error)
	GetExchangeRate(context.Context, string, string, time.Time) (float64, error)
	GetLatestExchangeRate(context.Context, string, string) (float64, error)
}

type client struct {
	httpClient http.Client
}

var ErrCurrencyExchangeRateNotFound = eris.New("could not find currency exchange rate")

func NewCurrencyClient() *client {
	return &client{
		httpClient: http.Client{},
	}
}

func (c *client) FetchCurrencies(ctx context.Context) (map[string]string, error) {
	log := logging.FromContext(ctx)

	req, err := http.NewRequest(http.MethodGet, "https://cdn.jsdelivr.net/gh/fawazahmed0/currency-api@1/latest/currencies.min.json", nil)
	if err != nil {
		msg := "could not create request to fetch currencies"
		log.Error(msg)
		return nil, eris.Wrap(err, msg)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		msg := "could not perform request to fetch currencies"
		log.Error(msg)
		return nil, eris.Wrap(err, msg)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		msg := "could not read response body after fetching currencies"
		log.Error(msg)
		return nil, eris.Wrap(err, msg)
	}

	var currencies map[string]string
	if err := json.Unmarshal(body, &currencies); err != nil {
		msg := "could not unmarshal response body after fetching currencies"
		log.Error(msg)
		return nil, eris.Wrap(err, msg)
	}

	for acronym, name := range currencies {
		if name == "" {
			delete(currencies, acronym)
		}
	}

	return currencies, nil
}

func (c *client) GetExchangeRate(ctx context.Context, src string, dest string, date time.Time) (float64, error) {
	return c.getExchangeRate(ctx, src, dest, date.Format("2006-01-02"))
}

func (c *client) GetLatestExchangeRate(ctx context.Context, src string, dest string) (float64, error) {
	return c.getExchangeRate(ctx, src, dest, "latest")
}

func (c *client) getExchangeRate(ctx context.Context, src string, dest string, date string) (float64, error) {
	src = strings.ToLower(src)
	dest = strings.ToLower(dest)
	log := logging.FromContext(ctx).With(
		logging.String("srcAcronym", src),
		logging.String("destAcronym", dest),
		logging.String("date", date),
	)

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"https://cdn.jsdelivr.net/gh/fawazahmed0/currency-api@1/%s/currencies/%s/%s.json",
			date,
			src,
			dest),
		nil)
	if err != nil {
		msg := "could not create request to get exchange rate"
		log.Error(msg)
		return 0, eris.Wrap(err, msg)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		msg := "could not perform request to get exchange rate"
		log.Error(msg)
		return 0, eris.Wrap(err, msg)
	}
	if res.StatusCode == http.StatusNotFound {
		log.Error("could not find currency exchange rate")
		return 0, ErrCurrencyExchangeRateNotFound
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		msg := "could not read response body after getting exchange rate"
		log.Error(msg)
		return 0, eris.Wrap(err, msg)
	}

	var rates map[string]interface{}
	if err := json.Unmarshal(body, &rates); err != nil {
		msg := "could not unmarshal response body after getting exchange rate"
		log.Error(msg)
		return 0, eris.Wrap(err, msg)
	}
	if rate, ok := rates[dest]; ok {
		if rate, ok := rate.(float64); ok {
			return rate, nil
		} else {
			msg := "exchange rate has invalid data format"
			log.Error(msg)
			return 0, eris.Wrap(err, msg)
		}
	} else {
		msg := "exchange rate not in result"
		log.Error(msg)
		return 0, eris.Wrap(err, msg)
	}
}
