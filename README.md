# Klustered

This go application was designed to prevent the other team in a Klustered challenge from being able to update a deployment from a v1 to a v2 image. There are several approaches it take to doing this.

## Usage

### Deployment

Just run `klustered deploy` and the pod will be created along with a service account and clusterrolebinding for that service account.

## Breaks

### MutatingAdmissionWebhook

The application registers a mutating admission webhook that modifies any pod with the klustered v2 image back to the klustered v1 image.

### ValidatingAdmissionWebhook

The application registers a validating admission webhook that reject any create or update of a deployment using the v2 image

### Self-Healing

If either of the hooks, or the service are deleted, there is a background loop that will re-create those resources.

Additionally, if the pod receives a shutdown signal (INT, KILL, etc.), it will create a new copy of itself.

### Hiding in the open

All of the resources created are named to look innocent. Things like `default`, `api-server`, `cilium-XXXXX`, etc.
