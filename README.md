# IF POSSIBLE, USE BLACKBOX_EXPORTER
Please use https://github.com/prometheus/blackbox_exporter instead.

# Blackbox Prober

Export blackbox telemetry like availability, request latencies and
request size for remote services.

## Supported URLs
### http/https
The exporter requests the given url and reads from it until EOF.

### tcp
The exporter connects to the given host:port. If any path is given, it
will try to read until EOF which is required for exposing the size.

### icmp
Execute `ping`. Port and path are ignored.

## Example
See blackbox_example.yml for configuration
