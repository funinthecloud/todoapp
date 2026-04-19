import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { createClient } from "./api.ts";

function getShadowCookie(): string | undefined {
  const match = document.cookie.match(/(?:^|;\s*)shadow=([^;]*)/);
  return match ? match[1] : undefined;
}

function getAuthUrl(): string {
  const parts = window.location.hostname.split(".");
  const baseDomain = parts.slice(1).join(".");
  return `https://auth.${baseDomain}`;
}

const shadowToken = getShadowCookie();
if (!shadowToken) {
  window.location.href = getAuthUrl();
} else {
  createClient(shadowToken).then(({ client, actor }) => {
    createRoot(document.getElementById("root")!).render(
      <StrictMode>
        <App client={client} actor={actor} />
      </StrictMode>,
    );
  });
}
