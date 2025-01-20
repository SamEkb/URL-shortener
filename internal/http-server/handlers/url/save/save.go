package save

import (
	"errors"
	"log/slog"
	"net/http"

	"URL-shortener/internal/lib/api/response"
	"URL-shortener/internal/lib/random"
	"URL-shortener/internal/storage"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

const aliasLength = 6

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		var req Request
		err := render.DecodeJSON(request.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", slog.String("error", err.Error()))

			render.JSON(writer, request, response.Error("failed to decode request body"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			validationErr := err.(validator.ValidationErrors)

			log.Error("failed to validate request", slog.String("error", err.Error()))

			render.JSON(writer, request, response.ValidationError(validationErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Error("url already exists", slog.String("url", req.URL))

			render.JSON(writer, request, response.Error("url already exists"))

			return
		}
		if err != nil {
			log.Error("failed to add url", slog.String("error", err.Error()))

			render.JSON(writer, request, response.Error("failed to add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		responseOk(writer, request, alias)
	}
}

func responseOk(writer http.ResponseWriter, request *http.Request, alias string) {
	render.JSON(writer, request, Response{
		Response: response.Ok(),
		Alias:    alias},
	)
}
