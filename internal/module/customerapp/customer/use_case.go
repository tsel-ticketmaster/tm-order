package customer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/internal/pkg/jwt"
	"github.com/tsel-ticketmaster/tm-order/internal/pkg/session"
	"github.com/tsel-ticketmaster/tm-order/internal/pkg/util"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/pubsub"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type CustomerUseCase interface {
	SignUp(ctx context.Context, req SignUpRequest) (SignUpResponse, error)
	SignIn(ctx context.Context, req SignInRequest) (SignInResponse, error)
	SignOut(ctx context.Context) error
	GetProfile(ctx context.Context) (GetProfileResponse, error)
	UpdateProfile(ctx context.Context, req UpdateProfileRequest) error
	ChangeEmail(ctx context.Context, req ChangeEmailRequest) (ChangeEmailResponse, error)
	ChangePassword(ctx context.Context, req ChangePasswordRequest) error
	Verify(ctx context.Context, req VerifyRequest) error
	VerifyChangeEmail(ctx context.Context, req ChangeEmailVerificationRequest) error
}

type CustomerUseCaseProperty struct {
	AppName            string
	Logger             *logrus.Logger
	Timeout            time.Duration
	TMUserBaseURL      string
	CryptoSecret       string
	JSONWebToken       *jwt.JSONWebToken
	Session            session.Session
	Cache              redis.UniversalClient
	Publisher          pubsub.Publisher
	CustomerRepository CustomerRepository
}

type customerUseCase struct {
	appName            string
	logger             *logrus.Logger
	timeout            time.Duration
	tmuserBaseURL      string
	cryptoSecret       string
	jsonWebToken       *jwt.JSONWebToken
	session            session.Session
	cache              redis.UniversalClient
	publisher          pubsub.Publisher
	customerRepository CustomerRepository
}

