import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { createClient, AuthError } from "./api.ts";

function getAuthUrl(): string {
  if (import.meta.env.VITE_AUTH_URL) {
    return import.meta.env.VITE_AUTH_URL;
  }
  const parts = window.location.hostname.split(".");
  if (parts.length < 3) {
    throw new Error(
      `Cannot derive auth URL from hostname "${window.location.hostname}". Set VITE_AUTH_URL.`,
    );
  }
  const baseDomain = parts.slice(1).join(".");
  return `https://auth.${baseDomain}`;
}

createClient()
  .then(({ client, actor }) => {
    createRoot(document.getElementById("root")!).render(
      <StrictMode>
        <App client={client} actor={actor} />
      </StrictMode>,
    );
  })
  .catch((err) => {
    if (err instanceof AuthError) {
      window.location.replace(getAuthUrl());
    } else {
      document.getElementById("root")!.textContent =
        "Failed to connect to the server. Please try again later.";
    }
  });
