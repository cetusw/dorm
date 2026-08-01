const DEFAULT_NOTIFICATION_TITLE = "Дежурства";
const DEFAULT_NOTIFICATION_URL = "/app/";
const DEFAULT_NOTIFICATION_ICON = "/app/icons/app-icon-192.png";
const DEFAULT_NOTIFICATION_BADGE = "/app/icons/app-icon-192.png";

self.addEventListener("install", () => {
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(self.clients.claim());
});

function parsePushPayload(event) {
  if (!event.data) {
    return {};
  }

  try {
    return event.data.json();
  } catch {
    return {
      body: event.data.text(),
    };
  }
}

function normalizeString(value, fallback) {
  return typeof value === "string" && value.trim() !== ""
    ? value.trim()
    : fallback;
}

function normalizeNotificationUrl(value, fallback) {
  if (typeof value !== "string") {
    return fallback;
  }

  try {
    const resolvedUrl = new URL(value, self.location.origin);

    if (resolvedUrl.origin !== self.location.origin) {
      return fallback;
    }

    if (!resolvedUrl.pathname.startsWith("/app/")) {
      return fallback;
    }

    return resolvedUrl.pathname + resolvedUrl.search + resolvedUrl.hash;
  } catch {
    return fallback;
  }
}

self.addEventListener("push", (event) => {
  const payload = parsePushPayload(event);
  const title = normalizeString(payload.title, DEFAULT_NOTIFICATION_TITLE);
  const body = normalizeString(payload.body, "");
  const url = normalizeNotificationUrl(payload.url, DEFAULT_NOTIFICATION_URL);
  const icon = normalizeString(payload.icon, DEFAULT_NOTIFICATION_ICON);
  const badge = normalizeString(payload.badge, DEFAULT_NOTIFICATION_BADGE);
  const tag =
    typeof payload.tag === "string" && payload.tag.trim() !== ""
      ? payload.tag.trim()
      : undefined;

  event.waitUntil(
    self.registration.showNotification(title, {
      body,
      icon,
      badge,
      tag,
      data: {
        url,
      },
    }),
  );
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();

  const targetPath = normalizeNotificationUrl(
    event.notification.data?.url,
    DEFAULT_NOTIFICATION_URL,
  );
  const targetUrl = new URL(targetPath, self.location.origin).href;

  event.waitUntil(
    self.clients
      .matchAll({
        type: "window",
        includeUncontrolled: true,
      })
      .then(async (windowClients) => {
        for (const client of windowClients) {
          const clientUrl = new URL(client.url);

          if (clientUrl.origin !== self.location.origin) {
            continue;
          }

          if ("navigate" in client) {
            await client.navigate(targetUrl);
          }

          if ("focus" in client) {
            return client.focus();
          }
        }

        return self.clients.openWindow(targetUrl);
      }),
  );
});
