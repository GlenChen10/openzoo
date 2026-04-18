import { getApiClient } from "../api/connect-client";
import type { RuntimeDevice } from "../types";

export async function listRuntimes(workspaceId: string): Promise<RuntimeDevice[]> {
  const res = await getApiClient().call<{ runtimes: RuntimeDevice[] }>("/rpc/runtime/list", { workspace_id: workspaceId });
  return res.runtimes ?? [];
}

export async function registerRuntime(data: {
  workspace_id: string;
  name: string;
  provider: string;
  runtime_mode?: string;
  device_info?: string;
}): Promise<RuntimeDevice> {
  return getApiClient().call<RuntimeDevice>("/rpc/runtime/register", data);
}

export async function getRuntime(workspaceId: string, runtimeId: string): Promise<RuntimeDevice> {
  return getApiClient().call<RuntimeDevice>("/rpc/runtime/get", { workspace_id: workspaceId, runtime_id: runtimeId });
}

export async function pingRuntime(workspaceId: string, runtimeId: string): Promise<unknown> {
  return getApiClient().call("/rpc/runtime/ping", { workspace_id: workspaceId, runtime_id: runtimeId });
}

export async function listRuntimeUsage(workspaceId: string, runtimeId: string, days = 30): Promise<unknown[]> {
  const res = await getApiClient().call<{ usage: unknown[] }>("/rpc/runtime/usage", { workspace_id: workspaceId, runtime_id: runtimeId, days });
  return res.usage ?? [];
}
