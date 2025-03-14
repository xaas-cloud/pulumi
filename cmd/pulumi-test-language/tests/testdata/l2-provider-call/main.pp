resource "prov" "pulumi:providers:call" {
    value = "bar"
}

output "from_identity" {
    value = call(prov, "identity", {}).result
}

output "from_prefixed" {
    value = call(prov, "prefixed", { prefix = "foo-" }).result
}
