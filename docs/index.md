# Nomad Ops

This application brings [Gitops](TOdO) to the [nomad](https://www.nomadproject.io/) world. In short:

> Your manifests of what should be deployed live in one or multiple git-repositories and the application makes sure to synchronize your desired state with the nomad cluster.

![git ops overview](https://codefresh.io/wp-content/uploads/2022/03/Codefresh-GitOps.jpeg)
> Source: https://codefresh.io/wp-content/uploads/2022/03/Codefresh-GitOps.jpeg

## UI

Nomad Ops comes with an integrated (simple) UI to onboard new repositories.

![create sources](./create.png)

![sources](./watching.png)

## Notifications

Nomad Ops is able to notify whenever the `current state` was changed.

The following notification channels are available.

### Slack

> TODO: link config

### Webhook

> Coming soon