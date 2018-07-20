# matt
Multi-Account Terraform Tool

matt is a terraform wrapper for working with multiple account, currently works with AWS.

You pass is a account list, authentication source, var path, backendpath, path to terraform files.


syntax

matt [options] [tf module path]

options:
-a (account source), default current IAM user
    l = list of aws accounts
    p = named profile

