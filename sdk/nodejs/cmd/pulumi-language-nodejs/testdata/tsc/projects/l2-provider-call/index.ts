import * as pulumi from "@pulumi/pulumi";
import * as call from "@pulumi/call";

const prov = new call.Provider("prov", {value: "bar"});
export const from_identity = prov.identity().apply(call => call.result);
export const from_prefixed = prov.prefixed(({
    prefix: "foo-",
})).apply(call => call.result);
