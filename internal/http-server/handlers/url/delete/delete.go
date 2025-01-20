package delete

import (
	"log/slog"
	"net/http"

	"URL-shortener/internal/lib/api/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		alias := chi.URLParam(request, "alias")
		if alias == "" {
			log.Info("invalid request")

			response.ErrorStatus(writer, request, "invalid request", http.StatusBadRequest)

			return
		}

		err := deleter.DeleteURL(alias)
		if err != nil {
			log.Info("failed to delete url")

			response.ErrorStatus(writer, request, "failed to delete url", http.StatusInternalServerError)

			return
		}

		log.Info("url deleted successfully", slog.String("alias", alias))
		response.SuccessStatus(writer, request, "url deleted successfully")
	}
}
