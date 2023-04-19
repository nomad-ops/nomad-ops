# nomad-ops

A simple operator for nomad which reconciles the running jobs in comparison to git repos

> Still work in progress, but it should do what it is supposed to :)  
> See [http://nomad-ops.org](http://nomad-ops.org) for more information

## Getting started with docker

> You need [nomad](https://developer.hashicorp.com/nomad/docs/install) and docker installed on your system

1. Clone the repo
2. Start 2 terminal sessions at the root of this repo

Run the following in the first terminal:

> Make sure that docker volumes are available  
> You can use the provided `.deployment/nomad/agent.hcl` as a reference

`nomad agent -dev -bind 0.0.0.0 -log-level INFO -config .deployment/nomad/agent.hcl`

This will bring up a nomad environment with docker volumes enabled. See [nomad docs](https://developer.hashicorp.com/nomad/docs/operations/nomad-agent) for more info.

Run the following in the second terminal:

`nomad namespace apply nomad-ops`

This makes sure that the namespace nomad-ops exists.

Deploy Nomad-Ops to nomad by running:

`nomad job run .deployment/nomad/docker.hcl`

Go to [http://localhost:8080/_/](http://localhost:8080/_/).
This will bring you to the login screen of [pocketbase](https://pocketbase.io).
Login using `admin@nomad-ops.org` and `simple-nomad-ops`.

Once you are logged in, you are able to create your first *normal* user. Just hit `+ New Record` for the `User` collection and fill out the form. Afterwards you can access the UI of Nomad-Ops at [http://localhost:8080/](http://localhost:8080/) and use your newly created credentials to login.

> The Admin User `admin@nomad-ops.org` is **only** capable to access the pocketbase ui at `http://{your-nomad-ops-host}/_/`. To access the Nomad-Ops UI you need to create a *normal* user first.

### Workload Identity 

```
nomad acl policy apply \
   -namespace nomad-ops -job nomad-ops -group nomad-ops-group -task operator \
   nomad-ops-policy .deployment/nomad/acl.hcl
```

> Requires nomad >= v1.5.x to use [Workload Identity](https://developer.hashicorp.com/nomad/docs/concepts/workload-identity)  
> Before that you need to supply a NOMAD_TOKEN yourself

## Thanks

> Thanks to https://github.com/pocketbase/pocketbase for providing a solid base : ) !

## Publishing docs

1. `docker run --rm -it -p 8000:8000 -v ${PWD}:/docs -v ~/.ssh:/root/.ssh --entrypoint /bin/sh --platform linux/amd64 squidfunk/mkdocs-material`
2. `apk add -U git openssh`
3. `git config --global url."git@github.com:".insteadOf "https://github.com/"`
4. `mkdocs gh-deploy`