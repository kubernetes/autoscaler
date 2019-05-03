The following is an excerpt from a CloudFormation template showing how a MixedInstancesPolicy can be used with ClusterAutoscaler:

```json
{
    "Resources": {
        "LaunchTemplate": {
            "Type": "AWS::EC2::LaunchTemplate",
            "Properties": {
                "LaunchTemplateName": "memory-opt-2xlarge",
                "LaunchTemplateData": {
                    "InstanceType": "r5.2xlarge"
                }
            }
        },
        "ASGA": {
            "Type": "AWS::AutoScaling::AutoScalingGroup",
            "Properties": {
                "MinSize": 1,
                "MaxSize": 10,
                "MixedInstancesPolicy": {
                    "InstancesDistribution": {
                        "OnDemandBaseCapacity": 0,
                        "OnDemandPercentageAboveBaseCapacity": 0
                    },
                    "LaunchTemplate": {
                        "LaunchTemplateSpecification": {
                            "LaunchTemplateId": {
                                "Ref": "LaunchTemplate"
                            },
                            "Version": {
                                "Fn::GetAtt": [
                                    "LaunchTemplate",
                                    "LatestVersionNumber"
                                ]
                            }
                        },
                        "Overrides": [
                            {
                                "InstanceType": "r5.2xlarge"
                            },
                            {
                                "InstanceType": "r5d.2xlarge"
                            },
                            {
                                "InstanceType": "i3.2xlarge"
                            },
                            {
                                "InstanceType": "r5a.2xlarge"
                            },
                            {
                                "InstanceType": "r5ad.2xlarge"
                            }
                        ]
                    }
                },
                "VPCZoneIdentifier": [
                    "subnet-###############"
                ],
            }
        },
        "ASGB": {},
        "ASGC": {}
    }
}
```

[r5.2xlarge](https://aws.amazon.com/ec2/instance-types/#Memory_Optimized) is the 'base' instance type, with overrides for r5d.2xlarge, i3.2xlarge, r5a.2xlarge and r5ad.2xlarge. 

Note how one Auto Scaling Group is created per Availability Zone, since CA does not currently support ASGs that span multiple Availability Zones. See [Common Notes and Gotchas](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/cloudprovider/aws#common-notes-and-gotchas).
