# nomad-ops

Let's build a simple operator for nomad which reconciles the running jobs in comparison to git repos

> Still work in progress, but it should do what it is supposed to :)

## Getting started with docker

> Make sure that docker volumes are available

`nomad agent -dev -bind 0.0.0.0 -log-level INFO`

`nomad namespace apply nomad-ops`

Adjust the settings in `.deployment/nomad/docker.hcl`

`nomad job run .deployment/nomad/docker.hcl`



```
nomad acl policy apply \
   -namespace nomad-ops -job nomad-ops -group nomad-ops-group -task operator \
   nomad-ops-policy .deployment/nomad/acl.hcl
```

> Requires nomad >= v1.5.x to use [Workload Identity](https://developer.hashicorp.com/nomad/docs/concepts/workload-identity)  
> Before that you need to supply a NOMAD_TOKEN yourself

> Thanks to https://github.com/pocketbase/pocketbase for providing a solid base : ) !
