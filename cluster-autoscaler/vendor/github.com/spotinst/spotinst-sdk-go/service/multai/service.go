package multai

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
	ListLoadBalancers(context.Context, *ListLoadBalancersInput) (*ListLoadBalancersOutput, error)
	CreateLoadBalancer(context.Context, *CreateLoadBalancerInput) (*CreateLoadBalancerOutput, error)
	ReadLoadBalancer(context.Context, *ReadLoadBalancerInput) (*ReadLoadBalancerOutput, error)
	UpdateLoadBalancer(context.Context, *UpdateLoadBalancerInput) (*UpdateLoadBalancerOutput, error)
	DeleteLoadBalancer(context.Context, *DeleteLoadBalancerInput) (*DeleteLoadBalancerOutput, error)

	ListListeners(context.Context, *ListListenersInput) (*ListListenersOutput, error)
	CreateListener(context.Context, *CreateListenerInput) (*CreateListenerOutput, error)
	ReadListener(context.Context, *ReadListenerInput) (*ReadListenerOutput, error)
	UpdateListener(context.Context, *UpdateListenerInput) (*UpdateListenerOutput, error)
	DeleteListener(context.Context, *DeleteListenerInput) (*DeleteListenerOutput, error)

	ListRoutingRules(context.Context, *ListRoutingRulesInput) (*ListRoutingRulesOutput, error)
	CreateRoutingRule(context.Context, *CreateRoutingRuleInput) (*CreateRoutingRuleOutput, error)
	ReadRoutingRule(context.Context, *ReadRoutingRuleInput) (*ReadRoutingRuleOutput, error)
	UpdateRoutingRule(context.Context, *UpdateRoutingRuleInput) (*UpdateRoutingRuleOutput, error)
	DeleteRoutingRule(context.Context, *DeleteRoutingRuleInput) (*DeleteRoutingRuleOutput, error)

	ListMiddlewares(context.Context, *ListMiddlewaresInput) (*ListMiddlewaresOutput, error)
	CreateMiddleware(context.Context, *CreateMiddlewareInput) (*CreateMiddlewareOutput, error)
	ReadMiddleware(context.Context, *ReadMiddlewareInput) (*ReadMiddlewareOutput, error)
	UpdateMiddleware(context.Context, *UpdateMiddlewareInput) (*UpdateMiddlewareOutput, error)
	DeleteMiddleware(context.Context, *DeleteMiddlewareInput) (*DeleteMiddlewareOutput, error)

	ListTargetSets(context.Context, *ListTargetSetsInput) (*ListTargetSetsOutput, error)
	CreateTargetSet(context.Context, *CreateTargetSetInput) (*CreateTargetSetOutput, error)
	ReadTargetSet(context.Context, *ReadTargetSetInput) (*ReadTargetSetOutput, error)
	UpdateTargetSet(context.Context, *UpdateTargetSetInput) (*UpdateTargetSetOutput, error)
	DeleteTargetSet(context.Context, *DeleteTargetSetInput) (*DeleteTargetSetOutput, error)

	ListTargets(context.Context, *ListTargetsInput) (*ListTargetsOutput, error)
	CreateTarget(context.Context, *CreateTargetInput) (*CreateTargetOutput, error)
	ReadTarget(context.Context, *ReadTargetInput) (*ReadTargetOutput, error)
	UpdateTarget(context.Context, *UpdateTargetInput) (*UpdateTargetOutput, error)
	DeleteTarget(context.Context, *DeleteTargetInput) (*DeleteTargetOutput, error)

	ListDeployments(context.Context, *ListDeploymentsInput) (*ListDeploymentsOutput, error)
	CreateDeployment(context.Context, *CreateDeploymentInput) (*CreateDeploymentOutput, error)
	ReadDeployment(context.Context, *ReadDeploymentInput) (*ReadDeploymentOutput, error)
	UpdateDeployment(context.Context, *UpdateDeploymentInput) (*UpdateDeploymentOutput, error)
	DeleteDeployment(context.Context, *DeleteDeploymentInput) (*DeleteDeploymentOutput, error)

	ListCertificates(context.Context, *ListCertificatesInput) (*ListCertificatesOutput, error)
	CreateCertificate(context.Context, *CreateCertificateInput) (*CreateCertificateOutput, error)
	ReadCertificate(context.Context, *ReadCertificateInput) (*ReadCertificateOutput, error)
	UpdateCertificate(context.Context, *UpdateCertificateInput) (*UpdateCertificateOutput, error)
	DeleteCertificate(context.Context, *DeleteCertificateInput) (*DeleteCertificateOutput, error)

	ListRuntimes(context.Context, *ListRuntimesInput) (*ListRuntimesOutput, error)
	ReadRuntime(context.Context, *ReadRuntimeInput) (*ReadRuntimeOutput, error)
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
