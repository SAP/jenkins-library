# Generated Fortify SSC client

Internalized copy of the go-swagger generated Fortify SSC REST API client from
[github.com/piper-validation/fortify-client-go](https://github.com/piper-validation/fortify-client-go)
at commit `7b3e9a72af01` (v0.0.0-20220126145513-7b3e9a72af01).

It is reduced to what the `fortifyExecuteScan` step actually uses:

- `fortify/`: the 14 controller packages consumed by `pkg/fortify`, plus a trimmed `fortify_client.go`
- `models/`: the transitive closure of models referenced by those controllers and `pkg/fortify`

Do not edit by hand (except for further pruning). To update, regenerate the client with
[go-swagger](https://goswagger.io/) from the Fortify SSC swagger spec and re-apply the pruning.
