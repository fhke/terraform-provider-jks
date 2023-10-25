resource "random_password" "keystore" {
  length = 16
}

data "jks_keystore" "this" {
  password = random_password.keystore.result

  key_pair {
    alias       = "cert"
    certificate = var.server_cert
    private_key = var.private_key

    intermediate_certificates = [
      var.intermediate_cert,
    ]
  }
}