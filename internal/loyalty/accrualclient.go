package loyalty

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type AccrualForOrder struct {
	Order         string  `json:"order"`
	Status        string  `json:"status"`
	Accrual       float32 `json:"accrual,omitempty"`
	RetryInterval int     `json:"retry,omitempty"`
}

type AccrualClient struct {
	config *AccrualClientConfig
}

type AccrualClientConfig struct {
	address string
}

var (
	ErrAccruallOrderNotExist        = errors.New("order not load in accruall system")
	ErrAccruallNumberRequestExceeds = errors.New("number or request in exceeeds")
	ErrAccruallInternalServError    = errors.New("internal error accrual service")
)

func NewAccuralClientConfig(accrualAddress string) *AccrualClientConfig {
	return &AccrualClientConfig{address: accrualAddress}
}

func NewAccuralClient(config *AccrualClientConfig) *AccrualClient {
	return &AccrualClient{config: config}
}

func (client *AccrualClient) GetAccural(number string) (*AccrualForOrder, error) {
	requestPattern := "%s/api/orders/%s"
	tr := &http.Transport{}
	httpClient := &http.Client{Transport: tr}
	response, err := httpClient.Get(fmt.Sprintf(requestPattern, client.config.address, number))
	if err != nil {
		return nil, fmt.Errorf("error send GET request. error:  %w", err)
	}
	defer response.Body.Close()
	accuralData := AccrualForOrder{}
	switch response.StatusCode {
	case 200:
		if err := json.NewDecoder(response.Body).Decode(&accuralData); err != nil {
			return nil, fmt.Errorf("error decode accural response. error:  %w", err)
		}
	case 204:
		return nil, ErrAccruallOrderNotExist
	case 429:
		retry := response.Header.Get("Retry-After")
		retryInterval, err := strconv.Atoi(retry)
		if err != nil {
			return nil, fmt.Errorf("error conver retry interval fro header. error:  %w", err)
		}
		accuralData.RetryInterval = retryInterval
		return &accuralData, ErrAccruallNumberRequestExceeds
	case 500:
		return nil, ErrAccruallInternalServError
	}
	return &accuralData, nil
}
