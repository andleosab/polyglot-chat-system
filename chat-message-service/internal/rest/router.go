package rest

import (
	"chat-message-service/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"chat-message-service/internal/rest/interceptor"
)

type HandlerRegister interface {
	RegisterRoutes(r chi.Router)
}

type RootRouter struct {
	path   string
	router chi.Router
}

func NewRootRouter(path string, config *config.JWTConfig) *RootRouter {
	return &RootRouter{path: path, router: newRouter(config)}
}

func newRouter(config *config.JWTConfig) chi.Router {

	r := chi.NewRouter()

	// r.Use(cors.Handler(cors.Options{
	// 	AllowedOrigins:   []string{"https://*", "http://*"},
	// 	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
	// 	AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
	// 	AllowCredentials: true,
	// 	MaxAge:           300,
	// }))

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(interceptor.AuthMiddleware(config))

	return r

}

func (root *RootRouter) BuildRouter(registers ...HandlerRegister) {

	root.router.Route(root.path, func(r chi.Router) {
		for _, reg := range registers {
			reg.RegisterRoutes(r)
		}
	})

}

func (root *RootRouter) GetRouter() chi.Router {
	return root.router
}
