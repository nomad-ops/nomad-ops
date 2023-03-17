plugin "docker" {
    config {
        volumes {
            enabled      = true
            selinuxlabel = "z"
        }
    }
}