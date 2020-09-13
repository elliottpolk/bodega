package bodega

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/elliottpolk/bodega/config"
	"github.com/elliottpolk/bodega/product"
	"github.com/elliottpolk/bodega/purchase"
	"github.com/elliottpolk/bodega/shop"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func ServeREST(ctx context.Context, comp *config.Composition) error {
	var (
		mux  = runtime.NewServeMux()
		opts = []grpc.DialOption{grpc.WithInsecure()}
	)

	_ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// register services
	if err := product.RegisterServiceHandlerFromEndpoint(_ctx, mux, fmt.Sprintf(":%s", comp.Server.RpcPort), opts); err != nil {
		return errors.Wrap(err, "unable to register product service handler")
	}

	if err := purchase.RegisterServiceHandlerFromEndpoint(_ctx, mux, fmt.Sprintf(":%s", comp.Server.RpcPort), opts); err != nil {
		return errors.Wrap(err, "unable to register product service handler")
	}

	if err := shop.RegisterServiceHandlerFromEndpoint(_ctx, mux, fmt.Sprintf(":%s", comp.Server.RpcPort), opts); err != nil {
		return errors.Wrap(err, "unable to register product service handler")
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", comp.Server.HttpPort),
		Handler: mux,
	}

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// TODO:
			// sig is a ^C, handle it
		}

		_, cancel := context.WithTimeout(_ctx, 5*time.Second)
		defer cancel()

		log.Infoln("shutting down HTTP/RESTful gateway...")
		if err := server.Shutdown(_ctx); err != nil {
			log.Error(err)
		}
	}()

	//  start HTTPS listener in a seperate go routine since it is a blocking func
	go func() {
		cert, key := comp.Server.TlsCert, comp.Server.TlsKey
		if len(cert) < 1 || len(key) < 1 {
			return // skip if no cert nor key
		}

		if _, err := os.Stat(cert); err != nil {
			log.Error(errors.Wrap(err, "unable to access TLS cert file"))
			return
		}

		if _, err := os.Stat(key); err != nil {
			log.Error(errors.Wrap(err, "unable to access TLS key file"))
			return
		}

		server := &http.Server{
			Addr:    fmt.Sprintf(":%s", comp.Server.HttpsPort),
			Handler: mux,
		}

		// graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for range c {
				// TODO:
				// sig is a ^C, handle it
			}

			_, cancel := context.WithTimeout(_ctx, 5*time.Second)
			defer cancel()

			log.Infoln("shutting down HTTPS/RESTful gateway...")
			if err := server.Shutdown(_ctx); err != nil {
				log.Error(err)
			}
		}()

		log.Infoln("starting HTTPSRESTful gateway...")
		log.Fatal(server.ListenAndServeTLS(cert, key))
	}()

	log.Infoln("starting HTTP/RESTful gateway...")
	return server.ListenAndServe()
}
