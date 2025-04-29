package handlers

import (
	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
	"github.com/joelvarghese6/mitigate-dust-attacks/internal/middleware"
)

func Handler(r *chi.Mux) {
	r.Use(chimiddle.StripSlashes)

	r.Route("/account", func(router chi.Router) {
		//middleware for /account route
		router.Use(middleware.Authorization)

		//
		router.Get("/coin", GetCoinBalance)
	})

	r.Route("/api", func(router chi.Router) {
		router.Get("/check-dusted", CheckAccountDusted)
		router.Get("/check-address-poisoning", CheckAddressPoisoning)
		// router.Get("/full-scan", FullScan)
	})
}