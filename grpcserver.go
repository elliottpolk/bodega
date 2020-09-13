package bodega

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/elliottpolk/bodega/config"
	"github.com/elliottpolk/bodega/product"
	"github.com/elliottpolk/bodega/purchase"
	"github.com/elliottpolk/bodega/shop"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/grpc"
)

func ServeGRPC(ctx context.Context, comp *config.Composition) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", comp.Server.RpcPort))
	if err != nil {
		return errors.Wrap(err, "unable to create tcp listener")
	}

	server := grpc.NewServer()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mdbc, err := mongo.Connect(ctx, options.Client().ApplyURI(comp.Db.ConnString()))
	if err != nil {
		return errors.Wrap(err, "unable to generate mongo client")
	}
	defer mdbc.Disconnect(ctx)

	if err := mdbc.Ping(ctx, readpref.Primary()); err != nil {
		return errors.Wrap(err, "unable to verify connection to mongo")
	}

	// register services
	product.RegisterServiceServer(server, product.NewServer(comp, mdbc))
	shop.RegisterServiceServer(server, shop.NewServer(comp, mdbc))
	purchase.RegisterServiceServer(server, purchase.NewServer(comp, mdbc))

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		// receiving an interrupt signal, similar to a 'Ctrl+C'
		for range c {
			log.Println("shutting down gRPC server...")
			server.GracefulStop()

			<-ctx.Done()
		}
	}()

	log.Println("starting gRPC server...")
	return server.Serve(listener)
}
