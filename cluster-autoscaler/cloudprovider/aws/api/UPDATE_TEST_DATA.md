Update AWS test data
=====

Install AWS CLI Tools

Update

```bash
cd ${GOPATH}/src/k8s.io/autoscaler

# update on demand data
aws pricing get-products \
  --region=us-east-1 \
  --service-code=AmazonEC2 \
  --filter Type=TERM_MATCH,Field=capacitystatus,Value=Used \
           Type=TERM_MATCH,Field=preInstalledSw,Value=NA \
           Type=TERM_MATCH,Field=location,Value="EU (Ireland)" \
           Type=TERM_MATCH,Field=instanceType,Value="m4.xlarge" \
  > ./cluster-autoscaler/cloudprovider/aws/api/pricing_ondemand_eu-west-1.json

# update spot data

```