import { ProtosourceClient } from "@protosource/client";
import type { AuthProvider } from "@protosource/client";
import { TodoListHTTPClient } from "./gen/showcase/app/todolist/v1/todolist_v1.protosource.client.js";

const baseURL = import.meta.env.VITE_API_URL || "http://localhost:8080";

class HeaderAuth implements AuthProvider {
  private readonly _actor: string;

  constructor(actor: string) {
    this._actor = actor;
  }

  authenticate(headers: Headers): void {
    headers.set("X-Actor", this._actor);
  }

  actor(): string {
    return this._actor;
  }
}

const client = new ProtosourceClient(baseURL, new HeaderAuth("demo-user"), {
  useJSON: true,
});

export const todoListClient = new TodoListHTTPClient(client);
