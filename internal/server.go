/*
Copyright © 2020 Jody Scott <jody@thescottsweb.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package internal

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jodydadescott/libtokenmachine"
	"github.com/jodydadescott/libtokenmachine/serverlib"
	"go.uber.org/zap"
)

// Config ...
type Config struct {
	Policy                                                              string
	NonceLifetime, SecretMaxLifetime, SecretMinLifetime, KeytabLifetime int64
	SecretSecrets                                                       []*libtokenmachine.Secret
	KeytabKeytabs                                                       []*libtokenmachine.Keytab
	Listen, TLSCert, TLSKey                                             string
	HTTPPort, HTTPSPort                                                 int
}

// Server ...
type Server struct {
	closed                  chan struct{}
	wg                      sync.WaitGroup
	httpServer, httpsServer *http.Server
	libTokenMachine         libtokenmachine.LibTokenMachine
}

// Build Returns a new Server
func (config *Config) Build() (*Server, error) {

	zap.L().Info(fmt.Sprintf("Starting"))

	if config.HTTPPort < 0 {
		return nil, fmt.Errorf("HTTPPort must be 0 or greater")
	}

	if config.HTTPSPort < 0 {
		return nil, fmt.Errorf("HTTPSPort must be 0 or greater")
	}

	if config.HTTPPort == 0 && config.HTTPSPort == 0 {
		return nil, fmt.Errorf("Must enable http or https")
	}

	if config.Policy == "" {
		return nil, fmt.Errorf("Policy is required")
	}

	libTokenMachineConfig := &libtokenmachine.Config{
		Policy:         config.Policy,
		NonceLifetime:  config.NonceLifetime,
		SecretSecrets:  config.SecretSecrets,
		KeytabKeytabs:  config.KeytabKeytabs,
		KeytabLifetime: config.KeytabLifetime,
	}

	libTokenMachine, err := serverlib.NewInstance(libTokenMachineConfig)
	if err != nil {
		return nil, err
	}

	server := &Server{
		closed:          make(chan struct{}),
		libTokenMachine: libTokenMachine,
	}

	if config.HTTPPort > 0 {
		listen := config.Listen
		if strings.ToLower(listen) == "any" {
			listen = ""
		}
		listener := listen + ":" + strconv.Itoa(config.HTTPPort)
		zap.L().Debug("Starting HTTP")
		server.httpServer = &http.Server{Addr: listener, Handler: server}
		go func() {
			server.httpServer.ListenAndServe()
		}()
	}

	if config.HTTPSPort > 0 {
		listen := config.Listen
		if strings.ToLower(listen) == "any" {
			listen = ""
		}
		listener := listen + ":" + strconv.Itoa(config.HTTPSPort)

		zap.L().Debug("Starting HTTPS")

		if config.TLSCert == "" {
			return nil, fmt.Errorf("TLSCert is required when HTTPS port is set")
		}

		if config.TLSKey == "" {
			return nil, fmt.Errorf("TLSKey is required when HTTPS port is set")
		}

		cert, err := tls.X509KeyPair([]byte(config.TLSCert), []byte(config.TLSKey))
		if err != nil {
			return nil, err
		}

		server.httpsServer = &http.Server{Addr: listener, Handler: server, TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}}}

		go func() {
			server.httpsServer.ListenAndServeTLS("", "")
		}()

	}

	go func() {

		for {
			select {
			case <-server.closed:
				zap.L().Debug("Shutting down")

				if server.httpServer != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					server.httpServer.Shutdown(ctx)
				}

				if server.httpsServer != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					server.httpsServer.Shutdown(ctx)
				}

			}
		}
	}()

	return server, nil
}

// ServeHTTP HTTP/HTTPS Handler
func (t *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	zap.L().Debug(fmt.Sprintf("Entering ServeHTTP path=%s method=%s", r.URL.Path, r.Method))

	defer zap.L().Debug(fmt.Sprintf("Exiting ServeHTTP path=%s method=%s", r.URL.Path, r.Method))

	w.Header().Set("Content-Type", "application/json")

	token := getBearerToken(r)

	if token == "" {
		http.Error(w, newErrorResponse("Token required")+"\n", http.StatusConflict)
		return
	}

	switch r.URL.Path {
	case "/getnonce":
		nonce, err := t.libTokenMachine.GetNonce(r.Context(), token)
		if handleERR(w, err) {
			return
		}
		fmt.Fprintf(w, nonce.JSON()+"\n")
		return

	case "/getkeytab":

		principal := getKey(r, "principal")
		if principal == "" {
			http.Error(w, newErrorResponse("Principal required")+"\n", http.StatusConflict)
			return
		}

		keytab, err := t.libTokenMachine.GetKeytab(r.Context(), token, principal)
		if handleERR(w, err) {
			return
		}

		fmt.Fprintf(w, keytab.JSON()+"\n")
		return

	case "/getsecret":
		name := getKey(r, "name")
		if name == "" {
			http.Error(w, newErrorResponse("Parameter 'name' required")+"\n", http.StatusConflict)
			return
		}

		result, err := t.libTokenMachine.GetSecret(r.Context(), token, name)
		if handleERR(w, err) {
			return
		}
		fmt.Fprintf(w, result.JSON()+"\n")
		return
	}

	http.Error(w, newErrorResponse("Path "+r.URL.Path+" not mapped")+"\n", http.StatusConflict)

	zap.L().Debug(fmt.Sprintf("Exiting ServeHTTP"))
}

func newErrorResponse(message string) string {
	return "{\"error\":\"" + message + "\"}"
}

func handleERR(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	http.Error(w, newErrorResponse(err.Error())+"\n", http.StatusConflict)
	return true
}

func getBearerToken(r *http.Request) string {
	// If the Bearer token is present it may be in Authorization Header in the format 'Authorization: Bearer TOKEN' or as a parameter
	token := r.Header.Get("Authorization")
	if token != "" {
		tokenSlice := strings.Split(token, " ")
		if len(tokenSlice) > 1 {
			if strings.ToLower(tokenSlice[0]) == "bearer" {
				return tokenSlice[1]
			}
		}
	}
	return getKey(r, "bearertoken")
}

func getKey(r *http.Request, name string) string {
	keys, ok := r.URL.Query()[name]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return string(keys[0])
}

// Shutdown Server
func (t *Server) Shutdown() {
	zap.L().Info(fmt.Sprintf("Stopping"))
	t.libTokenMachine.Shutdown()
	close(t.closed)
	t.wg.Wait()
}

// StatusBadRequest                   = 400 // RFC 7231, 6.5.1
// StatusUnauthorized                 = 401 // RFC 7235, 3.1
// StatusPaymentRequired              = 402 // RFC 7231, 6.5.2
// StatusForbidden                    = 403 // RFC 7231, 6.5.3
// StatusNotFound                     = 404 // RFC 7231, 6.5.4
// StatusMethodNotAllowed             = 405 // RFC 7231, 6.5.5
// StatusNotAcceptable                = 406 // RFC 7231, 6.5.6
// StatusProxyAuthRequired            = 407 // RFC 7235, 3.2
// StatusRequestTimeout               = 408 // RFC 7231, 6.5.7
// StatusConflict                     = 409 // RFC 7231, 6.5.8
// StatusGone                         = 410 // RFC 7231, 6.5.9
// StatusLengthRequired               = 411 // RFC 7231, 6.5.10
// StatusPreconditionFailed           = 412 // RFC 7232, 4.2
// StatusRequestEntityTooLarge        = 413 // RFC 7231, 6.5.11
// StatusRequestURITooLong            = 414 // RFC 7231, 6.5.12
// StatusUnsupportedMediaType         = 415 // RFC 7231, 6.5.13
// StatusRequestedRangeNotSatisfiable = 416 // RFC 7233, 4.4
// StatusExpectationFailed            = 417 // RFC 7231, 6.5.14
// StatusTeapot                       = 418 // RFC 7168, 2.3.3
// StatusMisdirectedRequest           = 421 // RFC 7540, 9.1.2
// StatusUnprocessableEntity          = 422 // RFC 4918, 11.2
// StatusLocked                       = 423 // RFC 4918, 11.3
