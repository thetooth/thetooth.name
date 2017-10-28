job "thetooth.name" {
  datacenters = ["dc1"]
  type = "service"

  update {
    stagger = "10s"
    max_parallel = 1
  }

  group "thetooth.name-service" {
    count = 1
    ephemeral_disk {
      size = 32
    }

    restart {
      attempts = 10
      interval = "5m"
      delay = "25s"
      mode = "delay"
    }

    task "thetooth.name-server" {
      driver = "docker"
      config {
        image = "thetooth/thetooth.name:latest"
        port_map { http = 9000 }
      }

      service { # consul service checks
        name = "thetooth.name"
        tags = ["http"]
        port = "http"
        check {
          type     = "http"
          interval = "10s"
          timeout  = "2s"
          path = "/"
        }
      }

      resources {
        cpu    = 512 # MHz 
        memory = 256 # MB 
        network {
          mbits = 10
          port "http" {}
        }
      }

      logs {
        max_files     = 3
        max_file_size = 2
      }
    }
  }
}
