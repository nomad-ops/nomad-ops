# Getting Started

## Getting started with docker

> You need [nomad](https://developer.hashicorp.com/nomad/docs/install) and docker installed on your system

1. Clone the repo
2. Start 2 terminal sessions at the root of this repo

Run the following in the first terminal:

> Make sure that docker volumes are available  
> You can use the provided `.deployment/nomad/agent.hcl` as a reference  
> Remember to set the `NOMAD_ADDR` environment variable to the address of your nomad cluster

`nomad agent -dev -bind 0.0.0.0 -log-level INFO -config .deployment/nomad/agent.hcl`

This will bring up a nomad environment with docker volumes enabled. See [nomad docs](https://developer.hashicorp.com/nomad/docs/operations/nomad-agent) for more info.

Run the following in the second terminal:

`nomad namespace apply nomad-ops`

This makes sure that the namespace nomad-ops exists.

Deploy Nomad-Ops to nomad by running:

`nomad job run .deployment/nomad/docker.hcl`

### Access the Admin UI

Go to [http://localhost:8080/_/](http://localhost:8080/_/).
This will bring you to the login screen of [pocketbase](https://pocketbase.io).
Login using `admin@nomad-ops.org` and `simple-nomad-ops`.

> You only need to access this UI to create additional users. The main UI is available at [http://localhost:8080/](http://localhost:8080/)

### Access the Nomad-Ops UI

You can access the UI of Nomad-Ops at [http://localhost:8080/](http://localhost:8080/) and use your newly created credentials to login.

By default the following user is created:

- email: `user@nomad-ops.org`
- password: `simple-nomad-ops`

> You can change the default user by setting the environment variables `DEFAULT_USER_EMAIL` and `DEFAULT_USER_PASSWORD`.

> The Admin User `admin@nomad-ops.org` is **only** capable to access the pocketbase ui at `http://{your-nomad-ops-host}/_/`. To access the Nomad-Ops UI you need to use a *normal* user.

## Installation Notes

Nomad Ops can be installed inside or outside of your nomad-cluster.

> If you install Nomad Ops inside your cluster you can leverage the [Workload Identities](https://developer.hashicorp.com/nomad/docs/concepts/workload-identity) of Nomad v1.5.x.

Keep in mind that Nomad Ops stores its data on the local file system. Please use a NFS-mount or similar to enable scheduling around the cluster.

> There are performace implications if using a network mount.    
> Prefer to use a local disk if possible.  

For a simple deployment, consider restricting the deployment to a specific node.

### Configuration 

| ENVIRONMENT Variable   | Default                   | Description                                                                    |
| ---------------------- | ------------------------- | ------------------------------------------------------------------------------ |
| DEFAULT_ADMIN_EMAIL    | admin@nomad-ops.org       | On first startup an admin user is created with this email                      |
| DEFAULT_ADMIN_PASSWORD | simple-nomad-ops          | On first startup an admin user is created with this password                   |
| NOMAD_ADDR             | ''                        | Nomad addr                                                                     |
| NOMAD_TOKEN            | ''                        | Nomad token to access the Nomad API                                            |
| NOMAD_TOKEN_FILE       | ''                        | If set will ignore NOMAD_TOKEN and read from this file instead                 |
| TRACE                  | FALSE                     | If set to `TRUE` enables detailed logging                                      |
| SLACK_WEBHOOK_URL      | ''                        | Set to your Webhook URL if you want to receive notifications about deployments |
| SLACK_BASE_URL         | 'localhost:3000/ui/'      | included in the slack message as a link                                        |
| SLACK_ICON_SUCCESS     | ':check:'                 | Icon to use for successful deployments                                         |
| SLACK_ICON_ERROR       | ':check-no:'              | Icon to use for unsuccessful deployments                                       |
| SLACK_ENV_INFO_TEXT    | 'Sent by nomad-ops (dev)' | Send as a footer in the slack message                                          |

There are a couple of [Pocketbase](https://pocketbase.io) settings that you can set as well. See [here](https://github.com/nomad-ops/nomad-ops/blob/main/backend/cmd/nomad-ops-server/main.go#L65).

#### Email Settings

[Pocketbase](https://pocketbase.io) integrates a couple of workflows for user management (confirmation, password reset, ...). To use that please adjust the environment variables according to the [docs](https://pocketbase.io/docs/api-settings/). See [here](https://github.com/nomad-ops/nomad-ops/blob/main/backend/cmd/nomad-ops-server/main.go#L65) for the corresponding environment variables in Nomad-Ops.

## Security

Deploy keys are saved in plain text. Please make sure that the application is only accessible by authorized personnel. This includes setting up TLS, users and a hardened runtime-environment.

## Workflow

Nomad Ops pulls the `desired state` from a git-repository on a regular basis. Additionally, certain events trigger a re-evaluation of the state as well.

After the `desired state` has been fetched, the `current state` is queried from the `nomad`-cluster. The `reconciler` performs the necessary steps to bring the `cluster state` closer to the `desired state` by adding, updating or deleting jobs.

## User management

Users are currently managed by the [admin interface of pocketbase](https://pocketbase.io/docs/)

## Restrictions

Nomad Ops does **not** perform any templating or rendering and expects the manifests in the repository to be `ready-to-run`. Adjust your CI/CD pipeline to include the rendering step before you commit the file in the repository. 

> Do not store secrets in plain text in your repository. Consult the nomad docs on best practices to provide secrets to your jobs.
