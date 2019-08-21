workflow "build docker images" {
  on = "push"
  resolves = ["GitHub Action for Docker"]
}

action "GitHub Action for Docker" {
  uses = "actions/docker/cli@master"
  args = ["build", "-t", "fuzzitdev/fuzzit:stretch-llvm8", "./docker/debian/stretch/llvm8", "-f", "./docker/debian/stretch/llvm8/Dockerfile"]
}
