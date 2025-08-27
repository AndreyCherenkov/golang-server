package api

import "net/http"

// Router — обёртка над http.ServeMux с поддержкой middleware.
type Router struct {
	mux        *http.ServeMux
	middleware []func(http.Handler) http.Handler
}

// NewRouter создаёт новый экземпляр Router.
func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// ServeHTTP реализует интерфейс http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// RegisterRoute регистрирует путь и обработчик.
// Все middleware применяются автоматически при вызове Handler().
func (r *Router) RegisterRoute(path string, handler http.Handler) {
	r.mux.Handle(path, handler)
}

// Use добавляет middleware в стек.
// Middleware выполняются в порядке добавления (первый добавленный — первый вызываемый).
func (r *Router) Use(mw func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, mw)
}

// Handler возвращает http.Handler с применёнными middleware.
// Middleware оборачиваются в обратном порядке (последний добавленный — первый вызываемый).
func (r *Router) Handler() http.Handler {
	var handler http.Handler = r.mux
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}
	return handler
}