// ChangeEmail implements CustomerUseCase.
func (u *customerUseCase) ChangeEmail(ctx context.Context, req ChangeEmailRequest) (ChangeEmailResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	acc, err := session.GetAccountFromCtx(ctx)
	if err != nil {
		return ChangeEmailResponse{}, err
	}

	c, err := u.customerRepository.FindByID(ctx, acc.ID, nil)
	if err != nil {
		return ChangeEmailResponse{}, err
	}

	if req.Email == c.Email {
		return ChangeEmailResponse{}, errors.New(http.StatusBadRequest, status.BAD_REQUEST, "the new email is the same as existing email")
	}

	now := time.Now()
	linkExpiresIn := time.Minute * 5
	linkExpiresAt := now.Add(linkExpiresIn)
	verificationToken := util.GenerateRandomHEX(32)
	verificationKey := fmt.Sprintf(changeEmailVerificationKeyPrefix, verificationToken)
	verificationLink := fmt.Sprintf("%s%s?token=%s", u.tmuserBaseURL, ChangeEmailVerificationURLPath, verificationToken)
	changeEmailEvent := ChangeEmailEvent{
		ID:              c.ID,
		Name:            c.Name,
		ExistingEmail:   c.Email,
		NewEmail:        req.Email,
		VerficationLink: verificationLink,
	}

	changeEmailEventBuff, _ := json.Marshal(changeEmailEvent)

	if err := u.cache.Set(ctx, verificationKey, changeEmailEventBuff, linkExpiresIn).Err(); err != nil {
		u.logger.WithContext(ctx).WithError(err).Error()
		return ChangeEmailResponse{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occured while changing customer's email")
	}

	messageHeader := pubsub.MessageHeaders{
		"origin": u.appName,
	}
	u.publisher.Publish(ctx, "customer-change-email", fmt.Sprintf("customer:%d", c.ID), messageHeader, changeEmailEventBuff)

	if err := u.session.Delete(ctx, fmt.Sprintf("customer:%d", c.ID)); err != nil {
		return ChangeEmailResponse{}, err
	}

	resp := ChangeEmailResponse{
		VerificationExpiresAt: linkExpiresAt,
	}

	return resp, nil
}

// ChangePassword implements CustomerUseCase.
func (u *customerUseCase) ChangePassword(ctx context.Context, req ChangePasswordRequest) error {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	acc, err := session.GetAccountFromCtx(ctx)
	if err != nil {
		return err
	}

	c, err := u.customerRepository.FindByID(ctx, acc.ID, nil)
	if err != nil {
		return err
	}

	hashedExistingPassword := util.GenerateSecret(fmt.Sprintf("%s%s", u.cryptoSecret, req.ExistingPassword), c.PasswordSalt, 256)
	if hashedExistingPassword != c.Password {
		return errors.New(http.StatusBadRequest, status.BAD_REQUEST, "invalid customer's existing password")
	}

	newPasswordSalt := util.GenerateRandomHEX(16)
	newHashedPassword := util.GenerateSecret(fmt.Sprintf("%s%s", u.cryptoSecret, req.NewPassword), newPasswordSalt, 256)

	c.Password = newHashedPassword
	c.PasswordSalt = newPasswordSalt

	if err := u.customerRepository.Update(ctx, c.ID, c, nil); err != nil {
		return err
	}

	if err := u.session.Delete(ctx, fmt.Sprintf("customer:%d", c.ID)); err != nil {
		return err
	}

	return nil
}

// GetProfile implements CustomerUseCase.
func (u *customerUseCase) GetProfile(ctx context.Context) (GetProfileResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	acc, err := session.GetAccountFromCtx(ctx)
	if err != nil {
		return GetProfileResponse{}, err
	}

	c, err := u.customerRepository.FindByID(ctx, acc.ID, nil)
	if err != nil {
		return GetProfileResponse{}, err
	}

	resp := GetProfileResponse{
		ID:                 c.ID,
		Name:               c.Name,
		Email:              c.Email,
		VerificationStatus: c.VerificationStatus,
		MemberStatus:       c.MemberStatus,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
	}

	return resp, nil
}

// SignIn implements CustomerUseCase.
func (u *customerUseCase) SignIn(ctx context.Context, req SignInRequest) (SignInResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	c, err := u.customerRepository.FindByEmail(ctx, req.Email, nil)
	if err != nil {
		return SignInResponse{}, err
	}

	if c.VerificationStatus == VerificationStatusUnverified {
		return SignInResponse{}, errors.New(http.StatusForbidden, status.FORBIDDEN, "customer is not verified")
	}

	hashedPassword := util.GenerateSecret(fmt.Sprintf("%s%s", u.cryptoSecret, req.Password), c.PasswordSalt, 256)
	if c.Password != hashedPassword {
		return SignInResponse{}, errors.New(http.StatusBadRequest, status.BAD_REQUEST, "invalid customer's email or password")
	}

	now := time.Now()
	expiresIn := time.Hour * 1
	expiresAt := now.Add(expiresIn)
	subject := fmt.Sprintf("customer:%d", c.ID)
	userType := "CUSTOMER"

	claim := jwt.Claim{}
	claim.Subject = subject
	claim.IssuedAt = now.Unix()
	claim.ExpiresAt = expiresAt.Unix()
	claim.Name = c.Name
	claim.Email = c.Email
	claim.Type = userType
	claim.Issuer = "ticket-master"

	idToken, err := u.jsonWebToken.Sign(ctx, claim)
	if err != nil {
		u.logger.WithContext(ctx).WithError(err).Error()
		return SignInResponse{}, err
	}

	if err := u.session.Set(ctx, fmt.Sprintf("%s:%d", "customer", c.ID), session.Account{
		ID:   c.ID,
		Name: c.Name,
		Type: userType,
	}, expiresIn); err != nil {
		return SignInResponse{}, err
	}

	resp := SignInResponse{
		Token:     idToken,
		ExpiresAt: expiresAt,
	}

	return resp, nil
}

// SignOut implements CustomerUseCase.
func (u *customerUseCase) SignOut(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	acc, err := session.GetAccountFromCtx(ctx)
	if err != nil {
		return err
	}

	_, err = u.customerRepository.FindByID(ctx, acc.ID, nil)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("customer:%d", acc.ID)
	if err := u.session.Delete(ctx, key); err != nil {
		return err
	}

	return nil
}

// SignUp implements CustomerUseCase.
func (u *customerUseCase) SignUp(ctx context.Context, req SignUpRequest) (SignUpResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	_, err := u.customerRepository.FindByEmail(ctx, req.Email, nil)
	if err == nil {
		return SignUpResponse{}, errors.New(http.StatusConflict, status.ALREADY_EXIST, fmt.Sprintf("customer with email '%s' is already registered", req.Email))
	}

	if !errors.MatchStatus(err, status.NOT_FOUND) {
		return SignUpResponse{}, err
	}

	now := time.Now()
	passwordSalt := util.GenerateRandomHEX(32)
	hashedPassword := util.GenerateSecret(fmt.Sprintf("%s%s", u.cryptoSecret, req.Password), passwordSalt, 256)
	c := Customer{
		Name:               req.Name,
		Email:              req.Email,
		Password:           hashedPassword,
		PasswordSalt:       passwordSalt,
		VerificationStatus: VerificationStatusUnverified,
		MemberStatus:       MemberStatusActive,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	ID, err := u.customerRepository.Save(ctx, c, nil)
	if err != nil {
		return SignUpResponse{}, err
	}

	c.ID = ID

	linkExpiresIn := time.Minute * 5
	linkExpiresAt := now.Add(linkExpiresIn)
	verificationToken := util.GenerateRandomHEX(32)
	verificationKey := fmt.Sprintf(verificationKeyPrefix, verificationToken)
	verificationLink := fmt.Sprintf("%s%s?token=%s", u.tmuserBaseURL, VerificationURLPath, verificationToken)
	signUpEvent := SignUpEvent{
		ID:                 c.ID,
		Name:               c.Name,
		Email:              c.Email,
		VerificationStatus: c.VerificationStatus,
		MemberStatus:       c.MemberStatus,
		CreatedAt:          c.CreatedAt,
		VerificationLink:   verificationLink,
	}

	signUpEventBuff, _ := json.Marshal(signUpEvent)

	if err := u.cache.Set(ctx, verificationKey, signUpEventBuff, linkExpiresIn).Err(); err != nil {
		u.logger.WithContext(ctx).WithError(err).Error()
		return SignUpResponse{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occured while signing up customer")
	}

	messageHeader := pubsub.MessageHeaders{
		"origin": u.appName,
	}
	u.publisher.Publish(ctx, "customer-sign-up", fmt.Sprintf("customer:%d", c.ID), messageHeader, signUpEventBuff)

	resp := SignUpResponse{
		VerificationExpiresAt: linkExpiresAt,
	}

	return resp, nil

}

// UpdateProfile implements CustomerUseCase.
func (u *customerUseCase) UpdateProfile(ctx context.Context, req UpdateProfileRequest) error {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	acc, err := session.GetAccountFromCtx(ctx)
	if err != nil {
		return err
	}

	c, err := u.customerRepository.FindByID(ctx, acc.ID, nil)
	if err != nil {
		return err
	}

	c.Name = req.Name

	if err := u.customerRepository.Update(ctx, c.ID, c, nil); err != nil {
		return err
	}

	return nil
}

// Verify implements CustomerUseCase.
func (u *customerUseCase) Verify(ctx context.Context, req VerifyRequest) error {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	key := fmt.Sprintf(verificationKeyPrefix, req.Token)
	signUpEventBuff, err := u.cache.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return errors.New(http.StatusForbidden, status.FORBIDDEN, "invalid verification token")
		}
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occured whil verifying user after sign up")
	}

	var signUpEvent SignUpEvent
	json.Unmarshal(signUpEventBuff, &signUpEvent)

	c, err := u.customerRepository.FindByID(ctx, signUpEvent.ID, nil)
	if err != nil {
		if errors.MatchStatus(err, status.NOT_FOUND) {
			return errors.New(http.StatusForbidden, status.FORBIDDEN, "token is not match any customer data")
		}
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occured while verifying user after sign up")
	}

	now := time.Now()
	c.VerificationStatus = VerficationStatusVerified
	c.UpdatedAt = now

	if err := u.customerRepository.Update(ctx, c.ID, c, nil); err != nil {
		return err
	}

	if err := u.cache.Del(ctx, key).Err(); err != nil {
		u.logger.WithContext(ctx).WithError(err).Error()
	}

	return nil
}

// VerifyChangeEmail implements CustomerUseCase.
func (u *customerUseCase) VerifyChangeEmail(ctx context.Context, req ChangeEmailVerificationRequest) error {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	key := fmt.Sprintf(changeEmailVerificationKeyPrefix, req.Token)
	changeEmailEventBuff, err := u.cache.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return errors.New(http.StatusForbidden, status.FORBIDDEN, "invalid change email verification token")
		}
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occured while verifying user after changing email")
	}

	var changeEmailEvent ChangeEmailEvent
	json.Unmarshal(changeEmailEventBuff, &changeEmailEvent)

	c, err := u.customerRepository.FindByID(ctx, changeEmailEvent.ID, nil)
	if err != nil {
		if errors.MatchStatus(err, status.NOT_FOUND) {
			return errors.New(http.StatusForbidden, status.FORBIDDEN, "token is not match any customer data")
		}
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occured whil verifying user after sign up")
	}

	now := time.Now()
	c.Email = changeEmailEvent.NewEmail
	c.VerificationStatus = VerficationStatusVerified
	c.UpdatedAt = now

	if err := u.customerRepository.Update(ctx, c.ID, c, nil); err != nil {
		return err
	}

	if err := u.cache.Del(ctx, key).Err(); err != nil {
		u.logger.WithContext(ctx).WithError(err).Error()
	}

	return nil
}

func NewCustomerUseCase(props CustomerUseCaseProperty) CustomerUseCase {
	return &customerUseCase{
		appName:            props.AppName,
		logger:             props.Logger,
		timeout:            props.Timeout,
		tmuserBaseURL:      props.TMUserBaseURL,
		cryptoSecret:       props.CryptoSecret,
		jsonWebToken:       props.JSONWebToken,
		session:            props.Session,
		cache:              props.Cache,
		publisher:          props.Publisher,
		customerRepository: props.CustomerRepository,
	}
}
