import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { createClient } from "./api.ts";

function getShadowCookie(): string | undefined {
  const match = document.cookie.match(/(?:^|;\s*)shadow=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : undefined;
}

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

const shadowToken = getShadowCookie();
if (!shadowToken) {
  window.location.replace(getAuthUrl());
} else {
  createClient(shadowToken)
    .then(({ client, actor }) => {
      createRoot(document.getElementById("root")!).render(
        <StrictMode>
          <App client={client} actor={actor} />
        </StrictMode>,
      );
    })
    .catch(() => {
      window.location.replace(getAuthUrl());
    });
}
