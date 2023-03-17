job "beta-test-nomad-ops" {
  namespace = "beta-test"
  # Specify this job should run in the region named "us". Regions
  # are defined by the Nomad servers' configuration.
  #region = "us"

  # Spread the tasks in this job between us-west-1 and us-east-1.
  datacenters = ["dus2"]

  # Run this job as a "service" type. Each job type has different
  # properties. See the documentation below for more examples.
  type = "service"

  # Specify this job to have rolling updates, two-at-a-time, with
  # 30 second intervals.
  update {
    stagger      = "30s"
    max_parallel = 1
  }

  # A group defines a series of tasks that should be co-located
  # on the same client (host). All tasks within a group will be
  # placed on the same host.
  group "nomad-ops-group" {
    # Specify the number of these tasks we want.
    count = 1
    
    network {
      port "http" {}
    }

    # The service block tells Nomad how to register this service
    # with Consul for service discovery and monitoring.
    service {
      name = "nomad-ops"
      tags = ["http","view"]

      # This tells Consul to monitor the service on the port
      # labelled "http". Since Nomad allocates high dynamic port
      # numbers, we use labels to refer to them.
      port = "http"

      check {
        type     = "http"
        path     = "/nomad-ops/api/v1/live"
        interval = "10s"
        timeout  = "2s"
      }
    }

    # Create an individual task (unit of work). This particular
    # task utilizes a Docker container to front a web application.
    task "nomad-ops" {
      # Specify the driver to be "docker". Nomad supports
      # multiple drivers.
      driver = "docker"

      env {
        # local database
        # http://10.1.3.155/api/interpreter
        # public database
        # https://overpass.kumi.systems/api/interpreter
        LEVELDB_PATH = "/tmp/nomad-ops.db"
        CLUSTER_ADDR = "http://nomad-ui.prod.eu.tcs.trv.cloud"
        CLUSTER_TOKEN = "TODO"
        HOST = "0.0.0.0"
        PORT = "${NOMAD_PORT_http}"
      }

      # Configuration is specific to each driver.
      config {
        image = "myimg/nomad-ops:latest"

        ports = [
            "http",
        ]
      }

      # Specify the maximum resources required to run the task,
      # include CPU, memory, and bandwidth.
      resources {
        cpu    = 500 # MHz
        memory = 256 # MB
      }
    }
  }
}