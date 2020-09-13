package shop

import (
	context "context"

	"github.com/elliottpolk/bodega/config"
	"github.com/elliottpolk/bodega/record"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	cmp    *config.Composition
	client *mongo.Client
}

func NewServer(cmp *config.Composition, client *mongo.Client) ServiceServer {
	return &Server{
		cmp:    cmp,
		client: client,
	}
}

func (s *Server) Create(ctx context.Context, req *Request) (*Empty, error) {
	empty := &Empty{RequestId: req.RequestId}

	if s.client == nil {
		return empty, record.ErrNoMongoClient
	}

	client := s.client
	if err := client.UseSession(ctx, func(session mongo.SessionContext) error {
		defer session.EndSession(ctx)

		if err := Create(session, req.Payload, client.Database(s.cmp.Db.Name)); err != nil {
			defer session.AbortTransaction(ctx)
			return err
		}

		return nil
	}); err != nil {
		return empty, err
	}
	return empty, nil
}

func (s *Server) Retrieve(ctx context.Context, req *Request) (*Response, error) {
	if s.client == nil {
		return nil, record.ErrNoMongoClient
	}

	result := &Response{
		RequestId: req.RequestId,
	}

	client := s.client
	if err := client.UseSession(ctx, func(session mongo.SessionContext) error {
		defer session.EndSession(ctx)

		// retrieve 1 and return by ID if provided in request
		if id := req.Id; len(id) > 0 {
			item, err := RetrieveOne(ctx, id, client.Database(s.cmp.Db.Name))
			if err != nil {
				return errors.Wrapf(err, "unable to retrieve record for id %s", id)
			}
			result.Payload = []*Shop{item}

			return nil
		}

		items, err := Retrieve(ctx, bson.D{}, client.Database(s.cmp.Db.Name))
		if err != nil {
			return errors.Wrap(err, "unable to retrieve records")
		}
		result.Payload = items

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Server) Update(ctx context.Context, req *Request) (*Response, error) {
	if s.client == nil {
		return nil, record.ErrNoMongoClient
	}

	result := &Response{
		RequestId: req.RequestId,
	}

	client := s.client
	if err := client.UseSession(ctx, func(session mongo.SessionContext) error {
		defer session.EndSession(ctx)

		if err := Update(session, req.Username, bson.D{}, req.Payload, client.Database(s.cmp.Db.Name)); err != nil {
			defer session.AbortTransaction(ctx)
			return errors.Wrap(err, "unable to update records")
		}
		result.Payload = req.Payload

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Server) Delete(ctx context.Context, req *Request) (*Empty, error) {
	empty := &Empty{RequestId: req.RequestId}

	if s.client == nil {
		return empty, record.ErrNoMongoClient
	}

	client := s.client
	if err := client.UseSession(ctx, func(session mongo.SessionContext) error {
		defer session.EndSession(ctx)

		if err := Delete(session, req.Username, req.Payload, client.Database(s.cmp.Db.Name)); err != nil {
			defer session.AbortTransaction(ctx)
			return errors.Wrap(err, "unable to delete records")
		}

		return nil
	}); err != nil {
		return empty, err
	}

	return empty, nil
}
