import { ProtosourceClient, BearerTokenAuth } from "@protosource/client";
import { TodoListHTTPClient } from "./gen/showcase/app/todolist/v1/todolist_v1.protosource.client.js";

const baseURL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export async function createClient(shadowToken: string) {
  const resp = await fetch(`${baseURL}/whoami`, {
    headers: { Authorization: `Bearer ${shadowToken}` },
  });
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
