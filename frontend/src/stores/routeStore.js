import { writable } from "svelte/store";

export const page = writable({
  url: new URL(window.location.href),
  params: {},
  route: {
    id: window.location.pathname
  }
});

export function updatePageStore() {
  page.update(() => ({
    url: new URL(window.location.href),
    params: getRouteParams(),
    route: {
      id: window.location.pathname
    }
  }));
}

function getRouteParams() {
  const path = window.location.pathname;
  const params = {};

  const userProfileMatch = path.match(/^\/user\/([^\/]+)$/);
  if (userProfileMatch) {
    params.userId = userProfileMatch[1];
  }

  return params;
}

updatePageStore();

if (typeof window !== "undefined") {
  window.addEventListener("popstate", updatePageStore);

  const originalPushState = history.pushState;
  history.pushState = function(state, title, url) {
    originalPushState.call(this, state, title, url);
    updatePageStore();
  };

  const originalReplaceState = history.replaceState;
  history.replaceState = function(state, title, url) {
    originalReplaceState.call(this, state, title, url);
    updatePageStore();
  };
}