import { ProtosourceClient, NoAuth } from "@protosource/client";
import { TodoListHTTPClient } from "./gen/showcase/app/todolist/v1/todolist_v1.protosource.client.js";

const baseURL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export class AuthError extends Error {
  constructor(status: number) {
    super(`/whoami returned ${status}`);
    this.name = "AuthError";
  }
}

function fetchWithCredentials(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<Response> {
  return fetch(input, { ...init, credentials: "include" });
}

export async function createClient() {
  const resp = await fetchWithCredentials(`${baseURL}/whoami`);
  if (resp.status === 401 || resp.status === 403) {
    throw new AuthError(resp.status);
  }
  if (!resp.ok) {
    throw new Error(`/whoami failed: ${resp.status}`);
  }
  const { actor } = await resp.json();

  const client = new ProtosourceClient(baseURL, new NoAuth(actor), {
    useJSON: true,
    fetch: fetchWithCredentials,
  });
  return { client: new TodoListHTTPClient(client), actor };
}
