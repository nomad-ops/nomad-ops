# Getting Started

## Installation

Nomad Ops can be installed inside or outside of your nomad-cluster.

> If you install Nomad Ops inside your cluster you can leverage the [Workload Identities](https://developer.hashicorp.com/nomad/docs/concepts/workload-identity) of Nomad v1.5.x.

Keep in mind that Nomad Ops stores its data on the local file system. Please use a NFS-mount or similar to enable scheduling around the cluster.

For a simple deployment, consider restricting the deployment to a specific node.

### Configuration 

> TODO

## Security

Deploy keys are saved in plain text. Please make sure that the application is only accessible by authorized personnel. This includes setting up TLS, users and a hardened runtime-environment.

## Workflow

Nomad Ops pulls the `desired state` from a git-repository on a regular basis. Additionally, certain events trigger a re-evaluation of the state as well.

After the `desired state` has been fetched, the `current state` is queried from the `nomad`-cluster. The `reconciler` performs the necessary steps to bring the `cluster state` closer to the `desired state` by adding, updating or deleting jobs.

## User management

Users a currently managed by the [admin interface of pocketbase](https://pocketbase.io/docs/)

## Restrictions

Nomad Ops does **not** perform any templating or rendering and expects the manifests in the repository to be `ready-to-run`. Adjust your CI/CD pipeline to include the rendering step before you commit the file in the repository. 

> Do not store secrets in plain text in your repository. Consult the nomad docs on best practices to provide secrets to your jobs.