import { ProtosourceClient, BearerTokenAuth } from "@protosource/client";
import { TodoListHTTPClient } from "./gen/showcase/app/todolist/v1/todolist_v1.protosource.client.js";

const baseURL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export class AuthError extends Error {
  constructor(status: number) {
    super(`/whoami returned ${status}`);
    this.name = "AuthError";
  }
}

export async function createClient(shadowToken: string) {
  const resp = await fetch(`${baseURL}/whoami`, {
    headers: { Authorization: `Bearer ${shadowToken}` },
  });
  if (resp.status === 401 || resp.status === 403) {
    throw new AuthError(resp.status);
  }
  if (!resp.ok) {
    throw new Error(`/whoami failed: ${resp.status}`);
  }
  const { actor } = await resp.json();

  const client = new ProtosourceClient(
    baseURL,
    new BearerTokenAuth(shadowToken, actor),
    { useJSON: true },
  );
  return { client: new TodoListHTTPClient(client), actor };
}
