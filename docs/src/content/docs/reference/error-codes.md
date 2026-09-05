---
title: Error Codes
description: JSON-RPC error codes used across the OriginLang protocol.
---

Errors are returned as a JSON-RPC error object: `{ "code": ..., "message": ..., "data": ... }`.

## Standard JSON-RPC 2.0 codes

| Code | Name | Meaning |
| --- | --- | --- |
| `-32700` | `ParseError` | Invalid JSON was received by the server |
| `-32600` | `InvalidRequest` | The JSON sent is not a valid Request object |
| `-32601` | `MethodNotFound` | The method does not exist / is not available |
| `-32602` | `InvalidParams` | Invalid method parameter(s) |
| `-32603` | `InternalError` | Internal JSON-RPC error |

## Server error reserved range

Codes from `-32099` to `-32000` are reserved for implementation-defined server errors.

## Application error codes

Application-level codes in the `-32000..-31900` range are host/plugin-specific and stable across transports.

| Code | Name | Meaning |
| --- | --- | --- |
| `-32001` | `PluginNotFound` | No plugin with the given ID is loaded |
| `-32002` | `PluginStopped` | The plugin has no usable endpoint (closed / not connected) |
| `-32003` | `Timeout` | An RPC call timed out |
| `-32004` | `UnknownMethod` | The requested method is not served by the target |
| `-32005` | `OperationFailed` | Generic operation failure on the remote side |

## Future / reserved codes

As the security, tenant, and quota subsystems land, these application codes are planned:

| Code | Name | Meaning |
| --- | --- | --- |
| `-32006` | `Unauthorized` | Permission denied by the security engine |
| `-32007` | `MethodForbidden` | Method not declared in the plugin's capabilities |
| `-32008` | `TenantIsolated` | Cross-tenant access denied |
| `-32009` | `QuotaExceeded` | Resource quota exceeded |
| `-32010` | `CircuitBroken` | Circuit breaker is open |
| `-32011` | `DependencyMissing` | A plugin dependency is unsatisfied |

## Error object shape

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32002,
    "message": "endpoint closed",
    "data": { ... optional, transport-specific ... }
  }
}
```

Error objects double as Go errors: a failed call surfaces the code and message through the normal error path, and the payload can carry structured `data`.

## Notes

- A request to an unregistered method yields `-32601 MethodNotFound` regardless of transport.
- A call to a plugin whose endpoint is closed yields `-32002 PluginStopped` so hosts can distinguish "plugin vanished" from "plugin answered with an error".