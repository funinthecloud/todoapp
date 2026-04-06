import { ProtosourceClient, NoAuth } from "@protosource/client";
import { TodoListHTTPClient } from "./gen/showcase/app/todolist/v1/todolist_v1.protosource.client.js";

const baseURL = import.meta.env.VITE_API_URL || "http://localhost:8080";

const client = new ProtosourceClient(baseURL, new NoAuth("demo-user"), {
  useJSON: true,
});

export const todoListClient = new TodoListHTTPClient(client);
