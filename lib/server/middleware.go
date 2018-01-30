package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/centrifugal/centrifugo/lib/logger"
)

func (s *HTTPServer) apiAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		s.RLock()
		apiKey := s.config.APIKey
		apiInsecure := s.config.APIInsecure
		s.RUnlock()
		if apiKey == "" {
			logger.ERROR.Println("no API key found in configuration")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !apiInsecure {
			parts := strings.Fields(authorization)
			if len(parts) != 2 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			authMethod := strings.ToLower(parts[0])
			if authMethod != "apikey" || parts[1] != apiKey {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

// checkAdminAuthToken checks admin connection token which Centrifugo returns after admin login.
// func (s *HTTPServer) checkAdminAuthToken(token string) error {

// 	s.RLock()
// 	adminPassword := s.node.Config().AdminPassword
// 	adminSecret := s.node.Config().AdminSecret
// 	adminInsecure := s.node.Config().AdminInsecure
// 	s.RUnlock()

// 	if adminInsecure {
// 		return nil
// 	}

// 	if adminSecret == "" {
// 		logger.ERROR.Println("no admin secret set in configuration")
// 		return proto.ErrUnauthorized
// 	}

// 	if token == "" {
// 		return proto.ErrUnauthorized
// 	}

// 	auth := auth.CheckAdminToken(adminSecret, token)
// 	if !auth {
// 		return proto.ErrUnauthorized
// 	}
// 	return nil
// }

// apiAuth ...
// func (s *HTTPServer) adminAPIAuth(h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		authorization := r.Header.Get("Authorization")

// 		parts := strings.Fields(authorization)
// 		if len(parts) != 2 {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}
// 		authMethod := strings.ToLower(parts[0])
// 		if authMethod != "token" || s.checkAdminAuthToken(parts[1]) != nil {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		h.ServeHTTP(w, r)
// 	})
// }

// wrapShutdown will return http Handler.
// If Application in shutdown it will return http.StatusServiceUnavailable.
func (s *HTTPServer) wrapShutdown(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.RLock()
		shutdown := s.shutdown
		s.RUnlock()
		if shutdown {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// log middleware logs request.
func (s *HTTPServer) log(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var start time.Time
		if logger.DEBUG.Enabled() {
			start = time.Now()
		}
		h.ServeHTTP(w, r)
		if logger.DEBUG.Enabled() {
			addr := r.Header.Get("X-Real-IP")
			if addr == "" {
				addr = r.Header.Get("X-Forwarded-For")
				if addr == "" {
					addr = r.RemoteAddr
				}
			}
			logger.DEBUG.Printf("%s %s from %s completed in %s\n", r.Method, r.URL.Path, addr, time.Since(start))
		}
		return
	})
}