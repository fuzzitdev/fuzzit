workflow "build docker images" {
  on = "push"
  resolves = ["GitHub Action for Docker-1"]
}

action "GitHub Action for Docker" {
  uses = "actions/docker/cli@master"
  args = ["build", "-t", "docker.pkg.github.com/fuzzitdev/fuzzit/stretch-llvm8:latest", "./docker/debian/stretch/llvm8", "-f", "./docker/debian/stretch/llvm8/Dockerfile"]
}

action "Docker Registry" {
  uses = "actions/docker/cli@master"
  args = ["login", "-u", "yevgenypats", "-p", "$GITHUB_ACTIONS_TOKEN"]
  needs = ["GitHub Action for Docker"]
}

action "GitHub Action for Docker-1" {
  uses = "actions/docker/cli@fe7ed3ce992160973df86480b83a2f8ed581cd50"
  args = ["docker", "push", "docker.pkg.github.com/fuzzitdev/fuzzit/stretch-llvm8:latest"]
  needs = ["Docker Registry"]
  secrets = ["GITHUB_ACTIONS_TOKEN"]
}
