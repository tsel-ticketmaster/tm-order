package midtrans

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type MidtransRepository interface {
	Charge(ctx context.Context, req ChargeRequest) (ChargeResponse, error)
}

type midtransRepository struct {
	baseURL      string
	basicAuthKey string
	logger       *logrus.Logger
	hc           *http.Client
}

func NewMidtransRepository(baseURL string, basicAuthKey string, logger *logrus.Logger, hc *http.Client) MidtransRepository {
	return &midtransRepository{
		baseURL:      baseURL,
		basicAuthKey: basicAuthKey,
		logger:       logger,
		hc:           hc,
	}
}

// Charge implements MidtransRepository.
func (r *midtransRepository) Charge(ctx context.Context, req ChargeRequest) (ChargeResponse, error) {
	reqBuff, _ := json.Marshal(req)
	body := bytes.NewBuffer(reqBuff)
	url := fmt.Sprintf("%s/v2/charge", r.baseURL)

	hr, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return ChargeResponse{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while charge payment through midtrans")
	}

	hr.Header.Add("Content-Type", "application/json")
	hr.Header.Add("Accept", "application/json")
	hr.Header.Add("Authorization", fmt.Sprintf("Basic %s", r.basicAuthKey))

	hresp, err := r.hc.Do(hr)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return ChargeResponse{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while charge payment through midtrans")
	}

	defer hresp.Body.Close()

	respBody, err := io.ReadAll(hresp.Body)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return ChargeResponse{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while charge payment through midtrans")
	}

	if hresp.StatusCode < 200 && hresp.StatusCode > 299 {
		err := fmt.Errorf(string(respBody))
		r.logger.WithContext(ctx).WithError(err).Error()
		return ChargeResponse{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while charge payment through midtrans")
	}

	var resp ChargeResponse

	errBody := json.Unmarshal(respBody, &resp)
	if errBody != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return ChargeResponse{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while charge payment through midtrans")
	}

	return resp, nil
}
