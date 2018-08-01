# matt
Multi-Account Terraform Tool

matt is a terraform wrapper for working with multiple account, currently works with AWS.

matt.ini defines the static vars to pass as var flags.



````
{
    "AWSTemplateFormatVersion": "2010-09-09",
    "Description": "Creates an IAM role with cross-account access for Terraform.",
    "Metadata": {
        "VersionDate": {
            "Value": "20170717"
        },
        "Identifier": {
            "Value": "terraform-service"
        }
    },
    "Resources": {
        "CrossAccountRole": {
            "Type": "AWS::IAM::Role",
            "Properties": {
                "RoleName": "terraform-service",
                "AssumeRolePolicyDocument": {
                    "Version": "2012-10-17",
                    "Statement": [
                        {
                            "Effect": "Allow",
                            "Principal": {
                                "AWS": [
                                    "arn:aws:iam::CHANGEME:root"
                                ]
                            },
                            "Action": "sts:AssumeRole"
                        }
                    ]
                },
                "Path": "/",
                "Policies": [
                    {
                        "PolicyName": "AdministratorAccess",
                        "PolicyDocument": {
                            "Version": "2012-10-17",
                            "Statement": [
                                {
                                    "Effect": "Allow",
                                    "Action": "*",
                                    "Resource": "*"
                                }
                            ]
                        }
                    }
                ]
            }
        }
    }
}
````