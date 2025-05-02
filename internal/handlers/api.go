package handlers

import (
	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
)

func Handler(r *chi.Mux) {
	r.Use(chimiddle.StripSlashes)

	r.Route("/api", func(router chi.Router) {
		router.Get("/check-dusted", CheckAccountDusted)
		router.Get("/check-address-poisoning", CheckAddressPoisoning)
		router.Get("/filter-transactions", FilterTransactions)
	})
}