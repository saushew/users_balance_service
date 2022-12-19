package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/saushew/users-balance-service/cmd/model"
	"github.com/saushew/users-balance-service/cmd/store"
	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/saushew/users-balance-service/docs" // import swagger docs
)

const (
	ctxKeyRequestID ctxKey = iota
)

type ctxKey int8

type server struct {
	router *mux.Router
	logger *logrus.Logger
	store  store.Store
}

func newServer(store store.Store) *server {
	s := &server{
		router: mux.NewRouter(),
		logger: logrus.New(),
		store:  store,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)

	s.router.HandleFunc("/balance", s.handleGetBalance()).Methods("GET")
	s.router.HandleFunc("/transactions", s.handleGetTransactions()).Methods("GET")

	s.router.HandleFunc("/deposit", s.handleDeposit()).Methods("POST")
	s.router.HandleFunc("/withdraw", s.handleWithdraw()).Methods("POST")
	s.router.HandleFunc("/transfer", s.handleTransfer()).Methods("POST")

	s.router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8000/swagger/doc.json"), //The url pointing to API definition
	)).Methods(http.MethodGet)
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestID),
		})
		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		var level logrus.Level
		switch {
		case rw.code >= 500:
			level = logrus.ErrorLevel
		case rw.code >= 400:
			level = logrus.WarnLevel
		default:
			level = logrus.InfoLevel
		}
		logger.Logf(
			level,
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Now().Sub(start),
		)
	})
}

//	@Summery		Balance
//	@Tags			User data
//	@Description	Returns the account balance by user id
//	@ID				balance
//	@Accept			json
//	@Produce		json
//	@Param			user_id	query		int	true	"User ID"
//	@Success		200		{object}	model.User
//	@Router			/balance [get]
func (s *server) handleGetBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := r.URL.Query().Get("user_id")

		id, err := strconv.Atoi(req)
		if err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, errIncorrectID)
			return
		}

		u, err := s.store.User().Find(id)
		if err != nil {
			s.error(w, r, http.StatusNotFound, err)
			return
		}

		s.respond(w, r, http.StatusOK, u)
	}
}

//	@Summery		Transactions
//	@Tags			User data
//	@Description	Provides paginated list of user transactions
//	@ID				transactions
//	@Accept			json
//	@Produce		json
//	@Param			user_id	query		int		true	"User ID"
//	@Param			page	query		int		false	"Page"	default(1)
//	@Param			limit	query		int		false	"Limit"	default(10)
//	@Param			order	query		string	false	"Order"	Enums(desc, asc)
//	@Success		200		{object}	[]model.Transaction
//	@Router			/transactions [get]
func (s *server) handleGetTransactions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_id := r.URL.Query().Get("user_id")
		_page := r.URL.Query().Get("page")
		_limit := r.URL.Query().Get("limit")
		order := r.URL.Query().Get("order")

		if _page == "" {
			_page = "1"
		}
		if _limit == "" {
			_limit = "10"
		}
		if order == "" {
			order = "desc"
		}

		order = strings.ToLower(order)
		if order != "" && order != "desc" && order != "asc" {
			s.error(w, r, http.StatusUnprocessableEntity, errIncorrectOrder)
			return
		}

		id, err := strconv.Atoi(_id)
		if err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, errIncorrectID)
			return
		}

		page, err := strconv.Atoi(_page)
		if err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, errIncorrectPage)
			return
		}

		limit, err := strconv.Atoi(_limit)
		if err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, errIncorrectLimit)
			return
		}

		offset := (page - 1) * limit

		u, err := s.store.Transaction().Get(id, offset, limit, order)
		if err != nil {
			s.error(w, r, http.StatusNotFound, err)
			return
		}

		s.respond(w, r, http.StatusOK, u)
	}
}

type depositRequest struct {
	ID     int     `json:"user_id" example:"1"`
	Amount float64 `json:"amount" example:"99.90"`
}

//	@Summery		Deposit
//	@Tags			Funds
//	@Description	Adding funds to the user's balance
//	@ID				deposit
//	@Accept			json
//	@Produce		json
//	@Param			input	body		depositRequest	true	"Deposit info"
//	@Success		200		{object}	model.Transaction
//	@Router			/deposit [post]
func (s *server) handleDeposit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &depositRequest{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		id := req.ID
		amount := req.Amount
		if amount < 0 {
			s.error(w, r, http.StatusUnprocessableEntity, errIncorrectAmount)
			return
		}

		t := &model.Transaction{
			UserID:    id,
			Amount:    amount,
			Type:      model.TxDeposit,
			Details:   "receiving external funds",
			Timestamp: time.Now().Unix(),
		}

		if err := s.store.Transaction().Create(t); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusOK, t)
	}
}

type withdrawRequest struct {
	ID     int     `json:"user_id" example:"1"`
	Amount float64 `json:"amount" example:"99.90"`
}

//	@Summery		Withdrawal
//	@Tags			Funds
//	@Description	Withdrawal of funds from the user's balance
//	@ID				withdraw
//	@Accept			json
//	@Produce		json
//	@Param			input	body		withdrawRequest	true	"Withdrawal info"
//	@Success		200		{object}	model.Transaction
//	@Router			/withdraw [post]
func (s *server) handleWithdraw() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &withdrawRequest{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		id := req.ID
		amount := req.Amount
		if amount < 0 {
			s.error(w, r, http.StatusUnprocessableEntity, errIncorrectAmount)
			return
		}

		t := &model.Transaction{
			UserID:    id,
			Amount:    amount,
			Type:      model.TxWithdraw,
			Details:   "withdrawal of funds",
			Timestamp: time.Now().Unix(),
		}
		if err := s.store.Transaction().Create(t); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusOK, t)
	}
}

type transferRequest struct {
	From   int     `json:"from" example:"1"`
	To     int     `json:"to" example:"2"`
	Amount float64 `json:"amount" example:"99.90"`
}

//	@Summery		Transfer
//	@Tags			Funds
//	@Description	Transfers funds from one user to another
//	@ID				transfer
//	@Accept			json
//	@Produce		json
//	@Param			input	body		transferRequest	true	"Transfer info"
//	@Success		200		{object}	[]model.Transaction
//	@Router			/transfer [post]
func (s *server) handleTransfer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &transferRequest{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		from := req.From
		to := req.To
		amount := req.Amount
		if amount < 0 {
			s.error(w, r, http.StatusUnprocessableEntity, errIncorrectAmount)
			return
		}

		ts := time.Now().Unix()
		t1 := &model.Transaction{
			UserID:    from,
			Amount:    amount,
			Type:      model.TxWithdraw,
			Details:   fmt.Sprintf("transfer to user_id = %d", to),
			Timestamp: ts,
		}
		t2 := &model.Transaction{
			UserID:    to,
			Amount:    amount,
			Type:      model.TxDeposit,
			Details:   fmt.Sprintf("transfer from user_id = %d", from),
			Timestamp: ts,
		}

		if err := s.store.Transaction().Transfer(t1, t2); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusOK, []*model.Transaction{t1, t2})
	}
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
