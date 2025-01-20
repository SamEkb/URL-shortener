package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	"URL-shortener/internal/lib/api/response"
	"URL-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLGetter
type URLGetter interface {
	GetURL(aliasToFind string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		alias := chi.URLParam(request, "alias")
		if alias == "" {
			log.Info("alias is empty")

			response.ErrorStatus(writer, request, "invalid request", http.StatusBadRequest)

			return
		}

		url, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)

			response.ErrorStatus(writer, request, "url not found", http.StatusBadRequest)

			return
		}
		if err != nil {
			log.Error("failed to get url", slog.String("error", err.Error()))

			response.ErrorStatus(writer, request, "internal error", http.StatusInternalServerError)

			return
		}

		log.Info("got url", slog.String("url", url))

		http.Redirect(writer, request, url, http.StatusFound)
	}
}
