workflow "build docker images" {
  on = "push"
  resolves = ["GitHub Action for Docker"]
}

action "GitHub Action for Docker" {
  uses = "actions/docker/cli@fe7ed3ce992160973df86480b83a2f8ed581cd50"
  args = ["build", "-t", "fuzzitdev/fuzzit:stretch-llvm8", "./docker/debian/stretch/llvm8", "-f", "./docker/debian/stretch/llvm8/Dockerfile"]
}
