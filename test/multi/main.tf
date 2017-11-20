resource "local_file" "foo" {
  content  = "foo!"
  filename = "${path.module}/tmp/foo.txt"
}

resource "tls_private_key" "example" {
  algorithm = "ECDSA"
}
