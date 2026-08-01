export async function registerServiceWorker(): Promise<ServiceWorkerRegistration | null> {
  if (!("serviceWorker" in navigator)) {
    return null
  }

  const baseUrl = import.meta.env.BASE_URL
  const serviceWorkerUrl = `${baseUrl}service-worker.js`

  try {
    return await navigator.serviceWorker.register(serviceWorkerUrl, {
      scope: baseUrl,
    })
  } catch (error) {
    console.error('Failed to register service worker', error)
    return null
  }
}
