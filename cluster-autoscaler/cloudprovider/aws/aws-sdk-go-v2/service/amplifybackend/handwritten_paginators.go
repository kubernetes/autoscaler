package amplifybackend

import (
	"context"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
)

// ListBackendJobsPaginatorOptions is the paginator options for ListBackendJobs
type ListBackendJobsPaginatorOptions struct {
	// (Optional) The maximum number of shards to return in a single call
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListBackendJobsPaginator is a paginator for ListBackendJobs
type ListBackendJobsPaginator struct {
	options     ListBackendJobsPaginatorOptions
	client      ListBackendJobsAPIClient
	params      *ListBackendJobsInput
	firstPage   bool
	nextToken   *string
	isTruncated bool
}

// ListBackendJobsAPIClient is a client that implements the ListBackendJobs operation.
type ListBackendJobsAPIClient interface {
	ListBackendJobs(context.Context, *ListBackendJobsInput, ...func(*Options)) (*ListBackendJobsOutput, error)
}

// NewListBackendJobsPaginator returns a new ListBackendJobsPaginator
func NewListBackendJobsPaginator(client ListBackendJobsAPIClient, params *ListBackendJobsInput, optFns ...func(options *ListBackendJobsPaginatorOptions)) *ListBackendJobsPaginator {
	if params == nil {
		params = &ListBackendJobsInput{}
	}

	options := ListBackendJobsPaginatorOptions{}
	options.Limit = aws.ToInt32(params.MaxResults)

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListBackendJobsPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListBackendJobsPaginator) HasMorePages() bool {
	return p.firstPage || p.isTruncated
}

// NextPage retrieves the next ListBackendJobs page.
func (p *ListBackendJobsPaginator) NextPage(ctx context.Context, optFns ...func(*Options)) (*ListBackendJobsOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.NextToken = p.nextToken

	var limit int32
	if p.options.Limit > 0 {
		limit = p.options.Limit
	}
	params.MaxResults = aws.Int32(limit)

	result, err := p.client.ListBackendJobs(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.isTruncated = result.NextToken != nil
	p.nextToken = nil
	if result.NextToken != nil {
		p.nextToken = result.NextToken
	}

	if p.options.StopOnDuplicateToken &&
		prevToken != nil &&
		p.nextToken != nil &&
		*prevToken == *p.nextToken {
		p.isTruncated = false
	}

	return result, nil
}
