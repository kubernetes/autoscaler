package healthcheck

import (
	"context"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
)

// Service provides the API operation methods for making requests to
// endpoints of the Spotinst API. See this package's package overview docs
// for details on the service.
type Service interface {
	List(context.Context, *ListHealthChecksInput) (*ListHealthChecksOutput, error)
	Create(context.Context, *CreateHealthCheckInput) (*CreateHealthCheckOutput, error)
	Read(context.Context, *ReadHealthCheckInput) (*ReadHealthCheckOutput, error)
	Update(context.Context, *UpdateHealthCheckInput) (*UpdateHealthCheckOutput, error)
	Delete(context.Context, *DeleteHealthCheckInput) (*DeleteHealthCheckOutput, error)
}

type ServiceOp struct {
	Client *client.Client
}

var _ Service = &ServiceOp{}

func New(sess *session.Session, cfgs ...*spotinst.Config) *ServiceOp {
	cfg := &spotinst.Config{}
	cfg.Merge(sess.Config)
	cfg.Merge(cfgs...)

	return &ServiceOp{
		Client: client.New(sess.Config),
	}
}
