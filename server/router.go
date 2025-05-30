package server

import (
	"net/http"

	"github.com/ssongin/tartarus/api"
)

type TartarusRouter struct{}

func (app *TartarusRouter) ApiRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/thanatos/", http.StripPrefix("/thanatos", api.GetArchiveRouter()))
	return mux
}

func (app *TartarusRouter) UiRouter() *http.ServeMux {
	mux := http.NewServeMux()
	return mux
}

func (app *TartarusRouter) Route() *http.ServeMux {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.Handle("/api/", http.StripPrefix("/api", app.ApiRouter()))
	mux.Handle("/", app.UiRouter())

	return mux
}
